package cmd

import (
	"fmt"

	sbom_charts "github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/sbom/charts"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chartImagesCmd represents the chart-images command
var chartImagesCmd = &cobra.Command{
	Use:   "chart-images",
	Short: "List all images used by a Helm chart",
	Long:  `This command lists all container images referenced within a given Helm chart or a set of charts defined in a YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		chartsFile := viper.GetString("charts")
		log.Debug().Msgf("Charts file: %s", chartsFile)
		if chartsFile == "" {
			fmt.Println("Error: --charts flag is required")
			return
		}
		fmt.Printf("Listing images for charts in: %s\n", chartsFile)
		imageList, _ := sbom_charts.ExtractImagesFromCharts(chartsFile)
		log.Debug().Interface("images", imageList).Msg("Images extracted from charts")
	},
}

func init() {
	listCmd.AddCommand(chartImagesCmd)

	chartImagesCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	_ = viper.BindPFlag("charts", chartImagesCmd.Flags().Lookup("charts"))
}
