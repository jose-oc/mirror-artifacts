package logging

import (
	"io"
	"os"
	"runtime"
	"strconv"
	"time"

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

func SetupLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logLevel, err := zerolog.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	var output io.Writer = os.Stdout
	if logFile := viper.GetString("log-file"); logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to open log file: %s", logFile)
		}
		output = file
	}

	var logger zerolog.Logger
	if viper.GetBool("prod-mode") {
		logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: "15:04.0000",
			NoColor:    !viper.GetBool("log-color"),
		}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}

	// Add the new custom CallerHook
	logger = logger.Hook(CallerHook{})

	log.Logger = logger.Level(logLevel)
}
