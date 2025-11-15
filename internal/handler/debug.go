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
	// Simple masking - just show first 20 chars
	if len(url) > 50 {
		return url[:50] + "..."
	}
	return url
}
