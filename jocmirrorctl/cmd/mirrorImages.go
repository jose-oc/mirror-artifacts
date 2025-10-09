/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/cmdutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mirrorImagesCmd represents the images command
var mirrorImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Mirror container images to GAR",
	Long:  `Mirrors container images specified in a YAML file to Google Artifact Registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO remove this chivato
		fmt.Println("images called")
		cmdutils.MirrorImages(ctx, cmd)
	},
}

func init() {
	mirrorCmd.AddCommand(mirrorImagesCmd)
	mirrorImagesCmd.Flags().String("images", "", "Path to YAML file with list of container images")
	viper.BindPFlag("images", mirrorImagesCmd.Flags().Lookup("images"))
}
