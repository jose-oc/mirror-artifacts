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

// mirrorAllCmd represents the all command
var mirrorAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Mirror both Helm charts and container images to GAR",
	Long:  `Mirrors both Helm charts and container images specified in YAML files to Google Artifact Registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("all called")
		cmdutils.MirrorAll(ctx, cmd)
	},
}

func init() {
	mirrorCmd.AddCommand(mirrorAllCmd)
	mirrorAllCmd.Flags().String("images", "", "Path to YAML file with list of container images")
	mirrorAllCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	viper.BindPFlag("images", mirrorAllCmd.Flags().Lookup("images"))
	viper.BindPFlag("charts", mirrorAllCmd.Flags().Lookup("charts"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// allCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// allCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
