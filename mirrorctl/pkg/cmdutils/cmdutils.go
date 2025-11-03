package cmdutils

import (
	"encoding/json"
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

func ExtractImagesFromHelmCharts(ctx *appcontext.AppContext, cmd *cobra.Command) {
	chartsFile := viper.GetString("charts")
	outputFile := viper.GetString("output-file")

	log.Debug().Msgf("Charts file: %s", chartsFile)
	if chartsFile == "" {
		fmt.Println("Error: --charts flag is required")
		return
	}

	fmt.Printf("Listing images for charts in: %s\n", chartsFile)
	imageListByChart, err := chartscanner.ExtractImagesFromCharts(ctx, chartsFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract images from charts")
		return
	}
	log.Info().Interface("images", imageListByChart).Msg("Images extracted from charts")

	if outputFile != "" {
		allImages := make(map[string]types.Image)
		for _, images := range imageListByChart {
			for _, image := range images {
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

		imagesList := types.ImagesList{Images: sortedImages}

		var out []byte
		ext := filepath.Ext(outputFile)
		if strings.ToLower(ext) == ".json" {
			out, err = json.MarshalIndent(imagesList, "", "  ")
		} else {
			out, err = yaml.Marshal(imagesList)
		}

		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal images list")
			return
		}

		err = os.WriteFile(outputFile, out, 0644)
		if err != nil {
			log.Error().Err(err).Str("file", outputFile).Msg("Failed to write images to file")
			return
		}
		log.Info().Str("file", outputFile).Msg("Wrote images to file")
	}
}
