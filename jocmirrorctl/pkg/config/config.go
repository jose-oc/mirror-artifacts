package config

import (
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/logging"
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
	DryRun             bool   `mapstructure:"dry_run"`
	Verbose            bool   `mapstructure:"verbose"`
	Suffix             string `mapstructure:"suffix"`
	NotifyTagMutations bool   `mapstructure:"notify_tag_mutations"`
	SBOMFormat         string `mapstructure:"sbom_format"`
	ProvenanceStore    string `mapstructure:"provenance_store"`
}

// LoadConfig loads configuration from file or environment variables
func LoadConfig(cfgFile string) (*Config, error) {
	logger := logging.NewLogger()
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		logger.Error().Err(err).Msg("Failed to unmarshal config")
		return nil, err
	}
	logger.Info().Str("config_file", viper.ConfigFileUsed()).Msg("Configuration loaded")
	return &cfg, nil
}
