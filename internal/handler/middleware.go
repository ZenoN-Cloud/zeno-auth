package handler

import (
	"net/http"
	"net/url"
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
			// Safe token prefix logging
			tokenPrefix := tokenString
			if len(tokenString) > 20 {
				tokenPrefix = tokenString[:20]
			}
			log.Error().Err(err).Str("token_prefix", tokenPrefix).Msg("Token validation failed")
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
		// HSTS - Force HTTPS (EU requirement)
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS Protection (legacy browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// CSP - Strict policy
		c.Header("Content-Security-Policy", "default-src 'none'; script-src 'self'; connect-src 'self'; img-src 'self'; style-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy (EU privacy requirement)
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		// Cache control for sensitive data
		if c.Request.URL.Path != "/health" && c.Request.URL.Path != "/jwks" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
		}

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
			// Validate origin URL to prevent XSS
			if _, err := url.Parse(origin); err == nil && !strings.ContainsAny(origin, "<>\"'") {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
			}
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

		// Safe metrics recording with error handling
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Dur("duration", duration).Msg("Metrics recording panic")
			}
		}()

		if m != nil {
			m.RecordRequestDuration(duration)
		} else {
			log.Warn().Dur("duration", duration).Msg("Metrics collector is nil, skipping recording")
		}
	}
}

// OriginCheckMiddleware validates Origin/Referer for CSRF protection on public endpoints
func OriginCheckMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET, HEAD, OPTIONS
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
		}

		if origin != "" {
			originURL, err := url.Parse(origin)
			if err != nil {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: "Invalid origin"})
				c.Abort()
				return
			}

			// Check if origin is in allowed list
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				allowedURL, err := url.Parse(allowedOrigin)
				if err == nil && originURL.Host == allowedURL.Host {
					allowed = true
					break
				}
			}

			if !allowed {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: "Origin not allowed"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// CSRFMiddleware provides CSRF protection by validating X-CSRF-Token header and Origin
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "CSRF token required"})
			c.Abort()
			return
		}

		// Validate Origin or Referer for state-changing requests
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
		}
		if origin != "" {
			originURL, err := url.Parse(origin)
			if err != nil || originURL.Host != c.Request.Host {
				c.JSON(http.StatusForbidden, ErrorResponse{Error: "Invalid origin"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
