package cmd

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/cmdutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chartImagesCmd represents the chart-images command
var chartImagesCmd = &cobra.Command{
	Use:   "chart-images",
	Short: "List all images used by a Helm chart",
	Long:  `This command lists all container images referenced within a given Helm chart or a set of charts defined in a YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutils.ExtractImagesFromHelmCharts(ctx, cmd)
	},
}

func init() {
	listCmd.AddCommand(chartImagesCmd)

	chartImagesCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	_ = viper.BindPFlag("charts", chartImagesCmd.Flags().Lookup("charts"))

	chartImagesCmd.Flags().String("output-file", "", "Path to file to store the list of images (e.g., images.yaml, images.json)")
	_ = viper.BindPFlag("output-file", chartImagesCmd.Flags().Lookup("output-file"))
}
