/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/cmdutils"
	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify mirrored artifacts in GAR",
	Long:  `Verifies that mirrored Helm charts and container images in Google Artifact Registry are correctly configured and signed.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("verify called")
		cmdutils.Verify(ctx, cmd)
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// verifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// verifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
