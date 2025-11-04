package cmdutils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/charts"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/images"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/sbom/chartscanner"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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

	if chartsFile == "" {
		return fmt.Errorf("%w: %s", ErrMissingRequiredParam, "charts file path")
	}

	log.Debug().Msgf("Listing images for charts in: %s\n", chartsFile)
	imageListByChart, err := chartscanner.ExtractImagesFromCharts(ctx, chartsFile)
	if err != nil {
		return fmt.Errorf("failed to extract images from charts: %w", err)
	}
	log.Info().Interface("images", imageListByChart).Msg("Images extracted from charts")

	if outputFile != "" {
		sortedImages := flattenAndSortImageList(imageListByChart)

		err := writeImageListIntoFileJsonOrYaml(sortedImages, outputFile)
		if err != nil {
			return fmt.Errorf("failed to write images to file %s: %w", outputFile, err)
		}
	}

	return nil
}

func writeImageListIntoFileJsonOrYaml(sortedImages []types.Image, outputFile string) error {
	imagesList := types.ImagesList{Images: sortedImages}

	var out []byte
	var err error
	ext := filepath.Ext(outputFile)
	if strings.ToLower(ext) == ".json" {
		out, err = json.MarshalIndent(imagesList, "", "  ")
	} else {
		out, err = yaml.Marshal(imagesList)
	}

	if err != nil {
		//log.Error().Err(err).Msg("Failed to marshal images list")
		return fmt.Errorf("failed to marshal images list: %w", err)
	}

	err = os.WriteFile(outputFile, out, 0644)
	if err != nil {
		//log.Error().Err(err).Str("file", outputFile).Msg("Failed to write images to file")
		return fmt.Errorf("failed to write images to file %s: %w", outputFile, err)
	}
	log.Info().Str("file", outputFile).Msg("Wrote images to file")
	return nil
}

func flattenAndSortImageList(imageListByChart map[string][]types.Image) []types.Image {
	allImages := make(map[string]types.Image)
	for _, imgs := range imageListByChart {
		for _, image := range imgs {
			allImages[image.Source] = image
		}
	}

	var sortedImages []types.Image
	for _, image := range allImages {
		sortedImages = append(sortedImages, image)
	}

	sort.Slice(sortedImages, func(i, j int) bool {
		return sortedImages[i].Source < sortedImages[j].Source
	})
	return sortedImages
}
