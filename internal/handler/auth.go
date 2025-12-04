package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ZenoN-Cloud/zeno-auth/internal/errors"
	"github.com/ZenoN-Cloud/zeno-auth/internal/response"
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/ZenoN-Cloud/zeno-auth/internal/validator"
)

type AuthHandler struct {
	authService      service.AuthServiceInterface
	emailService     *service.EmailService
	passwordResetSvc *service.PasswordResetService
	auditService     AuditService
	metrics          MetricsCollector
}

func NewAuthHandler(authService service.AuthServiceInterface, emailService *service.EmailService, passwordResetSvc *service.PasswordResetService, auditService AuditService, metrics MetricsCollector) *AuthHandler {
	return &AuthHandler{
		authService:      authService,
		emailService:     emailService,
		passwordResetSvc: passwordResetSvc,
		auditService:     auditService,
		metrics:          metrics,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Validate and sanitize input
	inputValidator := validator.NewInputValidator()
	if err := inputValidator.ValidateEmail(req.Email); err != nil {
		httpErr := errors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}
	if err := inputValidator.ValidateName(req.FullName); err != nil {
		httpErr := errors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}
	if err := inputValidator.ValidateName(req.OrganizationName); err != nil {
		httpErr := errors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}

	req.Email = inputValidator.SanitizeEmail(req.Email)
	req.FullName = inputValidator.SanitizeName(req.FullName)
	req.OrganizationName = inputValidator.SanitizeName(req.OrganizationName)

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Password, req.FullName, req.OrganizationName)
	if err != nil {
		httpErr := errors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}

	// Audit log
	if h.auditService != nil {
		_ = h.auditService.Log(c.Request.Context(), &user.ID, "user_registered", map[string]interface{}{"email": user.Email}, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	// Metrics
	if h.metrics != nil {
		h.metrics.IncrementRegistrations()
	}

	response.Success(c, http.StatusCreated, UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		IsActive: user.IsActive,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		// Audit log failed login
		if h.auditService != nil {
			_ = h.auditService.Log(c.Request.Context(), nil, "login_failed", map[string]interface{}{"email": req.Email}, ipAddress, userAgent)
		}
		// Metrics
		if h.metrics != nil {
			h.metrics.IncrementLoginFailures()
		}
		httpErr := errors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}

	// Audit log successful login
	if h.auditService != nil {
		_ = h.auditService.Log(c.Request.Context(), nil, "user_logged_in", map[string]interface{}{"email": req.Email}, ipAddress, userAgent)
	}

	// Metrics
	if h.metrics != nil {
		h.metrics.IncrementLogins()
	}

	response.Success(c, http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	accessToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, userAgent, ipAddress)
	if err != nil {
		httpErr := errors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}

	// Metrics
	if h.metrics != nil {
		h.metrics.IncrementTokenRefreshes()
	}

	response.Success(c, http.StatusOK, gin.H{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	userIDString, ok := userID.(string)
	if !ok {
		response.BadRequest(c, "Invalid user ID format")
		return
	}

	userUUID, err := uuid.Parse(userIDString)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userUUID); err != nil {
		httpErr := errors.MapErrorToHTTP(err)
		response.Error(c, httpErr.StatusCode, httpErr.Code, httpErr.Message)
		return
	}

	// Audit log
	if h.auditService != nil {
		_ = h.auditService.Log(c.Request.Context(), &userUUID, "user_logged_out", nil, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	response.Success(c, http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	if h.emailService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Service unavailable"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.emailService.VerifyEmail(c.Request.Context(), req.Token, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	userIDString, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID format"})
		return
	}

	userUUID, err := uuid.Parse(userIDString)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	if h.emailService == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Service unavailable"})
		return
	}

	_, err = h.emailService.ResendVerification(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to resend verification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent",
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	if h.passwordResetSvc == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Service unavailable"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	_, err := h.passwordResetSvc.RequestPasswordReset(c.Request.Context(), req.Email, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request format"})
		return
	}

	if h.passwordResetSvc == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "Service unavailable"})
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := h.passwordResetSvc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword, ipAddress, userAgent); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid or expired reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully. Please login with your new password."})
}
