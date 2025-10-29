package appcontext

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
)

// AppContext holds shared application state
type AppContext struct {
	Config *config.Config
	DryRun bool
}

// NewAppContext creates a new application context
func NewAppContext(cfg *config.Config, dryRun bool) *AppContext {
	return &AppContext{
		Config: cfg,
		DryRun: dryRun,
	}
}
