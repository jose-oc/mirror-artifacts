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
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// MirrorImages handles the `mirror images` subcommand
func MirrorImages(ctx *appcontext.AppContext, _ *cobra.Command) {
	imagesFile := viper.GetString("images")
	if imagesFile == "" {
		log.Fatal().Msg("Images file path is required, please provide via --images flag")
	}
	if ctx.DryRun {
		log.Info().Msg("Dry-run: Would mirror images to GAR")
	}
	if _, _, err := images.MirrorImages(ctx, imagesFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to mirror images")
	}
}

// MirrorCharts handles the `mirror charts` subcommand
func MirrorCharts(ctx *appcontext.AppContext, cmd *cobra.Command) {
	chartsFile := viper.GetString("charts")
	if chartsFile == "" {
		log.Fatal().Msg("Charts file path is required")
	}
	if ctx.DryRun {
		log.Info().Msg("Dry-run: Would mirror charts to GAR")
	}
	if err := charts.MirrorHelmCharts(ctx, chartsFile); err != nil {
		log.Fatal().Err(err).Msg("Failed to mirror charts")
	}
}

var ErrMissingRequiredParam = errors.New("missing required parameter")

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

// validateFlagsExtractImagesFromHelmCharts validates the flags passed to the `ExtractImagesFromHelmCharts` function
// These are:
// - chartsFile: Path to the yaml charts file
// - outputFile: (Optional) Path to the output file, it has to have the .json, .yml or .yaml extension
// Returns an error if the flags are invalid
func validateFlagsExtractImagesFromHelmCharts(chartsFile string, outputFile string) error {
	if chartsFile == "" {
		return fmt.Errorf("%w: %s", ErrMissingRequiredParam, "charts file path")
	}
	if outputFile != "" {
		ext := strings.ToLower(filepath.Ext(outputFile))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			return fmt.Errorf("unsupported file extension: %s, must be .json, .yaml or .yml", ext)
		}
	}
	return nil
}
