package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// HealthCheck performs a basic health check
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "evolution-api-go",
		"version": "1.0.0",
		"time":    time.Now().UTC(),
	})
}

// ReadinessCheck performs a readiness check (includes dependencies)
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	checks := make(map[string]interface{})
	allHealthy := true

	// Check database
	dbStatus := h.checkDatabase()
	checks["database"] = dbStatus
	if !dbStatus["healthy"].(bool) {
		allHealthy = false
	}

	// Check Redis
	redisStatus := h.checkRedis()
	checks["redis"] = redisStatus
	if !redisStatus["healthy"].(bool) {
		allHealthy = false
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":  map[string]bool{"ok": allHealthy},
		"checks":  checks,
		"service": "evolution-api-go",
		"version": "1.0.0",
		"time":    time.Now().UTC(),
	})
}

// LivenessCheck performs a liveness check
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "evolution-api-go",
		"time":    time.Now().UTC(),
	})
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.db.PingContext(ctx)
	if err != nil {
		zap.L().Error("Database health check failed", zap.Error(err))
		return map[string]interface{}{
			"healthy": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"healthy": true,
		"status":  "connected",
	}
}

// checkRedis checks Redis connectivity
func (h *HealthHandler) checkRedis() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.redis.Ping(ctx).Err()
	if err != nil {
		zap.L().Error("Redis health check failed", zap.Error(err))
		return map[string]interface{}{
			"healthy": false,
			"error":   err.Error(),
		}
	}

	return map[string]interface{}{
		"healthy": true,
		"status":  "connected",
	}
}
