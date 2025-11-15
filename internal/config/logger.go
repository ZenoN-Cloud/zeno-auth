package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetupLogger(cfg *Config) error {
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var writers []io.Writer
	writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout})

	if cfg.Env == "dev" && cfg.Log.FilePath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Log.FilePath), 0755); err != nil {
			return err
		}
		
		file, err := os.OpenFile(cfg.Log.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		writers = append(writers, file)
	}

	multi := io.MultiWriter(writers...)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	return nil
}