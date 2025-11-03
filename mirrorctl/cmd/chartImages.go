package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/sbom/chartscanner"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// chartImagesCmd represents the chart-images command
var chartImagesCmd = &cobra.Command{
	Use:   "chart-images",
	Short: "List all images used by a Helm chart",
	Long:  `This command lists all container images referenced within a given Helm chart or a set of charts defined in a YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

func init() {
	listCmd.AddCommand(chartImagesCmd)

	chartImagesCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	_ = viper.BindPFlag("charts", chartImagesCmd.Flags().Lookup("charts"))

	chartImagesCmd.Flags().String("output-file", "", "Path to file to store the list of images (e.g., images.yaml, images.json)")
	_ = viper.BindPFlag("output-file", chartImagesCmd.Flags().Lookup("output-file"))
}
