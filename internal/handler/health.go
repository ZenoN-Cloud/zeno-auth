package handler

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var startTime = time.Now()

type HealthChecker struct {
	db *pgxpool.Pool
}

func NewHealthChecker(db *pgxpool.Pool) *HealthChecker {
	return &HealthChecker{db: db}
}

// Health - простая проверка (всегда возвращает 200)
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "zeno-auth",
	})
}

// HealthReady - readiness probe (проверяет зависимости)
func (h *HealthChecker) HealthReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	response := gin.H{
		"status":    "ready",
		"service":   "zeno-auth",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    gin.H{},
	}

	allHealthy := true
	checks := response["checks"].(gin.H)

	// Database check
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = gin.H{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			allHealthy = false
		} else {
			checks["database"] = gin.H{
				"status": "healthy",
			}
		}
	} else {
		checks["database"] = gin.H{
			"status": "not_configured",
		}
	}

	if !allHealthy {
		response["status"] = "not_ready"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthLive - liveness probe (проверяет что процесс жив)
func (h *HealthChecker) HealthLive(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime)

	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"service":   "zeno-auth",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    uptime.String(),
		"system": gin.H{
			"goroutines":   runtime.NumGoroutine(),
			"memory_alloc": m.Alloc / 1024 / 1024,      // MB
			"memory_total": m.TotalAlloc / 1024 / 1024, // MB
			"memory_sys":   m.Sys / 1024 / 1024,        // MB
			"gc_runs":      m.NumGC,
			"num_cpu":      runtime.NumCPU(),
		},
	})
}
