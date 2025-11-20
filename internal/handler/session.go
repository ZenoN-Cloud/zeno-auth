package handler

import (
	"context"
	"net/http"

	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SessionService interface {
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*model.RefreshToken, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
}

type SessionHandler struct {
	sessionService SessionService
}

func NewSessionHandler(sessionService SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

func (h *SessionHandler) GetSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	sessions, err := h.sessionService.GetActiveSessions(c.Request.Context(), uid)
	if err != nil {
		log.Error().Err(err).Str("user_id", uid.String()).Msg("Failed to get sessions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *SessionHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("id")
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	if err := h.sessionService.RevokeSession(c.Request.Context(), sid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked"})
}

func (h *SessionHandler) RevokeAllSessions(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.sessionService.RevokeAllSessions(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All sessions revoked"})
}
