package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZenoN-Cloud/zeno-auth/internal/handler"
	"github.com/ZenoN-Cloud/zeno-auth/internal/model"
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter(t)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "zeno-auth", response["service"])
}

func TestRegisterEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	router := setupTestRouter(t)

	registerReq := map[string]string{
		"email":     "test@example.com",
		"password":  "password123",
		"full_name": "Test User",
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return conflict since mock returns ErrEmailExists
	assert.Equal(t, http.StatusConflict, w.Code)
}

// Mock services for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, email, password, fullName string) (*model.User, error) {
	args := m.Called(ctx, email, password, fullName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, error) {
	args := m.Called(ctx, email, password, userAgent, ipAddress)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func setupTestRouter(t *testing.T) http.Handler {
	// Create mock services
	authService := &MockAuthService{}
	userService := &MockUserService{}
	
	// Create a minimal JWT manager for testing
	jwtManager, err := token.NewJWTManager(
		"-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCtURS5/MhZGHfr\naQRZebc/veam5QxMLNPijJUYai2aQJQwbpNHcvbb3xc0P6kpnXYKvco7/uxuxq2p\nxeTIBEjBnmG0tmxvGGLHeyYbe3xkUZBKoNikzj03jnNN5xPi303xe0En+CvhQFkR\nyxE7qQd3X2SoePLngcb3KleO9EwIFoNDhplDhQQQHxPuYt9hCJy65v/uoXAiRlJE\nkAZ2f8klJ7pLTfkbYntOP6gJwiEVBPmXy5rviUsPj3meMYBsJNNPrMWJm4cUA2ky\n3s2JbNCmnXeA3DJUngUepCklsyqlhkyMNeWzK5c9L5s8Y3o+LmCoy3AHkHtrJrZM\n+kTwEX3XAgMBAAECggEAQicFALpd5DflKcrzOI2vJpq+q2wYhgjENSAQlnmMf2hv\nx46lE2vrkl+z9SLpV/N8hzwKsVyrdNrLlVXt7XRJKvHffED24W6O4XH9SRcYkxfY\nuctr9XeswQRTuWPeYMV39Bhl9bIRWZAcjyCRqtJpAaS9AFrt5/ROc6/LLMrNLHZ9\nZqpOPa5MkSYX0KqzCIGifyIXIPhVRqmeNYYHIyNN874Ix1NGsc6E2H3uIByI/IDw\nRypE0S9AbGi40lPO0NCEYCsIaRS7Onb3K0Zzm0w0NvMlLLzkruENfLMP+4eTUbWs\nMHJI1/bRFo1HFHbOn0wFQvteeYUauzJwEI2Tr9WaDQKBgQDtDaHK6vqizUGH5vrl\n96q6pGHsMHvlR15/s0dKvufTTyLLKS25k3XHTE/ChMKdlHvqFGqBCfaDiEMh1Y5T\nET4HlktmtrRHLx1sgiLUmwlxzRJlGiuOJy4EHubjRbMjP5S0QtymW+jVBZx+SUJI\n7yWD05qKilO48cvZDhO/EAMyfQKBgQC7K1SBtpyVc0lIBqn9/o1VV1yqTPKzwW3O\n5vtThw8j4jqI6WsgIHtMRRXIKFZVV0hbo2LHnDmhHlDRZQXXoiCR/J0tbt53MR/0\n4Rq12CgW7aZF0slfbBkotF69MWlYFfWgzM+BElu5YE6/iSwkZbt410B3z9c9g95j\nxxa+T9bt4wKBgQDNwhheXmmgqBKqWMYMmFW73XUVotvXnoQayc0mxt/IXZcwypRi\n0OjZTZapm7ylNL395yyuxqwPbVX/5zK7TWsPANh/1jRS2UVr6uU6rzuaaMr/sKB/\nqehaMUxtlxEvlj+H28VULNDDHjTAtOvxDIr+isxIVlrnXBF5XKutGsP7rQKBgHGo\nVU/TiXCDqotvaIkRq9eYDnBn+7XGjxzmTNYjHMGInk0HmYLP1q+xABIk1JBMSWdE\nZzaZmrFJTIBrXUndbPPZt8SgH723ehVlIKguU+HgfGjIIHqulPSP2zv+Jl9ULm1w\nEc3qTQLcBdXvwXt0v4wZAk//SVBUpJZojloQ945LAoGAZsGEep26WBnw9sFKkxfC\nALT6Dz+TEfkPRrKsN10MVAl5JODhM34lgIIkadsmplLSxMzRSuTvZ+/GjBwOM1yy\nyKfRHYBmjuIuzSDRbGIIkY8bAh59a6gETdMmPL6q4RBfUxhgRgu3axt/I51HFqL0\nE6+XJ0hEgQ4TzMRglfhSUNk=\n-----END PRIVATE KEY-----",
		"", // public key will be derived
	)
	require.NoError(t, err)
	
	// Setup mock expectations for register test
	authService.On("Register", mock.Anything, "test@example.com", "password123", "Test User").Return(nil, service.ErrEmailExists)
	
	return handler.SetupRouter(authService, userService, jwtManager)
}