package app

import (
	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *config.Config
	db  *postgres.DB
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if err := config.SetupLogger(cfg); err != nil {
		return nil, err
	}

	db, err := postgres.New(cfg.Database.URL)
	if err != nil {
		return nil, err
	}

	return &App{cfg: cfg, db: db}, nil
}

func (a *App) Run() error {
	log.Info().
		Str("port", a.cfg.Server.Port).
		Str("env", a.cfg.Env).
		Msg("Zeno Auth service starting")

	// TODO: Initialize HTTP server
	return nil
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}