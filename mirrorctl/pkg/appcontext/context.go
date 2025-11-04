package appcontext

import (
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/config"
)

// AppContext holds shared application state, such as configuration and flags.
type AppContext struct {
	Config *config.Config // The application configuration.
	DryRun bool           // A flag to simulate actions without executing them.
}

// NewAppContext creates a new application context.
// It takes the application configuration and the dry-run flag as input.
// It returns a pointer to the new AppContext.
func NewAppContext(cfg *config.Config, dryRun bool) *AppContext {
	return &AppContext{
		Config: cfg,
		DryRun: dryRun,
	}
}
