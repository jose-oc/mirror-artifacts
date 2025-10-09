package images

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/appcontext"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
)

// Image represents an entry in images.yaml
type Image struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}

// ImagesList represents the structure of images.yaml
type ImagesList struct {
	Images []Image `yaml:"images"`
}

// MirrorImages mirrors container images to GAR
func MirrorImages(ctx *appcontext.AppContext, imagesFile string) error {
	logger := ctx.Logger
	if imagesFile == "" {
		return fmt.Errorf("images file path is required")
	}

	// Read images.yaml
	data, err := os.ReadFile(imagesFile)
	if err != nil {
		logger.Error().Err(err).Str("file", imagesFile).Msg("Failed to read images file")
		return err
	}

	var imagesList ImagesList
	if err := yaml.Unmarshal(data, &imagesList); err != nil {
		logger.Error().Err(err).Str("file", imagesFile).Msg("Failed to parse images file")
		return err
	}

	// Track failed images for GitHub Actions
	failedImages := make([]map[string]string, 0)

	for _, img := range imagesList.Images {
		logger.Info().Str("name", img.Name).Str("source", img.Source).Msg("Processing image")

		if ctx.DryRun {
			logger.Info().
				Str("command", fmt.Sprintf("oras cp %s %s/%s", img.Source, ctx.Config.GCP.GARRepoContainers, img.Name)).
				Msg("Dry-run: Would mirror image to GAR")
			continue
		}

		// Initialize ORAS source and target registries
		// Equivalent to: oras cp <source> <target>
		sourceRepo, err := remote.NewRepository(img.Source)
		if err != nil {
			logger.Error().Err(err).Str("source", img.Source).Msg("Failed to initialize source repository")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		targetRepoPath := fmt.Sprintf("%s/%s", ctx.Config.GCP.GARRepoContainers, img.Name)
		targetRepo, err := remote.NewRepository(targetRepoPath)
		if err != nil {
			logger.Error().Err(err).Str("target", targetRepoPath).Msg("Failed to initialize target repository")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		// Check if image already exists in GAR (idempotency)
		sourceDesc, err := sourceRepo.Resolve(context.Background(), sourceRepo.Reference.Reference)
		if err != nil {
			logger.Error().Err(err).Str("source", img.Source).Msg("Failed to resolve source image")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		targetDesc, err := targetRepo.Resolve(context.Background(), targetRepo.Reference.Reference)
		if err == nil && targetDesc.Digest == sourceDesc.Digest {
			logger.Info().Str("name", img.Name).Str("digest", sourceDesc.Digest.String()).Msg("Image already exists in GAR, skipping")
			continue
		} else if err == nil && targetDesc.Digest != sourceDesc.Digest && ctx.Config.Options.NotifyTagMutations {
			logger.Warn().
				Str("name", img.Name).
				Str("source_digest", sourceDesc.Digest.String()).
				Str("target_digest", targetDesc.Digest.String()).
				Msg("Tag points to different digest in GAR")
		}

		// Create a temporary local store for ORAS
		// localStore, err := oci.New("/tmp/oci-mirror")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create local OCI store")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}
		defer os.RemoveAll("/tmp/oci-mirror")

		// Mirror the image
		// Equivalent to: oras cp <source> <target>
		_, err = oras.Copy(context.Background(), sourceRepo, sourceRepo.Reference.Reference, targetRepo, targetRepo.Reference.Reference, oras.DefaultCopyOptions)
		if err != nil {
			logger.Error().Err(err).Str("source", img.Source).Str("target", targetRepoPath).Msg("Failed to mirror image")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		logger.Info().Str("name", img.Name).Str("target", targetRepoPath).Msg("Successfully mirrored image to GAR")
	}

	// Log failed images in JSON format for GitHub Actions
	if len(failedImages) > 0 {
		failedJSON, _ := json.Marshal(map[string][]map[string]string{"failed_images": failedImages})
		logger.Error().RawJSON("failed_images", failedJSON).Msg("Some images failed to mirror")
	}

	return nil
}
