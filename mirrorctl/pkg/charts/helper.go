package charts

import (
	"fmt"
	"os"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/appcontext"
	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/version"
	"github.com/rs/zerolog/log"
)

func createTempDir(ctx *appcontext.AppContext) (string, error) {
	// Create a temporary directory to download the chart
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("%s-", version.AppName))
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary directory")
		return "", err
	}

	// Only remove the temp directory if KeepTempDir is not set
	if !ctx.Config.Options.KeepTempDir {
		defer os.RemoveAll(tmpDir)
	} else {
		log.Debug().Str("temp_dir", tmpDir).Msg("Keeping temporary directory for inspection")
	}
	return tmpDir, err
}
