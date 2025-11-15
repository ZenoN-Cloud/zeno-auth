package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_AuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		t.Skip("E2E_BASE_URL not set, skipping E2E tests")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)
		assert.Equal(t, "healthy", health["status"])
	})

	// Test JWKS endpoint
	t.Run("JWKS Endpoint", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/jwks")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var jwks map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&jwks)
		require.NoError(t, err)
		assert.Contains(t, jwks, "keys")
	})

	// Test registration flow
	t.Run("Registration Flow", func(t *testing.T) {
		registerReq := map[string]string{
			"email":     "e2e-test@example.com",
			"password":  "testpassword123",
			"full_name": "E2E Test User",
		}

		body, _ := json.Marshal(registerReq)
		resp, err := client.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Registration might fail if user exists, that's OK for E2E
		assert.Contains(t, []int{http.StatusCreated, http.StatusConflict}, resp.StatusCode)
	})
}