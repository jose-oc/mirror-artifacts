/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/cmdutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mirrorChartsCmd represents the mirrorCharts command
var mirrorChartsCmd = &cobra.Command{
	Use:   "charts",
	Short: "Mirror Helm charts to GAR",
	Long:  `Mirrors Helm charts specified in a YAML file to Google Artifact Registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdutils.MirrorCharts(ctx, cmd)
	},
}

func init() {
	mirrorCmd.AddCommand(mirrorChartsCmd)
	mirrorChartsCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	_ = viper.BindPFlag("charts", mirrorChartsCmd.Flags().Lookup("charts"))
}
