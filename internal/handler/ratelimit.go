package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		panic("invalid rate limit format: " + err.Error())
	}

	store := memory.NewStore()
	instance := limiter.New(store, rateLimit)

	middleware := mgin.NewMiddleware(instance, mgin.WithKeyGetter(func(c *gin.Context) string {
		// Use IP address as key for rate limiting
		return c.ClientIP()
	}))

	return func(c *gin.Context) {
		middleware(c)

		// Check if rate limit was exceeded
		if c.Writer.Status() == http.StatusTooManyRequests {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Error: "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
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
