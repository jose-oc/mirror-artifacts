package cmd

import (
	"github.com/spf13/cobra"
)

// sbomCmd represents the `sbom` command, which is the parent of all sbom subcommands.
var sbomCmd = &cobra.Command{
	Use:   "sbom",
	Short: "Software Bill of Materials (SBOM) related commands",
	Long:  `Provides commands for generating and managing Software Bill of Materials (SBOMs) for various artifacts.`,
}

// init initializes the `sbom` command.
func init() {
	rootCmd.AddCommand(sbomCmd)
}
