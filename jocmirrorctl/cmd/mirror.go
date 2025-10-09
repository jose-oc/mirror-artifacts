/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// mirrorCmd represents the mirror command
var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Mirror artifacts to Google Artifact Registry",
	Long:  `Mirror Helm charts and/or container images to Google Artifact Registry (GAR).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mirror called")
	},
}

func init() {
	rootCmd.AddCommand(mirrorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mirrorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mirrorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
