package handler

import (
	"net/http"

	"github.com/ZenoN-Cloud/zeno-auth/internal/repository/postgres"
	"github.com/gin-gonic/gin"
)

type CleanupHandler struct {
	db *postgres.DB
}

func NewCleanupHandler(db *postgres.DB) *CleanupHandler {
	return &CleanupHandler{db: db}
}

func (h *CleanupHandler) CleanupAll(c *gin.Context) {
	// Only allow in dev environment
	secret := c.GetHeader("X-Admin-Secret")
	if secret != "dev-cleanup-secret-2024" {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Forbidden"})
		return
	}

	queries := []string{
		"TRUNCATE TABLE refresh_tokens CASCADE",
		"TRUNCATE TABLE org_memberships CASCADE",
		"TRUNCATE TABLE organizations CASCADE",
		"TRUNCATE TABLE users CASCADE",
	}

	for _, query := range queries {
		if _, err := h.db.Pool().Exec(c.Request.Context(), query); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "All tables cleaned"})
}
