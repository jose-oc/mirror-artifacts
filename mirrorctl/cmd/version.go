/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the `version` command.
// It is used to print the version of the application.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of mirrorctl",
	Long:  `Prints the version number of the mirrorctl CLI tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", version.AppName, version.Version)
	},
}

// init initializes the `version` command.
func init() {
	rootCmd.AddCommand(versionCmd)
}
