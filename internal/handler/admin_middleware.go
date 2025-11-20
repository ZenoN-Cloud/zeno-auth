package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware protects admin endpoints
// In production: use IP whitelist or basic auth
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		env := os.Getenv("ENV")
		
		// In development: allow all
		if env == "dev" || env == "development" {
			c.Next()
			return
		}

		// In production: check basic auth or IP whitelist
		// Option 1: Basic Auth
		username, password, hasAuth := c.Request.BasicAuth()
		if hasAuth {
			// Check credentials from env
			expectedUser := os.Getenv("ADMIN_USERNAME")
			expectedPass := os.Getenv("ADMIN_PASSWORD")
			
			if username == expectedUser && password == expectedPass && expectedUser != "" {
				c.Next()
				return
			}
		}

		// Option 2: IP Whitelist
		allowedIPs := os.Getenv("ADMIN_ALLOWED_IPS") // comma-separated
		if allowedIPs != "" {
			clientIP := c.ClientIP()
			ips := strings.Split(allowedIPs, ",")
			for _, ip := range ips {
				if strings.TrimSpace(ip) == clientIP {
					c.Next()
					return
				}
			}
		}

		// If no auth method configured in production, deny access
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "Access denied. Admin authentication required.",
		})
		c.Abort()
	}
}
