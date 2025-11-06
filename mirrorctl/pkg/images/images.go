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

// MirrorImagesFromFile mirrors a list of container images from a file to a Google Artifact Registry.
// It takes an application context and the path to the file containing the list of images as input.
//
// It returns three values:
//   - A map of strings to strings, where the keys are the source image names and the values are the destination image names.
//   - A map of strings to strings, where the keys are the source image names and the values are the image digests.
//   - An error if the mirroring fails.
func MirrorImagesFromFile(ctx *appcontext.AppContext, imagesFile string) (map[string]string, []types.FailedImage, error) {
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
	return MirrorImages(ctx, imagesList)
}

// MirrorImages mirrors a list of container images to a Google Artifact Registry.
// It takes an application context and a list of images as input.
//
// It returns three values:
//   - A map of strings to strings, where the keys are the source image names and the values are the destination image names.
//   - A list of types.FailedImage, of the images that failed to mirror. Each element of the list is a map with two keys: image and error, where error is the error message.
//   - An error if the mirroring fails.
func MirrorImages(ctx *appcontext.AppContext, imagesList types.ImagesList) (map[string]string, []types.FailedImage, error) {
	// Track failed images with error reasons
	failedImages := make([]types.FailedImage, 0)
	mirroredImages := make(map[string]string)

	for _, img := range imagesList.Images {
		log.Debug().Str("name", img.Name).Str("source", img.Source).Msg("Processing image")

		// Define a helper function to handle failure for cleaner flow
		handleFailure := func(err error, msg string) {
			log.Error().Err(err).Str("image", img.Source).Msg(msg)
			failedImages = append(failedImages, types.FailedImage{
				Image: img,
				Error: err.Error(),
			})
		}

		tag, err := getImageTag(img)
		if err != nil {
			handleFailure(err, "Failed to get image tag")
			continue
		}
		targetRepoPath := fmt.Sprintf("%s/%s:%s", ctx.Config.GCP.GARRepoContainers, img.Name, tag)

		if ctx.DryRun {
			log.Info().
				Str("equivalent command", fmt.Sprintf("oras cp %s %s", img.Source, targetRepoPath)).
				Msg("Dry-run: Would mirror image to GAR")
			mirroredImages[img.Source] = targetRepoPath
			continue
		}

		// Initialize ORAS source and target registries
		// Equivalent to: oras cp <source> <target>
		sourceRepo, err := remote.NewRepository(img.Source)
		if err != nil {
			handleFailure(err, "Failed to initialize source repository")
			continue
		}

		targetRepo, err := remote.NewRepository(targetRepoPath)
		if err != nil {
			handleFailure(err, "Failed to initialize target repository")
			continue
		}

		// Authenticate with Google Artifact Registry
		cmd := exec.Command("gcloud", "auth", "print-access-token")
		token, err := cmd.Output()
		if err != nil {
			handleFailure(err, "Failed to get gcloud access token")
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
			handleFailure(err, "Failed to resolve source image")
			continue
		}

		mirroredImages[img.Source] = targetRepoPath

		targetDesc, err := targetRepo.Resolve(context.Background(), targetRepo.Reference.Reference)
		if err == nil && targetDesc.Digest == sourceDesc.Digest {
			log.Info().Str("name", img.Name).Str("digest", sourceDesc.Digest.String()).Msg("Image already exists in GAR, skipping")
			continue
		} else if err == nil && targetDesc.Digest != sourceDesc.Digest && ctx.Config.Options.NotifyTagMutations {
			// TODO test this scenario
			mirrorErr := fmt.Errorf("image %s tag points to different digest in GAR, please manually check", img.Source)
			log.Warn().
				Str("name", img.Name).
				Str("source_digest", sourceDesc.Digest.String()).
				Str("target_digest", targetDesc.Digest.String()).
				Msg("Tag points to different digest in GAR, please manually check")
			handleFailure(mirrorErr, "Tag mutation detected")
			continue
		}

		// Mirror the image
		// Equivalent to: oras cp <source> <target>
		_, err = oras.Copy(context.Background(), sourceRepo, sourceRepo.Reference.Reference, targetRepo, targetRepo.Reference.Reference, oras.DefaultCopyOptions)
		if err != nil {
			handleFailure(err, "Failed to mirror image")
			continue
		}

		log.Info().Str("name", img.Name).
			Str("source", img.Source).
			Str("target", targetRepoPath).Str("tag", sourceRepo.Reference.Reference).
			Msg("Successfully mirrored image to GAR.")
	}

	// Log failed images in JSON format for GitHub Actions
	if len(failedImages) > 0 {
		failedJSON, _ := json.Marshal(map[string][]types.FailedImage{"failed_images": failedImages})
		log.Warn().RawJSON("failed_images", failedJSON).Msg("Some images failed to mirror")
	}

	return mirroredImages, failedImages, nil
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
