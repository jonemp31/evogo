package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CacheService handles caching operations
type CacheService struct {
	redis *redis.Client
}

// NewCacheService creates a new cache service
func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{
		redis: redis,
	}
}

// Set stores a value in cache with expiration
func (cs *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = cs.redis.Set(ctx, key, jsonData, expiration).Err()
	if err != nil {
		zap.L().Error("Failed to set cache value", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (cs *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := cs.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found")
		}
		zap.L().Error("Failed to get cache value", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (cs *CacheService) Delete(ctx context.Context, key string) error {
	err := cs.redis.Del(ctx, key).Err()
	if err != nil {
		zap.L().Error("Failed to delete cache value", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (cs *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := cs.redis.Exists(ctx, key).Result()
	if err != nil {
		zap.L().Error("Failed to check cache key existence", zap.String("key", key), zap.Error(err))
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	return count > 0, nil
}

// SetInstanceCache stores instance data in cache
func (cs *CacheService) SetInstanceCache(ctx context.Context, instanceID string, data interface{}) error {
	key := fmt.Sprintf("instance:%s", instanceID)
	return cs.Set(ctx, key, data, 24*time.Hour) // Cache for 24 hours
}

// GetInstanceCache retrieves instance data from cache
func (cs *CacheService) GetInstanceCache(ctx context.Context, instanceID string, dest interface{}) error {
	key := fmt.Sprintf("instance:%s", instanceID)
	return cs.Get(ctx, key, dest)
}

// SetMessageCache stores message data in cache
func (cs *CacheService) SetMessageCache(ctx context.Context, messageID string, data interface{}) error {
	key := fmt.Sprintf("message:%s", messageID)
	return cs.Set(ctx, key, data, 1*time.Hour) // Cache for 1 hour
}

// GetMessageCache retrieves message data from cache
func (cs *CacheService) GetMessageCache(ctx context.Context, messageID string, dest interface{}) error {
	key := fmt.Sprintf("message:%s", messageID)
	return cs.Get(ctx, key, dest)
}

// SetQRCodeCache stores QR code in cache
func (cs *CacheService) SetQRCodeCache(ctx context.Context, instanceID, qrCode string) error {
	key := fmt.Sprintf("qrcode:%s", instanceID)
	return cs.Set(ctx, key, qrCode, 2*time.Minute) // QR codes expire in 2 minutes
}

// GetQRCodeCache retrieves QR code from cache
func (cs *CacheService) GetQRCodeCache(ctx context.Context, instanceID string) (string, error) {
	key := fmt.Sprintf("qrcode:%s", instanceID)
	var qrCode string
	err := cs.Get(ctx, key, &qrCode)
	if err != nil {
		return "", err
	}
	return qrCode, nil
}

// ClearInstanceCache clears all cache for an instance
func (cs *CacheService) ClearInstanceCache(ctx context.Context, instanceID string) error {
	pattern := fmt.Sprintf("instance:%s*", instanceID)
	keys, err := cs.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	if len(keys) > 0 {
		err = cs.redis.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	zap.L().Info("Cleared instance cache", zap.String("instance", instanceID), zap.Int("keys", len(keys)))
	return nil
}

// IncrementCounter increments a counter in cache
func (cs *CacheService) IncrementCounter(ctx context.Context, key string) (int64, error) {
	val, err := cs.redis.Incr(ctx, key).Result()
	if err != nil {
		zap.L().Error("Failed to increment counter", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}
	return val, nil
}

// SetCounter sets a counter value in cache
func (cs *CacheService) SetCounter(ctx context.Context, key string, value int64, expiration time.Duration) error {
	err := cs.redis.Set(ctx, key, value, expiration).Err()
	if err != nil {
		zap.L().Error("Failed to set counter", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set counter: %w", err)
	}
	return nil
}

// GetCounter gets a counter value from cache
func (cs *CacheService) GetCounter(ctx context.Context, key string) (int64, error) {
	val, err := cs.redis.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		zap.L().Error("Failed to get counter", zap.String("key", key), zap.Error(err))
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}
	return val, nil
}
