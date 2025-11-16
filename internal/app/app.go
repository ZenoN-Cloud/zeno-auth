package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/handler"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
)

type App struct {
	cfg    *config.Config
	db     *postgres.DB
	server *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.SetupLogger(cfg); err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}

	// Start with minimal setup - just health endpoint
	log.Info().Msg("Starting with minimal configuration...")

	// Try to connect to database but don't fail if it's not available
	var db *postgres.DB
	log.Info().Msg("Attempting to connect to database...")
	db, err = postgres.New(cfg.Database.URL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database - starting without DB")
		db = nil
	} else {
		log.Info().Msg("Connected to PostgreSQL")
	}

	log.Info().Msg("Initializing JWT manager...")
	jwtManager, err := token.NewJWTManager(cfg.JWT.PrivateKey, cfg.JWT.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}
	log.Info().Msg("JWT manager initialized")

	// Initialize services only if database is available
	var router *gin.Engine
	if db != nil {
		// Initialize repositories
		userRepo := postgres.NewUserRepo(db)
		orgRepo := postgres.NewOrganizationRepo(db)
		membershipRepo := postgres.NewMembershipRepo(db)
		refreshRepo := postgres.NewRefreshTokenRepo(db)

		// Initialize services
		refreshManager := token.NewRefreshManager()
		passwordManager := token.NewPasswordManager()
		serviceConfig := service.NewConfig(cfg)
		authService := service.NewAuthService(
			userRepo, orgRepo, membershipRepo, refreshRepo, jwtManager, refreshManager, passwordManager, serviceConfig,
		)
		userService := service.NewUserService(userRepo, membershipRepo)

		// Setup router with full services
		router = handler.SetupRouter(authService, userService, jwtManager, db)
	} else {
		// Setup minimal router with just health endpoint
		router = handler.SetupRouter(nil, nil, jwtManager, nil)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	return &App{
		cfg:    cfg,
		db:     db,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log.Info().
		Str("port", a.cfg.Server.Port).
		Str("env", a.cfg.Env).
		Msg("Zeno Auth service starting")

	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("HTTP server shutdown error")
		}
	}()

	log.Info().Str("addr", a.server.Addr).Msg("HTTP server listening")
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}
