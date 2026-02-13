package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(env string) {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Logger initialized in development mode")
	} else {
		log.Logger = zerolog.New(os.Stderr).
			With().
			Timestamp().
			Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Info().Str("env", env).Msg("Logger initialized in production mode")
	}
}

func Get() zerolog.Logger {
	return log.Logger
}

func WithContext(fields map[string]interface{}) zerolog.Logger {
	logger := log.Logger
	for key, value := range fields {
		logger = logger.With().Interface(key, value).Logger()
	}
	return logger
}

func SetLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
