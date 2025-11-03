package cmd

import (
	"github.com/spf13/cobra"
)

// sbomCmd represents the sbom command
var sbomCmd = &cobra.Command{
	Use:   "sbom",
	Short: "Software Bill of Materials (SBOM) related commands",
	Long:  `Provides commands for generating and managing Software Bill of Materials (SBOMs) for various artifacts.`,
}

func init() {
	rootCmd.AddCommand(sbomCmd)
}
