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

// Health - basic public health check (always 200)
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "zeno-auth",
	})
}

// HealthReady - readiness probe (checks external dependencies)
func (h *HealthChecker) HealthReady(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := gin.H{}
	allHealthy := true

	// Database check
	if h.db == nil {
		checks["database"] = gin.H{"status": "not_configured"}
		allHealthy = false
	} else {
		start := time.Now()
		err := h.db.Ping(ctx)
		latency := time.Since(start)

		if err != nil {
			checks["database"] = gin.H{
				"status":  "fail",
				"latency": latency.Milliseconds(),
			}
			allHealthy = false
		} else {
			checks["database"] = gin.H{
				"status":  "ok",
				"latency": latency.Milliseconds(),
			}
		}
	}

	response := gin.H{
		"service":   "zeno-auth",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    checks,
	}

	if !allHealthy {
		response["status"] = "not_ready"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response["status"] = "ready"
	c.JSON(http.StatusOK, response)
}

// HealthLive - liveness probe (checks process health)
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
			"memory_alloc": m.Alloc / 1024 / 1024,
			"memory_total": m.TotalAlloc / 1024 / 1024,
			"memory_sys":   m.Sys / 1024 / 1024,
			"gc_runs":      m.NumGC,
			"num_cpu":      runtime.NumCPU(),
		},
	})
}
