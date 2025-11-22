package bootstrap

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/metrics"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
)

type Container struct {
	Config  *config.Config
	DB      *postgres.DB
	Metrics *metrics.Metrics

	JWTManager      *token.JWTManager
	RefreshManager  *token.RefreshManager
	PasswordManager *token.PasswordManager

	AuthService     service.AuthServiceInterface
	UserService     service.UserServiceInterface
	ConsentService  *service.ConsentService
	AuditService    *service.AuditService
	CleanupService  *service.CleanupService
	GDPRService     *service.GDPRService
	PasswordService *service.PasswordService
	SessionService  *service.SessionService
	EmailService    *service.EmailService
}

func BuildContainer(cfg *config.Config) (*Container, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	container := &Container{Config: cfg}

	log.Info().Msg("Connecting to database...")
	db, err := postgres.New(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	container.DB = db
	log.Info().Msg("Connected to PostgreSQL")

	log.Info().Msg("Initializing JWT manager...")
	jwtManager, err := token.NewJWTManager(cfg.JWT.PrivateKey, cfg.JWT.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}
	container.JWTManager = jwtManager
	log.Info().Msg("JWT manager initialized")

	log.Info().Msg("Initializing metrics collector...")
	container.Metrics = metrics.New()
	log.Info().Msg("Metrics collector initialized")

	container.RefreshManager = token.NewRefreshManager()
	container.PasswordManager = token.NewPasswordManager()

	userRepo := postgres.NewUserRepo(db)
	orgRepo := postgres.NewOrganizationRepo(db)
	membershipRepo := postgres.NewMembershipRepo(db)
	refreshRepo := postgres.NewRefreshTokenRepo(db)
	consentRepo := postgres.NewConsentRepository(db.Pool())
	auditRepo := postgres.NewAuditLogRepository(db.Pool())
	emailVerificationRepo := postgres.NewEmailVerificationRepository(db.Pool())

	serviceConfig := service.NewConfig(cfg)
	container.AuditService = service.NewAuditService(auditRepo)
	container.EmailService = service.NewEmailService(emailVerificationRepo, userRepo, container.AuditService)
	container.AuthService = service.NewAuthService(
		userRepo, orgRepo, membershipRepo, refreshRepo,
		jwtManager, container.RefreshManager, container.PasswordManager,
		container.EmailService, serviceConfig, db,
	)
	container.UserService = service.NewUserService(userRepo, membershipRepo)
	container.ConsentService = service.NewConsentService(consentRepo)
	container.CleanupService = service.NewCleanupService(refreshRepo, auditRepo)
	container.GDPRService = service.NewGDPRService(
		userRepo, orgRepo, membershipRepo, refreshRepo, consentRepo, auditRepo, db,
	)
	container.PasswordService = service.NewPasswordService(
		userRepo, refreshRepo, container.PasswordManager, container.AuditService, container.EmailService, db,
	)
	container.SessionService = service.NewSessionService(refreshRepo)

	log.Info().Msg("All services initialized")

	return container, nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
}
