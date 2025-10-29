package cmd

import (
	"github.com/spf13/cobra"
)

// mirrorCmd represents the mirror command
var mirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Mirror artifacts to Google Artifact Registry",
	Long:  `Mirror Helm charts and/or container images to Google Artifact Registry (GAR).`,
}

func init() {
	rootCmd.AddCommand(mirrorCmd)
}
