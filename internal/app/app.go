package app

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
)

type App struct {
	cfg             *config.Config
	db              *postgres.DB
	jwtManager      *token.JWTManager
	refreshManager  *token.RefreshManager
	passwordManager *token.PasswordManager
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

	jwtManager, err := token.NewJWTManager(cfg.JWT.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:             cfg,
		db:              db,
		jwtManager:      jwtManager,
		refreshManager:  token.NewRefreshManager(),
		passwordManager: token.NewPasswordManager(),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
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
