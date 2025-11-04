package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config holds the application configuration.
// It is loaded from a configuration file or environment variables.
type Config struct {
	GCP     GCPConfig     `mapstructure:"gcp"`     // GCP-related configuration.
	Options OptionsConfig `mapstructure:"options"` // General options.
}

// GCPConfig holds GCP-related configuration.
// It contains the GCP project ID, region, and the names of the GAR repositories for charts and containers.
type GCPConfig struct {
	ProjectID         string `mapstructure:"project_id"`          // The GCP project ID.
	Region            string `mapstructure:"region"`              // The GCP region.
	GARRepoCharts     string `mapstructure:"gar_repo_charts"`     // The name of the GAR repository for Helm charts.
	GARRepoContainers string `mapstructure:"gar_repo_containers"` // The name of the GAR repository for container images.
}

// OptionsConfig holds general options for the application.
// It contains a suffix to be appended to the version of the mirrored charts,
// a flag to keep temporary directories, and a flag to notify about tag mutations.
type OptionsConfig struct {
	Suffix             string `mapstructure:"suffix"`               // A suffix to be appended to the version of the mirrored charts.
	KeepTempDir        bool   `mapstructure:"keep_temp_dir"`        // A flag to keep temporary directories for debugging purposes.
	NotifyTagMutations bool   `mapstructure:"notify_tag_mutations"` // A flag to notify about tag mutations.
}

// LoadConfig loads the application configuration from a configuration file or environment variables.
// It returns a pointer to a Config object and an error if the configuration cannot be loaded.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal config")
		return nil, err
	}
	log.Debug().Str("config_file", viper.ConfigFileUsed()).Interface("config", cfg).Msg("Configuration loaded")
	return &cfg, nil
}
