package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"

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
	cfg interface{},
	metricsCollector MetricsCollector,
) *gin.Engine {
	r := gin.New()
	r.Use(LoggingMiddleware())
	r.Use(SecurityHeadersMiddleware())

	// Metrics middleware
	if metricsCollector != nil {
		type metricsInterface interface {
			RecordRequestDuration(duration interface{})
		}
		if m, ok := interface{}(metricsCollector).(metricsInterface); ok {
			_ = m // Use metrics middleware if available
		}
	}

	// Extract CORS origins from config
	var corsOrigins []string
	if cfg != nil {
		type configWithCORS interface {
			GetCORSOrigins() []string
		}
		if c, ok := cfg.(configWithCORS); ok {
			corsOrigins = c.GetCORSOrigins()
		}
	}
	r.Use(CORSMiddleware(corsOrigins))
	r.Use(gin.Recovery())

	// Health endpoints
	r.GET("/health", Health)
	r.HEAD("/health", Health)

	// Enhanced health checks
	if db != nil {
		healthChecker := NewHealthChecker(db.Pool())
		r.GET("/health/ready", healthChecker.HealthReady)
		r.GET("/health/live", healthChecker.HealthLive)
	}

	// Debug and Metrics endpoints - protected in production
	r.GET("/debug", AdminAuthMiddleware(), Debug)

	// Metrics endpoint - protected in production
	r.GET("/metrics", AdminAuthMiddleware(), func(c *gin.Context) {
		if metricsCollector == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Metrics not available"})
			return
		}
		// Try to call GetMetricsInterface() method
		type metricsGetter interface {
			GetMetricsInterface() interface{}
		}
		if mg, ok := metricsCollector.(metricsGetter); ok {
			c.JSON(http.StatusOK, mg.GetMetricsInterface())
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Metrics not available"})
	})

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
		// EmailService is optional
		var emailService *service.EmailService
		if cfg != nil {
			type configWithEmail interface {
				GetEmailService() *service.EmailService
			}
			if c, ok := cfg.(configWithEmail); ok {
				emailService = c.GetEmailService()
			}
		}

		// PasswordResetService is optional
		var passwordResetSvc *service.PasswordResetService
		if cfg != nil {
			type configWithPasswordReset interface {
				GetPasswordResetService() *service.PasswordResetService
			}
			if c, ok := cfg.(configWithPasswordReset); ok {
				passwordResetSvc = c.GetPasswordResetService()
			}
		}

		authHandler := NewAuthHandler(authService, emailService, passwordResetSvc, auditService, metricsCollector)
		userHandler := NewUserHandler(userService, passwordService)
		jwksHandler := NewJWKSHandler(jwtManager)

		r.GET("/jwks", jwksHandler.GetJWKS)

		auth := r.Group("/auth")
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

		me := r.Group("/me", AuthMiddleware(jwtManager))
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
				gdprHandler := NewGDPRHandler(gdprService, auditService)
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

	// Admin endpoints - protected in production
	admin := r.Group("/admin", AdminAuthMiddleware())
	{
		// Compliance reports
		complianceHandler := NewComplianceHandler(nil)
		admin.GET("/compliance/report", complianceHandler.GetComplianceReport)
		admin.GET("/compliance/status", complianceHandler.GetComplianceStatus)
	}

	return r
}
