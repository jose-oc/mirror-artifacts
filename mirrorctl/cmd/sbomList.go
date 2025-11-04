package cmd

import (
	"github.com/spf13/cobra"
)

// listCmd represents the `list` command, which is the parent of all list subcommands.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List various artifacts",
	Long:  `Commands to list various artifacts, such as images used by Helm charts.`,
}

// init initializes the `list` command.
func init() {
	sbomCmd.AddCommand(listCmd)
}
