package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	var logWriter io.Writer

	switch logOutput := os.Getenv("LOG_OUTPUT"); logOutput {
	case "":
		fallthrough
	case "stderr":
		logWriter = os.Stderr
	case "stdout":
		logWriter = os.Stdout
	default:
		f, err := os.OpenFile(logOutput, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o666)
		if err != nil {
			panic(fmt.Errorf("failed to open log file: %q: %w", logOutput, err))
		}
		defer func() {
			_ = f.Sync()
			_ = f.Close()
		}()
		logWriter = f
	}

	switch logFormat := os.Getenv("LOG_FORMAT"); logFormat {
	case "":
		fallthrough
	case "console":
		logWriter = zerolog.ConsoleWriter{
			Out:        logWriter,
			TimeFormat: "2006-01-02T15:04:05.999Z07:00",
		}
	case "console-plain":
		logWriter = zerolog.ConsoleWriter{
			Out:        logWriter,
			TimeFormat: "2006-01-02T15:04:05.999Z07:00",
			NoColor:    true,
		}
	case "raw":
		// pass
	case "cbor":
		// pass
	case "json":
		// pass
	default:
		panic(fmt.Errorf("unknown log format %q, must be one of \"console\" or \"json\"", logFormat))
	}

	switch logLevel := os.Getenv("LOG_LEVEL"); logLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "":
		fallthrough
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		fallthrough
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "err":
		fallthrough
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		panic(fmt.Errorf("unknown log level %q, must be one of \"debug\", \"info\", \"warn\", \"error\"", logLevel))
	}

	log.Logger = zerolog.New(logWriter).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.DurationFieldUnit = time.Second
	zerolog.DurationFieldInteger = false
	zerolog.DefaultContextLogger = &log.Logger
}
