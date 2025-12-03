package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Health endpoint returns OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

		Health(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
		assert.Contains(t, w.Body.String(), "zeno-auth")
	})
}

func TestHealthChecker_HealthLive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Liveness endpoint returns OK", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/health/live", nil)

		checker := NewHealthChecker(nil)
		checker.HealthLive(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "alive")
		assert.Contains(t, w.Body.String(), "uptime")
	})
}
