package app

import (
	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if err := config.SetupLogger(cfg); err != nil {
		return nil, err
	}

	return &App{cfg: cfg}, nil
}

func (a *App) Run() error {
	log.Info().
		Str("port", a.cfg.Server.Port).
		Str("env", a.cfg.Env).
		Msg("Zeno Auth service starting")

	// TODO: Initialize database and HTTP server
	return nil
}