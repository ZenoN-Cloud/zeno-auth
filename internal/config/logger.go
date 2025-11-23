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

	// File logging (only non-production)
	if env != "production" && env != "prod" && cfg.Log.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Log.File), 0o755); err != nil {
			return err
		}

		file, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
		if err != nil {
			return err
		}

		writers = append(writers, file)
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
