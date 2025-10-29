package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	GCP     GCPConfig     `mapstructure:"gcp"`
	Options OptionsConfig `mapstructure:"options"`
}

// GCPConfig holds GCP-related configuration
type GCPConfig struct {
	ProjectID         string `mapstructure:"project_id"`
	Region            string `mapstructure:"region"`
	GARRepoCharts     string `mapstructure:"gar_repo_charts"`
	GARRepoContainers string `mapstructure:"gar_repo_containers"`
}

// OptionsConfig holds general options
type OptionsConfig struct {
	Suffix             string `mapstructure:"suffix"`
	KeepTempDir        bool   `mapstructure:"keep_temp_dir"`
	NotifyTagMutations bool   `mapstructure:"notify_tag_mutations"`
}

// LoadConfig loads configuration from file or environment variables
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal config")
		return nil, err
	}
	log.Debug().Str("config_file", viper.ConfigFileUsed()).Interface("config", cfg).Msg("Configuration loaded")
	return &cfg, nil
}
