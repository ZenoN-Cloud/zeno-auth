package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(cfg *Config) error {
	env := cfg.Env

	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var writers []io.Writer

	// Console vs JSON
	if cfg.Log.Format == "console" && env != "production" && env != "prod" {
		writers = append(
			writers, zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: "15:04:05",
			},
		)
	} else {
		writers = append(writers, os.Stdout)
	}

	// File logging (only non-production and when not in containerized environment)
	// Skip file logging if running in Cloud Run or similar container platforms
	if env != "production" && env != "prod" && cfg.Log.File != "" && os.Getenv("K_SERVICE") == "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Log.File), 0o750); err != nil {
			log.Warn().Err(err).Msg("Failed to create log directory, skipping file logging")
		} else {
			file, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to open log file, skipping file logging")
			} else {
				// Note: file handle kept open for application lifetime
				writers = append(writers, file)
			}
		}
	}

	// Multi writer
	multi := io.MultiWriter(writers...)

	// Global logger
	log.Logger = zerolog.New(multi).
		With().
		Timestamp().
		Str("app", cfg.AppName).
		Str("env", cfg.Env).
		Logger()

	return nil
}
