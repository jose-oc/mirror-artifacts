package cmd

import (
	"os"
	"strings"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/cmdutils"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/logging"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var dryRun bool
var keepTempDir bool
var cfg *config.Config
var ctx *appcontext.AppContext

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mirrorctl",
	Short: "mirrorctl mirrors Helm charts and container images to Google Artifact Registry",
	Long: `mirrorctl is a CLI tool that automates the mirroring of Helm charts and their container images into Google Artifact Registry (GAR). 
	It supports provenance tracking.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error
		viper.BindPFlags(cmd.Flags())
		cfg, err = config.LoadConfig()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load configuration")
			os.Exit(1)
		}
		// Initialize app context
		ctx = appcontext.NewAppContext(cfg, dryRun)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

// init initializes the command-line flags and binds them to viper.
func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mirrorctl.yaml)")
	rootCmd.PersistentFlags().Bool("prod-mode", false, "Enables production-style JSON logging.")
	rootCmd.PersistentFlags().Bool("log-color", true, "Enables colored output in development mode.")
	rootCmd.PersistentFlags().String("log-level", "info", "Sets the minimum log level (e.g., debug, info, warn, error).")
	rootCmd.PersistentFlags().String("log-file", "", "If set, writes logs to the specified file path instead of the console.")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simulate actions without executing")
	rootCmd.PersistentFlags().BoolVar(&keepTempDir, "keep-temp-dir", false, "Keep temporary directories for inspection")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output.")
	rootCmd.PersistentFlags().Bool("quiet", false, "Suppress all output.")

	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")

	// Bind the flag to viper so it can be accessed via viper
	_ = viper.BindPFlag("prod_mode", rootCmd.PersistentFlags().Lookup("prod-mode"))
	_ = viper.BindPFlag("log_color", rootCmd.PersistentFlags().Lookup("log-color"))
	_ = viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("log_file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindPFlag("options.keep_temp_dir", rootCmd.PersistentFlags().Lookup("keep-temp-dir"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
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

		// Search config in home directory with name ".mirrorctl" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mirrorctl")
	}

	viper.SetEnvPrefix("mirrorctl")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if cfgFile != "" {
				log.Fatal().Err(err).Str("config_file", cfgFile).Msg("Specified config file not found")
			} else {
				log.Fatal().Err(err).Msg("No config file found in default paths (current directory or home directory)")
			}
		} else {
			log.Fatal().Err(err).Msg("Failed to read config file")
		}
	} else {
		// log.Debug().Str("config_file", viper.ConfigFileUsed()).Msg("Using config file")
		cmdutils.PrintConfigFileInfo(ctx)
	}
	err := initLoggers()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize loggers")
		os.Exit(1)
	}
}

// initLoggers initializes the logging system.
func initLoggers() error {
	return logging.SetupLogger()
}
