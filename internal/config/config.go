package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	API      APIConfig
	Webhook  WebhookConfig
	Instance InstanceConfig
	Log      LogConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
	URL  string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL      string
	Provider string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// APIConfig holds API configuration
type APIConfig struct {
	Key       string
	DebugMode bool
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	Timeout       time.Duration
	RetryAttempts int
}

// InstanceConfig holds instance management configuration
type InstanceConfig struct {
	CleanupInterval time.Duration
	MaxIdleTime     time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional
		zap.L().Debug("No .env file found, using environment variables only")
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			URL:  getEnv("SERVER_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", "postgres://username:password@localhost:5432/evolution_api?sslmode=disable"),
			Provider: getEnv("DATABASE_PROVIDER", "postgresql"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		API: APIConfig{
			Key:       getEnv("API_KEY", ""),
			DebugMode: getEnvAsBool("DEBUG_MODE", false),
		},
		Webhook: WebhookConfig{
			Timeout:       time.Duration(getEnvAsInt("WEBHOOK_TIMEOUT", 30)) * time.Second,
			RetryAttempts: getEnvAsInt("WEBHOOK_RETRY_ATTEMPTS", 3),
		},
		Instance: InstanceConfig{
			CleanupInterval: time.Duration(getEnvAsInt("INSTANCE_CLEANUP_INTERVAL", 300)) * time.Second,
			MaxIdleTime:     time.Duration(getEnvAsInt("INSTANCE_MAX_IDLE_TIME", 3600)) * time.Second,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// Validate required configurations
	if config.API.Key == "" {
		zap.L().Warn("API_KEY is not set, authentication will be disabled")
	}

	return config, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
