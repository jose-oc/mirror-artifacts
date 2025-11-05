package cmd

import (
	"log"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/cmdutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mirrorChartsCmd represents the mirrorCharts command
var mirrorChartsCmd = &cobra.Command{
	Use:   "charts",
	Short: "Mirror Helm charts to GAR",
	Long:  `Mirrors Helm charts specified in a YAML file to Google Artifact Registry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutils.MirrorCharts(ctx, cmd)
	},
}

func init() {
	mirrorCmd.AddCommand(mirrorChartsCmd)
	mirrorChartsCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	if err := viper.BindPFlag("charts", mirrorChartsCmd.Flags().Lookup("charts")); err != nil {
		log.Fatalf("Error binding flag: %v", err)
	}
	mirrorChartsCmd.Flags().Bool("skip-image-mirroring", false, "Skip mirroring the container images used by the Helm charts")
	if err := viper.BindPFlag("skip_image_mirroring", mirrorChartsCmd.Flags().Lookup("skip-image-mirroring")); err != nil {
		log.Fatalf("Error binding flag: %v", err)
	}
}
