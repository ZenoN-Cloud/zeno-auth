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
	r.Use(SecurityHeadersMiddleware())

	// CORS
	var corsOrigins []string
	if cfg != nil {
		corsOrigins = cfg.GetCORSOrigins()
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

	// ENV
	env := ""
	if cfg != nil {
		env = strings.ToLower(cfg.GetEnv())
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
		r.POST("/debug/cleanup", AdminAuthMiddleware(), cleanupHandler.CleanupAll)
		if cleanupService != nil {
			r.POST("/debug/cleanup-expired", AdminAuthMiddleware(), cleanupHandler.CleanupExpired)
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
				auth.POST("/register", RegisterRateLimiter(), authHandler.Register)
				auth.POST("/login", LoginRateLimiter(), authHandler.Login)
				auth.POST("/refresh", RefreshRateLimiter(), authHandler.Refresh)
				auth.POST("/logout", AuthMiddleware(jwtManager), authHandler.Logout)
				auth.POST("/verify-email", authHandler.VerifyEmail)
				auth.POST("/resend-verification", AuthMiddleware(jwtManager), authHandler.ResendVerification)
				auth.POST("/forgot-password", authHandler.ForgotPassword)
				auth.POST("/reset-password", authHandler.ResetPassword)
			}

			me := v1.Group("/me", AuthMiddleware(jwtManager))
			{
				me.GET("", userHandler.GetProfile)
				if passwordService != nil {
					me.POST("/change-password", userHandler.ChangePassword)
				}

				if consentService != nil {
					consentHandler := NewConsentHandler(consentService)
					me.GET("/consents", consentHandler.GetConsents)
					me.POST("/consents", consentHandler.GrantConsent)
					me.DELETE("/consents/:type", consentHandler.RevokeConsent)
				}

				if gdprService != nil {
					gdprHandler := NewGDPRHandler(gdprService, auditService, emailService)
					me.GET("/data-export", gdprHandler.ExportData)
					me.DELETE("/account", gdprHandler.DeleteAccount)
				}

				// Session management
				if sessionService != nil {
					sessionHandler := NewSessionHandler(sessionService)
					me.GET("/sessions", sessionHandler.GetSessions)
					me.DELETE("/sessions/:id", sessionHandler.RevokeSession)
					me.DELETE("/sessions", sessionHandler.RevokeAllSessions)
				}
			}
		}

		// Legacy routes (without versioning) - for backward compatibility
		auth := r.Group("/auth")
		{
			auth.POST("/register", RegisterRateLimiter(), authHandler.Register)
			auth.POST("/login", LoginRateLimiter(), authHandler.Login)
			auth.POST("/refresh", RefreshRateLimiter(), authHandler.Refresh)
			auth.POST("/logout", AuthMiddleware(jwtManager), authHandler.Logout)
		}

		me := r.Group("/me", AuthMiddleware(jwtManager))
		{
			me.GET("", userHandler.GetProfile)
		}
	}

	// Admin endpoints - always protected, enabled in production
	if env == "production" || env == "prod" {
		admin := r.Group("/admin", AdminAuthMiddleware())
		{
			complianceHandler := NewComplianceHandler(nil)
			admin.GET("/compliance/report", complianceHandler.GetComplianceReport)
			admin.GET("/compliance/status", complianceHandler.GetComplianceStatus)
		}
	}

	return r
}
