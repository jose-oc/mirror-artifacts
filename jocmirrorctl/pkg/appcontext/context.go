package appcontext

import (
	"github.com/jose-oc/poc-mirror-artifacts/jocmirrorctl/pkg/config"
	"github.com/rs/zerolog"
)

// AppContext holds shared application state
type AppContext struct {
	Logger *zerolog.Logger
	Config *config.Config
	DryRun bool
}

// NewAppContext creates a new application context
func NewAppContext(logger *zerolog.Logger, cfg *config.Config, dryRun bool) *AppContext {
	return &AppContext{
		Logger: logger,
		Config: cfg,
		DryRun: dryRun,
	}
}
