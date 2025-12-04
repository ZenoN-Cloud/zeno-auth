package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	LoginAttempts      string // e.g., "5-M" = 5 requests per minute
	RegisterAttempts   string // e.g., "10-H" = 10 requests per hour
	RefreshAttempts    string // e.g., "20-M" = 20 requests per minute
	GeneralAPIRequests string // e.g., "100-M" = 100 requests per minute
}

// DefaultRateLimitConfig returns secure default rate limits
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		LoginAttempts:      "5-M",   // 5 attempts per minute
		RegisterAttempts:   "10-H",  // 10 registrations per hour
		RefreshAttempts:    "20-M",  // 20 refresh requests per minute
		GeneralAPIRequests: "100-M", // 100 general API requests per minute
	}
}

// NewRateLimiter creates a new rate limiter with the given rate
func NewRateLimiter(rate string) gin.HandlerFunc {
	rateLimit, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		// Log the error for debugging
		log.Error().Err(err).Str("rate", rate).Msg("Invalid rate limit configuration, allowing all requests")
		// Return a middleware that always allows requests if rate limit config is invalid
		return func(c *gin.Context) {
			c.Header("X-RateLimit-Error", "Invalid rate limit configuration")
			c.Next()
		}
	}

	store := memory.NewStore()
	instance := limiter.New(store, rateLimit)

	middleware := mgin.NewMiddleware(instance, mgin.WithKeyGetter(func(c *gin.Context) string {
		// Use IP address as key for rate limiting
		return c.ClientIP()
	}))

	return func(c *gin.Context) {
		// Safely execute middleware with error recovery
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Msg("Rate limiter middleware panic")
				c.Header("X-RateLimit-Error", "Rate limiter error")
				c.Next()
			}
		}()

		middleware(c)

		// Check if rate limit was exceeded with proper error handling
		status := c.Writer.Status()
		if status == http.StatusTooManyRequests {
			// Ensure response hasn't been written yet
			if !c.Writer.Written() {
				c.JSON(http.StatusTooManyRequests, ErrorResponse{
					Error: "Rate limit exceeded. Please try again later.",
				})
			}
			c.Abort()
			return
		} else if status >= 400 && status < 600 {
			// Handle other error statuses from rate limiter
			log.Warn().Int("status", status).Msg("Rate limiter returned error status")
			if !c.Writer.Written() {
				c.Header("X-RateLimit-Error", "Rate limiter error")
			}
		}
	}
}

// LoginRateLimiter applies rate limiting to login endpoint
func LoginRateLimiter() gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	return NewRateLimiter(config.LoginAttempts)
}

// RegisterRateLimiter applies rate limiting to registration endpoint
func RegisterRateLimiter() gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	return NewRateLimiter(config.RegisterAttempts)
}

// RefreshRateLimiter applies rate limiting to token refresh endpoint
func RefreshRateLimiter() gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	return NewRateLimiter(config.RefreshAttempts)
}

// GeneralAPIRateLimiter applies rate limiting to general API endpoints
func GeneralAPIRateLimiter() gin.HandlerFunc {
	config := DefaultRateLimitConfig()
	return NewRateLimiter(config.GeneralAPIRequests)
}

// AdminRateLimiter applies strict rate limiting to admin operations
func AdminRateLimiter() gin.HandlerFunc {
	// Very strict: 2 requests per minute for destructive admin operations
	return NewRateLimiter("2-M")
}
