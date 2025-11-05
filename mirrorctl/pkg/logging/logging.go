package logging

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/jose-oc/mirror-artifacts/mirrorctl/pkg/version"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// CallerHook adds file and line number to log events for Error, Fatal, and Panic levels.
type CallerHook struct{}

func (h CallerHook) Run(e *zerolog.Event, level zerolog.Level, _ string) {
	if level >= zerolog.ErrorLevel {
		// Adjust the skip frame as needed.
		// 0: runtime.Caller
		// 1: h.Run
		// 2: zerolog's internal call to hook
		// 3: The actual log call (e.g., log.Error().Msg("..."))
		if _, file, line, ok := runtime.Caller(3); ok {
			e.Str("caller", file+":"+strconv.Itoa(line))
		}
	}
}

func SetupLogger() error {
	// Set global time format for zerolog. This applies to all JSON output.
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logLevel, err := zerolog.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// File logger setup
	logFile := viper.GetString("log_file")
	if logFile == "" {
		logFile = version.AppName + ".log"
	}
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to open log file: %s", logFile)
		return fmt.Errorf("failed to open log file: %s. %w", logFile, err)
	}

	var writers []io.Writer

	// 1. Setup Console Writer for STDOUT (for "verbose" mode)
	if viper.GetBool("verbose") {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04.0000",
			NoColor:    viper.GetBool("no_color"),
		}
		writers = append(writers, consoleWriter)
	}

	// 2. Setup File Writer (conditional based on "prod_mode")
	if viper.GetBool("prod_mode") {
		// Production mode: Use the default zerolog writer for the file, which is JSON.
		writers = append(writers, file)
	} else {
		// Non-production mode: Use ConsoleWriter for the file to get non-JSON,
		// human-readable output with the full time format.
		fileConsoleWriter := zerolog.ConsoleWriter{
			Out: file,
			// Use the full RFC3339Nano format for the file log
			TimeFormat: zerolog.TimeFieldFormat,
			NoColor:    true,
		}
		writers = append(writers, fileConsoleWriter)
	}

	// Combine all writers (STDOUT if verbose, and the file writer)
	multi := zerolog.MultiLevelWriter(writers...)

	// Create the logger
	var logger zerolog.Logger
	logger = zerolog.New(multi).With().Timestamp().Logger()

	// Add the custom CallerHook
	logger = logger.Hook(CallerHook{})

	// Set the global logger
	log.Logger = logger.Level(logLevel)
	return nil
}
