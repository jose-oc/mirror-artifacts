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

// mirrorChartsCmd represents the mirrorCharts command
var mirrorChartsCmd = &cobra.Command{
	Use:   "charts",
	Short: "Mirror Helm charts to GAR",
	Long:  `Mirrors Helm charts specified in a YAML file to Google Artifact Registry.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mirrorCharts called")
		cmdutils.MirrorCharts(ctx, cmd)
	},
}

func init() {
	mirrorCmd.AddCommand(mirrorChartsCmd)
	mirrorChartsCmd.Flags().String("charts", "", "Path to YAML file with list of Helm charts")
	viper.BindPFlag("charts", mirrorChartsCmd.Flags().Lookup("charts"))

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mirrorChartsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mirrorChartsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
