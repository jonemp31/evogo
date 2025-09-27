package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	WindowSize        time.Duration
}

// RateLimitMiddleware provides rate limiting functionality
func RateLimitMiddleware(cacheService *services.CacheService, config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address or API key)
		clientID := c.ClientIP()
		if apiKey := c.GetHeader("apikey"); apiKey != "" {
			clientID = fmt.Sprintf("api:%s", apiKey)
		}

		// Check rate limit
		allowed, err := checkRateLimit(c.Request.Context(), cacheService, clientID, config)
		if err != nil {
			zap.L().Error("Rate limit check failed", zap.Error(err))
			// Allow request if rate limit check fails
			c.Next()
			return
		}

		if !allowed {
			zap.L().Warn("Rate limit exceeded", zap.String("client", clientID))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per %v allowed", config.RequestsPerMinute, config.WindowSize),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// InstanceRateLimitMiddleware provides rate limiting per instance
func InstanceRateLimitMiddleware(cacheService *services.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		instanceName := c.Param("name")
		if instanceName == "" {
			c.Next()
			return
		}

		// Get client identifier
		clientID := c.ClientIP()
		if apiKey := c.GetHeader("apikey"); apiKey != "" {
			clientID = fmt.Sprintf("api:%s", apiKey)
		}

		// Create instance-specific rate limit key
		rateLimitKey := fmt.Sprintf("ratelimit:instance:%s:%s", instanceName, clientID)

		// Check rate limit (more restrictive for instance operations)
		config := RateLimitConfig{
			RequestsPerMinute: 30, // 30 requests per minute per instance
			WindowSize:        time.Minute,
		}

		allowed, err := checkRateLimit(c.Request.Context(), cacheService, rateLimitKey, config)
		if err != nil {
			zap.L().Error("Instance rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if !allowed {
			zap.L().Warn("Instance rate limit exceeded",
				zap.String("instance", instanceName),
				zap.String("client", clientID))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Instance rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per minute allowed for instance %s", config.RequestsPerMinute, instanceName),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// MessageRateLimitMiddleware provides rate limiting for message sending
func MessageRateLimitMiddleware(cacheService *services.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		instanceName := c.Param("name")
		if instanceName == "" {
			c.Next()
			return
		}

		// Get client identifier
		clientID := c.ClientIP()
		if apiKey := c.GetHeader("apikey"); apiKey != "" {
			clientID = fmt.Sprintf("api:%s", apiKey)
		}

		// Create message-specific rate limit key
		rateLimitKey := fmt.Sprintf("ratelimit:message:%s:%s", instanceName, clientID)

		// More restrictive rate limit for message sending
		config := RateLimitConfig{
			RequestsPerMinute: 60, // 60 messages per minute per instance
			WindowSize:        time.Minute,
		}

		allowed, err := checkRateLimit(c.Request.Context(), cacheService, rateLimitKey, config)
		if err != nil {
			zap.L().Error("Message rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if !allowed {
			zap.L().Warn("Message rate limit exceeded",
				zap.String("instance", instanceName),
				zap.String("client", clientID))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Message rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d messages per minute allowed for instance %s", config.RequestsPerMinute, instanceName),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit checks if a request is within rate limits using sliding window
func checkRateLimit(ctx context.Context, cacheService *services.CacheService, clientID string, config RateLimitConfig) (bool, error) {
	now := time.Now()
	windowStart := now.Truncate(config.WindowSize)

	// Create keys for current window and previous window
	currentWindowKey := fmt.Sprintf("%s:%d", clientID, windowStart.Unix())
	previousWindowKey := fmt.Sprintf("%s:%d", clientID, windowStart.Add(-config.WindowSize).Unix())

	// Get current window count
	currentCount, err := cacheService.GetCounter(ctx, currentWindowKey)
	if err != nil {
		return false, fmt.Errorf("failed to get current window count: %w", err)
	}

	// Get previous window count
	previousCount, err := cacheService.GetCounter(ctx, previousWindowKey)
	if err != nil {
		return false, fmt.Errorf("failed to get previous window count: %w", err)
	}

	// Calculate sliding window count
	// Weighted average of current and previous windows
	windowProgress := float64(now.Sub(windowStart)) / float64(config.WindowSize)
	slidingCount := float64(currentCount) + (float64(previousCount) * (1 - windowProgress))

	// Check if within limits
	if int(slidingCount) >= config.RequestsPerMinute {
		return false, nil
	}

	// Increment counter for current window
	_, err = cacheService.IncrementCounter(ctx, currentWindowKey)
	if err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}

	// Set expiration for current window key
	err = cacheService.SetCounter(ctx, currentWindowKey, currentCount+1, config.WindowSize*2)
	if err != nil {
		return false, fmt.Errorf("failed to set counter expiration: %w", err)
	}

	return true, nil
}

// GetRateLimitInfo returns current rate limit information for a client
func GetRateLimitInfo(ctx context.Context, cacheService *services.CacheService, clientID string) (map[string]interface{}, error) {
	now := time.Now()
	windowStart := now.Truncate(time.Minute)

	currentWindowKey := fmt.Sprintf("%s:%d", clientID, windowStart.Unix())
	previousWindowKey := fmt.Sprintf("%s:%d", clientID, windowStart.Add(-time.Minute).Unix())

	currentCount, err := cacheService.GetCounter(ctx, currentWindowKey)
	if err != nil {
		return nil, err
	}

	previousCount, err := cacheService.GetCounter(ctx, previousWindowKey)
	if err != nil {
		return nil, err
	}

	windowProgress := float64(now.Sub(windowStart)) / float64(time.Minute)
	slidingCount := float64(currentCount) + (float64(previousCount) * (1 - windowProgress))

	return map[string]interface{}{
		"current_window_count":  currentCount,
		"previous_window_count": previousCount,
		"sliding_window_count":  int(slidingCount),
		"window_progress":       windowProgress,
		"window_start":          windowStart,
		"window_end":            windowStart.Add(time.Minute),
	}, nil
}
