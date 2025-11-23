package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Debug(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"database_url": maskPassword(os.Getenv("DATABASE_URL")),
		"env":          os.Getenv("ENV"),
		"port":         os.Getenv("PORT"),
	})
}

func maskPassword(url string) string {
	if url == "" {
		return ""
	}
	// In production, completely redact sensitive info
	if os.Getenv("ENV") == "production" {
		return "[REDACTED]"
	}
	// In dev/staging, show only connection scheme
	if len(url) > 10 {
		return "postgres://***:***@***/***"
	}
	return "[REDACTED]"
}
