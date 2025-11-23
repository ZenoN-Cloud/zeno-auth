package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/bootstrap"
	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/handler"
)

type App struct {
	container *bootstrap.Container
	server    *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.SetupLogger(cfg); err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}

	// Set Gin mode based on environment
	switch cfg.Env {
	case "prod", "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	log.Info().Msg("Building application container...")
	container, err := bootstrap.BuildContainer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build container: %w", err)
	}

	router := handler.SetupRouter(
		container.AuthService,
		container.UserService,
		container.ConsentService,
		container.AuditService,
		container.CleanupService,
		container.GDPRService,
		container.PasswordService,
		container.SessionService,
		container.JWTManager,
		container.DB,
		container.Config,
		container.Metrics,
	)

	server := &http.Server{
		Addr:              ":" + container.Config.Server.Port,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		container: container,
		server:    server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log.Info().
		Str("port", a.container.Config.Server.Port).
		Str("env", a.container.Config.Env).
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
	if a.container != nil {
		a.container.Close()
	}
}
