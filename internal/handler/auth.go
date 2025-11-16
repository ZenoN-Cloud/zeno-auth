package handler

import (
	"net/http"

	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService service.AuthServiceInterface
}

func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		if err == service.ErrEmailExists {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Registration failed"})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		IsActive: user.IsActive,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		if err == service.ErrInvalidCredentials || err == service.ErrUserNotActive {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
			return
		}
		// Log the actual error for debugging
		c.Error(err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Login failed"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	accessToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userUUID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
