/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/appcontext"
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/config"
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool
var dryRun bool
var logger = logging.NewLogger()
var cfg *config.Config
var ctx *appcontext.AppContext

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jocmirrorctl",
	Short: "jocMirrorCtl mirrors Helm charts and container images to Google Artifact Registry",
	Long: `jocMirrorCtl is a Go-based CLI tool that automates the mirroring of Helm charts and their container images into Google Artifact Registry (GAR). 
	It supports provenance tracking, SBOM generation, and signing, optimized for GitHub Actions.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with verbose setting
		logging.SetLogLevel(logger, verbose)
		// Load configuration
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to load configuration")
		}
		logger.Debug().Interface("config", cfg).Msg("Configuration loaded")
		// Initialize app context
		ctx = appcontext.NewAppContext(logger, cfg, dryRun)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Fatal().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jocmirrorctl.yaml)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simulate actions without executing")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// TODO remove this toggle flag, it's an example of local flag
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".jocmirrorctl" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".jocmirrorctl")
	}

	// TODO try this out
	viper.SetEnvPrefix("JOCMIRRORCTL")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		logger.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Using config file")
	} else if cfgFile != "" {
		logger.Error().Err(err).Str("config_file", cfgFile).Msg("Failed to read config file")
	}
}
