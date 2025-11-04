package images

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// MirrorImages mirrors container images to GAR
// It reads the images from the provided YAML file, authenticates with Google Artifact Registry,
// and then mirrors each image.
//
// Parameters:
//
//	ctx: The application context containing configuration and other utilities.
//	imagesFile: The path to the YAML file containing the list of images to mirror.
//
// Returns:
//
//	A map[string]string where keys are source image names and values are their corresponding target repository paths in GAR.
//	A map[string]string where keys are source image names and values are their digests.
//	An error if the mirroring process fails for any reason, otherwise nil.
func MirrorImages(ctx *appcontext.AppContext, imagesFile string) (map[string]string, map[string]string, error) {
	if imagesFile == "" {
		return nil, nil, fmt.Errorf("images file path is required")
	}

	// Read images.yaml
	data, err := os.ReadFile(imagesFile)
	if err != nil {
		log.Error().Err(err).Str("file", imagesFile).Msg("Failed to read images file")
		return nil, nil, err
	}
	var imagesList types.ImagesList
	if err := yaml.Unmarshal(data, &imagesList); err != nil {
		log.Error().Err(err).Str("file", imagesFile).Msg("Failed to parse images file")
		return nil, nil, err
	}

	// Log the image list in a pretty format
	log.Info().Interface("images", imagesList).Str("file", imagesFile).Msg("Loaded images from file")
	return MirrorImageList(ctx, imagesList)
}

// MirrorImageList mirrors container images to GAR
//
// Parameters:
//
//	ctx: The application context containing configuration and other utilities.
//	imagesList: list of images to mirror.
//
// Returns:
//
//	A map[string]string where keys are source image names and values are their corresponding target repository paths in GAR.
//	A map[string]string where keys are source image names and values are their digests.
//	An error if the mirroring process fails for any reason, otherwise nil.
func MirrorImageList(ctx *appcontext.AppContext, imagesList types.ImagesList) (map[string]string, map[string]string, error) {
	// Track failed images for GitHub Actions
	failedImages := make([]map[string]string, 0)
	mirroredImages := make(map[string]string)
	imageDigests := make(map[string]string)

	for _, img := range imagesList.Images {
		log.Debug().Str("name", img.Name).Str("source", img.Source).Msg("Processing image")

		tag, err := getImageTag(img)
		if err != nil {
			log.Error().Err(err).Str("image", img.Source).Msg("Failed to get image tag")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}
		targetRepoPath := fmt.Sprintf("%s/%s:%s", ctx.Config.GCP.GARRepoContainers, img.Name, tag)

		if ctx.DryRun {
			log.Info().
				Str("equivalent command", fmt.Sprintf("oras cp %s %s", img.Source, targetRepoPath)).
				Msg("Dry-run: Would mirror image to GAR")
			mirroredImages[img.Source] = targetRepoPath
			imageDigests[img.Source] = "dry-run-digest"
			continue
		}

		// Initialize ORAS source and target registries
		// Equivalent to: oras cp <source> <target>
		sourceRepo, err := remote.NewRepository(img.Source)
		if err != nil {
			log.Error().Err(err).Str("source", img.Source).Msg("Failed to initialize source repository")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		targetRepo, err := remote.NewRepository(targetRepoPath)
		if err != nil {
			log.Error().Err(err).Str("target", targetRepoPath).Msg("Failed to initialize target repository")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		// Authenticate with Google Artifact Registry
		cmd := exec.Command("gcloud", "auth", "print-access-token")
		token, err := cmd.Output()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get gcloud access token")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		targetRepo.Client = &auth.Client{
			Client: http.DefaultClient,
			Cache:  auth.NewCache(),
			Credential: func(ctx context.Context, s string) (auth.Credential, error) {
				return auth.Credential{
					AccessToken: strings.TrimSpace(string(token)),
				}, nil
			},
		}

		// Check if image already exists in GAR (idempotency)
		sourceDesc, err := sourceRepo.Resolve(context.Background(), sourceRepo.Reference.Reference)
		if err != nil {
			log.Error().Err(err).Str("source", img.Source).Msg("Failed to resolve source image")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		mirroredImages[img.Source] = targetRepoPath
		imageDigests[img.Source] = sourceDesc.Digest.String()

		targetDesc, err := targetRepo.Resolve(context.Background(), targetRepo.Reference.Reference)
		if err == nil && targetDesc.Digest == sourceDesc.Digest {
			log.Info().Str("name", img.Name).Str("digest", sourceDesc.Digest.String()).Msg("Image already exists in GAR, skipping")
			continue
		} else if err == nil && targetDesc.Digest != sourceDesc.Digest && ctx.Config.Options.NotifyTagMutations {
			// TODO test this scenario
			log.Warn().
				Str("name", img.Name).
				Str("source_digest", sourceDesc.Digest.String()).
				Str("target_digest", targetDesc.Digest.String()).
				Msg("Tag points to different digest in GAR, please manually check")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": fmt.Errorf("image %s tag points to different digest in GAR, please manually check", img.Source).Error()})
		}

		// Mirror the image
		// Equivalent to: oras cp <source> <target>
		_, err = oras.Copy(context.Background(), sourceRepo, sourceRepo.Reference.Reference, targetRepo, targetRepo.Reference.Reference, oras.DefaultCopyOptions)
		if err != nil {
			log.Error().Err(err).Str("source", img.Source).Str("target", targetRepoPath).Msg("Failed to mirror image")
			failedImages = append(failedImages, map[string]string{"name": img.Name, "error": err.Error()})
			continue
		}

		log.Info().Str("name", img.Name).
			Str("source", img.Source).
			Str("target", targetRepoPath).Str("tag", sourceRepo.Reference.Reference).
			Msg("Successfully mirrored image to GAR.")
	}

	// Log failed images in JSON format for GitHub Actions
	if len(failedImages) > 0 {
		failedJSON, _ := json.Marshal(map[string][]map[string]string{"failed_images": failedImages})
		log.Warn().RawJSON("failed_images", failedJSON).Msg("Some images failed to mirror")
	}

	return mirroredImages, imageDigests, nil
}

// getImageTag extracts the tag from an image source string.
// It takes an image object as input.
// It returns the image tag and an error if the tag cannot be extracted.
func getImageTag(img types.Image) (string, error) {
	if img.Source == "" {
		return "", fmt.Errorf("image source cannot be empty")
	}
	if !strings.Contains(img.Source, ":") {
		return "", fmt.Errorf("image source must contain a tag")
	}
	return strings.Split(img.Source, ":")[1], nil
}
