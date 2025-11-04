package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List various artifacts",
	Long:  `Commands to list various artifacts, such as images used by Helm charts.`,
}

func init() {
	sbomCmd.AddCommand(listCmd)
}
