package cmdutils

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/charts"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/datastructures"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/images"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/sbom/chartscanner"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MirrorImages mirrors a list of container images to a Google Artifact Registry.
// It takes an application context and a cobra command as input.
func MirrorImages(ctx *appcontext.AppContext, _ *cobra.Command) {
	imagesFile := viper.GetString("images")
	if imagesFile == "" {
		log.Fatal().Msg("Images file path is required, please provide via --images flag")
	}
	if ctx.DryRun {
		log.Info().Msg("Dry-run: Would mirror images to GAR")
	}
	if _, _, err := images.MirrorImagesFromFile(ctx, imagesFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to mirror images")
	}
}

// MirrorCharts mirrors a list of Helm charts and their associated container images to a Google Artifact Registry.
// It takes an application context and a cobra command as input.
// It returns an error if the mirroring fails.
func MirrorCharts(ctx *appcontext.AppContext, cmd *cobra.Command) error {
	chartsFile := viper.GetString("charts")
	err := validateChartsFlag(chartsFile)
	if err != nil {
		return err
	}
	if ctx.DryRun {
		log.Info().Msg("Running in dry-run mode: nothing will be mirrored to GAR")
	}
	successfulCharts, failedCharts, err := charts.MirrorHelmCharts(ctx, chartsFile)
	if err != nil {
		return fmt.Errorf("failed to mirror charts: %w", err)
	}

	if !viper.GetBool("skip_image_mirroring") {
		log.Debug().Msg("mirror images to GAR")
		imageListByChart, err := chartscanner.ExtractImagesFromCharts(ctx, chartsFile)
		if err != nil {
			return fmt.Errorf("failed to extract images from charts: %w", err)
		}
		log.Info().Interface("images", imageListByChart).Msg("Images extracted from charts")

		sortedImages := datastructures.DeduplicateAndSortImages(imageListByChart)
		var imagesList types.ImagesList
		imagesList.Images = sortedImages
		imagesPushed, _, err := images.MirrorImages(ctx, imagesList)
		if err != nil {
			return fmt.Errorf("Failed to mirror images: %w", err)
		}
		log.Debug().Interface("images pushed", imagesPushed).Msg("Mirroring images")

		var values []string
		for _, value := range imagesPushed {
			values = append(values, value)
		}
		fmt.Println("Images pushed:\n", strings.Join(values, "\n "))
	}

	fmt.Println("Charts pushed:\n", strings.Join(successfulCharts, "\n "))
	if len(failedCharts) > 0 {
		fmt.Println("Charts failed to push:\n", strings.Join(failedCharts, "\n"))
	}
	PrintDryRunMessage(ctx)

	return nil
}

var ErrMissingRequiredParam = errors.New("missing required parameter")

// ExtractImagesFromHelmCharts extracts the container images from a list of Helm charts.
// It takes an application context and a cobra command as input.
// It returns an error if the extraction fails.
func ExtractImagesFromHelmCharts(ctx *appcontext.AppContext, cmd *cobra.Command) error {
	chartsFile := viper.GetString("charts")
	outputFile := viper.GetString("output-file")

	err := validateFlagsExtractImagesFromHelmCharts(chartsFile, outputFile)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Listing images for charts in: %s\n", chartsFile)
	imageListByChart, err := chartscanner.ExtractImagesFromCharts(ctx, chartsFile)
	if err != nil {
		return fmt.Errorf("failed to extract images from charts: %w", err)
	}
	log.Info().Interface("images", imageListByChart).Msg("Images extracted from charts")

	if outputFile != "" {
		sortedImages := datastructures.DeduplicateAndSortImages(imageListByChart)

		err := datastructures.WriteImagesToFile(sortedImages, outputFile)
		if err != nil {
			return fmt.Errorf("failed to write images to file %s: %w", outputFile, err)
		}
	}

	return nil
}

// validateFlagsExtractImagesFromHelmCharts validates the flags for the `extract-images-from-helm-charts` command.
// It takes the charts file path and the output file path as input.
// It returns an error if the flags are invalid.
func validateFlagsExtractImagesFromHelmCharts(chartsFile string, outputFile string) error {
	err := validateChartsFlag(chartsFile)
	if err != nil {
		return err
	}
	if outputFile != "" {
		ext := strings.ToLower(filepath.Ext(outputFile))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return fmt.Errorf("unsupported file extension: %s, must be .json, .yaml or .yml", ext)
		}
	}
	return nil
}

// validateChartsFlag validates the `--charts` flag.
// It takes the charts file path as input.
// It returns an error if the flag is invalid.
func validateChartsFlag(chartsFile string) error {
	if chartsFile == "" {
		return fmt.Errorf("%w: %s", ErrMissingRequiredParam, "charts file path")
	}
	return nil
}
