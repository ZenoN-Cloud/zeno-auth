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
	"github.com/ZenoN-Cloud/zeno-auth/internal/metrics"
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

	// Connect to database
	log.Info().Msg("Connecting to database...")
	db, err := postgres.New(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Info().Msg("Connected to PostgreSQL")

	log.Info().Msg("Initializing JWT manager...")
	jwtManager, err := token.NewJWTManager(cfg.JWT.PrivateKey, cfg.JWT.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}
	log.Info().Msg("JWT manager initialized")

	// Initialize metrics
	log.Info().Msg("Initializing metrics collector...")
	metricsCollector := metrics.New()
	log.Info().Msg("Metrics collector initialized")

	// Initialize services only if database is available
	var router *gin.Engine
	if db != nil {
		// Initialize repositories
		userRepo := postgres.NewUserRepo(db)
		orgRepo := postgres.NewOrganizationRepo(db)
		membershipRepo := postgres.NewMembershipRepo(db)
		refreshRepo := postgres.NewRefreshTokenRepo(db)
		consentRepo := postgres.NewConsentRepository(db.Pool())
		auditRepo := postgres.NewAuditLogRepository(db.Pool())
		emailVerificationRepo := postgres.NewEmailVerificationRepository(db.Pool())

		// Initialize services
		refreshManager := token.NewRefreshManager()
		passwordManager := token.NewPasswordManager()
		serviceConfig := service.NewConfig(cfg)
		auditService := service.NewAuditService(auditRepo)
		emailService := service.NewEmailService(emailVerificationRepo, userRepo, auditService)
		authService := service.NewAuthService(
			userRepo, orgRepo, membershipRepo, refreshRepo, jwtManager, refreshManager, passwordManager, emailService,
			serviceConfig,
		)
		userService := service.NewUserService(userRepo, membershipRepo)
		consentService := service.NewConsentService(consentRepo)
		cleanupService := service.NewCleanupService(refreshRepo, auditRepo)
		gdprService := service.NewGDPRService(userRepo, orgRepo, membershipRepo, refreshRepo, consentRepo, auditRepo)
		passwordService := service.NewPasswordService(
			userRepo, refreshRepo, passwordManager, auditService, emailService,
		)
		sessionService := service.NewSessionService(refreshRepo)

		// Setup router with full services
		router = handler.SetupRouter(
			authService, userService, consentService, auditService, cleanupService, gdprService, passwordService,
			sessionService, jwtManager, db, cfg, metricsCollector,
		)
	} else {
		// Setup minimal router with just health endpoint
		router = handler.SetupRouter(nil, nil, nil, nil, nil, nil, nil, nil, jwtManager, nil, cfg, metricsCollector)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
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
