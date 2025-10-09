package logging

import (
	"os"

	"github.com/rs/zerolog"
)

// NewLogger initializes a new Zerolog logger
func NewLogger() *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	return &logger
}

// SetLogLevel configures the logger based on verbose flag
func SetLogLevel(logger *zerolog.Logger, verbose bool) {
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger.Debug().Msg("Debug logging enabled")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
