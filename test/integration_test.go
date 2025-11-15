package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZenoN-Cloud/zeno-auth/internal/handler"
	"github.com/stretchr/testify/assert"
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

	// Should fail without proper database setup
	assert.Contains(t, []int{http.StatusInternalServerError, http.StatusBadRequest}, w.Code)
}

func setupTestRouter(t *testing.T) http.Handler {
	// For basic tests, create minimal router
	return handler.SetupRouter(nil, nil, nil)
}