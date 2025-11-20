package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/ZenoN-Cloud/zeno-auth/internal/metrics"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
)

func AuthMiddleware(jwtManager *token.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Bearer token required"})
			c.Abort()
			return
		}

		claims, err := jwtManager.Validate(c.Request.Context(), tokenString)
		if err != nil {
			log.Error().Err(err).Str("token_prefix", tokenString[:20]).Msg("Token validation failed")
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID.String())
		c.Set("org_id", claims.OrgID.String())
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.Info().
			Str("method", param.Method).
			Str("path", param.Path).
			Int("status", param.StatusCode).
			Dur("latency", param.Latency).
			Str("ip", param.ClientIP).
			Msg("HTTP request")
		return ""
	})
}

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else if len(allowedOrigins) == 0 {
			// Fallback for backward compatibility
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// MetricsMiddleware records request duration metrics
func MetricsMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		m.RecordRequestDuration(duration)
	}
}
