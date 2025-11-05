package cmd

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/cmdutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chartImagesCmd represents the `chart-images` command.
// It is used to list all the container images used by a Helm chart.
var chartImagesCmd = &cobra.Command{
	Use:   "chart-images",
	Short: "List all images used by a Helm chart",
	Long:  `This command lists all container images referenced within a given Helm chart or a set of charts defined in a YAML file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutils.ExtractImagesFromHelmCharts(ctx, cmd)
	},
}

// init initializes the `chart-images` command and its flags.
func init() {
	listCmd.AddCommand(chartImagesCmd)

	chartImagesCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	_ = viper.BindPFlag("charts", chartImagesCmd.Flags().Lookup("charts"))

	chartImagesCmd.Flags().String("output-dir", "", "Directory path to store the list of images per chart")
	_ = viper.BindPFlag("output_dir", chartImagesCmd.Flags().Lookup("output-dir"))
}
