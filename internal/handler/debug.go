package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Debug is disabled completely in production.
// In dev/staging it returns limited debugging info.
func Debug(c *gin.Context) {
	env := strings.ToLower(os.Getenv("ENV"))

	// Disable endpoint in production entirely
	if env == "production" || env == "prod" {
		c.JSON(
			http.StatusNotFound, gin.H{
				"error": "debug endpoint disabled in production",
			},
		)
		return
	}

	c.JSON(
		http.StatusOK, gin.H{
			"env":          env,
			"port":         os.Getenv("PORT"),
			"database_url": maskPassword(os.Getenv("DATABASE_URL")),
			"service":      "zeno-auth",
		},
	)
}

// maskPassword hides all sensitive info.
// Even in dev/staging it never exposes credentials.
func maskPassword(url string) string {
	if url == "" {
		return "[REDACTED]"
	}

	// Full redact in production (already handled in Debug(), but safe anyway)
	env := os.Getenv("ENV")
	if env == "" || strings.ToLower(env) == "production" {
		return "[REDACTED]"
	}

	// Always hide actual credentials
	// Example:
	// postgres://user:password@host:5432/dbname?sslmode=disable
	// becomes:
	// postgres://***:***@host/dbname
	parts := strings.Split(url, "@")
	if len(parts) < 2 {
		return "[REDACTED]"
	}

	// mask left part (credentials)
	creds := parts[0]
	right := parts[len(parts)-1] // Use last part in case of multiple @

	// remove scheme
	splitScheme := strings.SplitN(creds, "://", 2)
	if len(splitScheme) < 2 {
		return "[REDACTED]"
	}

	scheme := splitScheme[0]

	// right side still contains host/db
	return scheme + "://***:***@" + sanitizeRightSide(right)
}

// sanitizeRightSide strips query params for safety
func sanitizeRightSide(s string) string {
	if strings.Contains(s, "?") {
		s = strings.SplitN(s, "?", 2)[0]
	}
	return s
}
