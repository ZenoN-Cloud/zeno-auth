package app

import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/bootstrap"
	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/handler"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
)

type App struct {
	container *bootstrap.Container
	server    *http.Server
}

func New() (*App, error) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Init logger
	if err := config.SetupLogger(cfg); err != nil {
		return nil, fmt.Errorf("failed to setup logger: %w", err)
	}

	// Gin mode
	switch cfg.Env {
	case "prod", "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	log.Info().Msg("Building application container...")

	// Build DI container
	container, err := bootstrap.BuildContainer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build container: %w", err)
	}

	// Setup main router
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
		container.EmailService,
		container.PasswordResetService,
	)
	if router == nil {
		return nil, fmt.Errorf("router setup failed: nil router returned")
	}

	// Setup internal router for billing integration
	orgRepoImpl := postgres.NewOrganizationRepo(container.DB)
	internalRouter := handler.SetupInternalRouter(orgRepoImpl, log.Logger)
	// Mount internal routes
	router.Any("/internal/*path", gin.WrapH(internalRouter))

	if container.Config.Server.Port == "" {
		return nil, fmt.Errorf("invalid configuration: PORT must not be empty")
	}

	// Create HTTP server
	server := &http.Server{
		Addr:              ":" + container.Config.Server.Port,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          stdLoggerAdapter(),
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

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("HTTP server shutdown error")
			return
		}

		log.Info().Msg("HTTP server gracefully stopped")
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

// stdLoggerAdapter returns a *log.Logger that writes into zerolog.
func stdLoggerAdapter() *stdlog.Logger {
	writer := log.Logger.With().Str("component", "http-server").Logger()
	return stdlog.New(writer, "", 0)
}
