package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/config"
	"github.com/ZenoN-Cloud/zeno-auth/internal/middleware"
	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/rs/zerolog/log"
)

type ConsentService interface {
	GrantConsent(ctx context.Context, userID uuid.UUID, consentType model.ConsentType, version string) error
	RevokeConsent(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) error
	GetUserConsents(ctx context.Context, userID uuid.UUID) ([]*model.UserConsent, error)
	HasConsent(ctx context.Context, userID uuid.UUID, consentType model.ConsentType) (bool, error)
}

func SetupRouter(
	authService service.AuthServiceInterface,
	userService service.UserServiceInterface,
	consentService ConsentService,
	auditService AuditService,
	cleanupService CleanupService,
	gdprService GDPRService,
	passwordService PasswordService,
	sessionService SessionService,
	jwtManager *token.JWTManager,
	db *postgres.DB,
	cfg *config.Config,
	metricsCollector MetricsCollector,
	emailService *service.EmailService,
	passwordResetService *service.PasswordResetService,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(LoggingMiddleware())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RequestSizeLimit(1 * 1024 * 1024)) // 1MB limit

	// ENV
	env := ""
	if cfg != nil {
		env = strings.ToLower(cfg.GetEnv())
	}

	// CORS
	var corsOrigins []string
	if cfg != nil {
		corsOrigins = cfg.GetCORSOrigins()
	}
	// Add localhost for development testing
	if env != "production" && env != "prod" {
		corsOrigins = append(corsOrigins, "http://localhost:3000", "http://localhost:8000", "http://127.0.0.1:3000", "http://127.0.0.1:8000")
	}
	r.Use(CORSMiddleware(corsOrigins))

	// Health endpoints
	r.GET("/health", Health)
	r.HEAD("/health", Health)

	// Enhanced health checks
	if db != nil {
		healthChecker := NewHealthChecker(db.Pool())
		r.GET("/health/ready", healthChecker.HealthReady)
		r.GET("/health/live", healthChecker.HealthLive)
	}

	// Debug endpoint - disabled in prod/prodution (и сам handler доп. проверяет ENV)
	if env != "production" && env != "prod" {
		r.GET("/debug", AdminAuthMiddleware(), Debug)
	}

	// Metrics handler
	metricsHandler := func(c *gin.Context) {
		if metricsCollector == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Metrics not available"})
			return
		}
		type metricsGetter interface {
			GetMetricsInterface() interface{}
		}
		if mg, ok := metricsCollector.(metricsGetter); ok {
			c.JSON(http.StatusOK, mg.GetMetricsInterface())
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Metrics not available"})
	}

	// Metrics endpoint - protected in prod, public in dev
	if env == "production" || env == "prod" {
		r.GET("/metrics", AdminAuthMiddleware(), metricsHandler)
	} else {
		r.GET("/metrics", metricsHandler)
	}

	// Cleanup endpoints (protected)
	if db != nil {
		cleanupHandler := NewCleanupHandler(db, cleanupService)
		r.POST("/debug/cleanup", AdminAuthMiddleware(), CSRFMiddleware(), AdminRateLimiter(), cleanupHandler.CleanupAll)
		if cleanupService != nil {
			r.POST("/debug/cleanup-expired", AdminAuthMiddleware(), CSRFMiddleware(), AdminRateLimiter(), cleanupHandler.CleanupExpired)
		}
	}

	// Only add other endpoints if services are available
	if authService != nil && userService != nil && jwtManager != nil {
		// PasswordResetService is now optional via parameter
		authHandler := NewAuthHandler(authService, emailService, passwordResetService, auditService, metricsCollector)
		userHandler := NewUserHandler(userService, passwordService)
		jwksHandler := NewJWKSHandler(jwtManager)

		// JWKS endpoint (no versioning for standards compliance)
		r.GET("/.well-known/jwks.json", jwksHandler.GetJWKS)
		r.GET("/jwks", jwksHandler.GetJWKS) // Legacy endpoint

		// API v1 routes
		v1 := r.Group("/v1")
		{
			auth := v1.Group("/auth")
			{
				auth.POST("/register", RegisterRateLimiter(), OriginCheckMiddleware(corsOrigins), CSRFMiddleware(), authHandler.Register)
				auth.POST("/login", LoginRateLimiter(), OriginCheckMiddleware(corsOrigins), CSRFMiddleware(), authHandler.Login)
				auth.POST("/refresh", RefreshRateLimiter(), OriginCheckMiddleware(corsOrigins), CSRFMiddleware(), authHandler.Refresh)
				auth.POST("/logout", AuthMiddleware(jwtManager), CSRFMiddleware(), authHandler.Logout)
				auth.POST("/verify-email", CSRFMiddleware(), authHandler.VerifyEmail)
				auth.POST("/resend-verification", AuthMiddleware(jwtManager), CSRFMiddleware(), authHandler.ResendVerification)
				auth.POST("/forgot-password", middleware.StrictRateLimit(), CSRFMiddleware(), authHandler.ForgotPassword)
				auth.POST("/reset-password", middleware.StrictRateLimit(), CSRFMiddleware(), authHandler.ResetPassword)
			}

			me := v1.Group("/me", AuthMiddleware(jwtManager))
			{
				me.GET("", userHandler.GetProfile)
				me.GET("/status", userHandler.GetProfile)
				if passwordService != nil {
					me.POST("/change-password", CSRFMiddleware(), userHandler.ChangePassword)
				}

				// Organizations
				if db != nil {
					orgRepo := postgres.NewOrganizationRepo(db)
					orgHandler := NewOrganizationHandlerWithRepo(orgRepo, log.Logger)
					me.GET("/organizations", orgHandler.GetUserOrganizations)
				}

				if consentService != nil {
					consentHandler := NewConsentHandler(consentService)
					me.GET("/consents", consentHandler.GetConsents)
					me.POST("/consents", CSRFMiddleware(), consentHandler.GrantConsent)
					me.DELETE("/consents/:type", CSRFMiddleware(), consentHandler.RevokeConsent)
				}

				if gdprService != nil {
					gdprHandler := NewGDPRHandler(gdprService, auditService, emailService)
					me.GET("/data-export", gdprHandler.ExportData)
					me.DELETE("/account", CSRFMiddleware(), gdprHandler.DeleteAccount)
				}

				// Session management
				if sessionService != nil {
					sessionHandler := NewSessionHandler(sessionService)
					me.GET("/sessions", sessionHandler.GetSessions)
					me.DELETE("/sessions/:id", CSRFMiddleware(), sessionHandler.RevokeSession)
					me.DELETE("/sessions", CSRFMiddleware(), sessionHandler.RevokeAllSessions)
				}
			}

			// Organizations at v1 level
			if db != nil {
				orgRepo := postgres.NewOrganizationRepo(db)
				orgHandler := NewOrganizationHandlerWithRepo(orgRepo, log.Logger)
				v1.GET("/organizations", AuthMiddleware(jwtManager), orgHandler.GetUserOrganizations)
				v1.GET("/status", AuthMiddleware(jwtManager), userHandler.GetProfile)
			}
		}

		// Legacy routes (without versioning) - for backward compatibility
		auth := r.Group("/auth")
		{
			auth.POST("/register", RegisterRateLimiter(), OriginCheckMiddleware(corsOrigins), CSRFMiddleware(), authHandler.Register)
			auth.POST("/login", LoginRateLimiter(), OriginCheckMiddleware(corsOrigins), CSRFMiddleware(), authHandler.Login)
			auth.POST("/refresh", RefreshRateLimiter(), OriginCheckMiddleware(corsOrigins), CSRFMiddleware(), authHandler.Refresh)
			auth.POST("/logout", AuthMiddleware(jwtManager), CSRFMiddleware(), authHandler.Logout)
		}

		me := r.Group("/me", AuthMiddleware(jwtManager))
		{
			me.GET("", userHandler.GetProfile)
		}

		// Legacy routes for frontend compatibility
		if db != nil {
			orgRepo := postgres.NewOrganizationRepo(db)
			orgHandler := NewOrganizationHandlerWithRepo(orgRepo, log.Logger)
			r.GET("/status", AuthMiddleware(jwtManager), userHandler.GetProfile)
			r.GET("/organizations", AuthMiddleware(jwtManager), orgHandler.GetUserOrganizations)
		}
	}

	// Admin endpoints - always protected, enabled in production
	if env == "production" || env == "prod" {
		_ = r.Group("/admin", AdminAuthMiddleware())
		// TODO: Implement ComplianceReporter interface for AuditService
		// admin := r.Group("/admin", AdminAuthMiddleware())
		// if auditService != nil {
		// 	complianceHandler := NewComplianceHandler(auditService)
		// 	admin.GET("/compliance/report", complianceHandler.GetComplianceReport)
		// 	admin.GET("/compliance/status", complianceHandler.GetComplianceStatus)
		// }
		_ = auditService // prevent unused variable error
	}

	return r
}
