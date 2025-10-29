package cmd

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/cmdutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mirrorImagesCmd represents the images command
var mirrorImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Mirror container images to GAR",
	Long:  `Mirrors container images specified in a YAML file to Google Artifact Registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutils.MirrorImages(ctx, cmd)
	},
}

func init() {
	mirrorCmd.AddCommand(mirrorImagesCmd)
	mirrorImagesCmd.Flags().String("images", "", "Path to YAML file with list of container images")
	_ = viper.BindPFlag("images", mirrorImagesCmd.Flags().Lookup("images"))
}
