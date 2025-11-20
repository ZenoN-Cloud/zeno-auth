package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
)

type PasswordService interface {
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword, ipAddress, userAgent string) error
}

type UserHandler struct {
	userService     service.UserServiceInterface
	passwordService PasswordService
}

func NewUserHandler(userService service.UserServiceInterface, passwordService PasswordService) *UserHandler {
	return &UserHandler{
		userService:     userService,
		passwordService: passwordService,
	}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		IsActive: user.IsActive,
	})
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	if h.passwordService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Service unavailable"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.passwordService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully. All sessions have been logged out."})
}
