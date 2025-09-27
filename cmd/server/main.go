package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evolution-api/evolution-go/internal/config"
	"github.com/evolution-api/evolution-go/internal/repository"
	"github.com/evolution-api/evolution-go/internal/routes"
	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("Failed to load configuration", zap.Error(err))
	}

	zap.L().Info("Starting Evolution API Go", zap.String("version", "1.0.0"))

	// Connect to database
	db, err := sql.Open(cfg.Database.Provider, cfg.Database.URL)
	if err != nil {
		zap.L().Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		zap.L().Fatal("Failed to ping database", zap.Error(err))
	}

	zap.L().Info("Connected to database successfully")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		zap.L().Fatal("Failed to connect to Redis", zap.Error(err))
	}

	zap.L().Info("Connected to Redis successfully")

	// Create WhatsApp store container
	container, err := sqlstore.New(cfg.Database.Provider, cfg.Database.URL, waLog.Stdout("Database", "INFO", false))
	if err != nil {
		zap.L().Fatal("Failed to create WhatsApp store container", zap.Error(err))
	}

	// Create repositories
	instanceRepo := repository.NewInstanceRepository(db)

	// Create services
	instanceManager := services.NewInstanceManager(instanceRepo, container)
	webhookService := services.NewWebhookService(cfg.Webhook.Timeout, cfg.Webhook.RetryAttempts)
	cacheService := services.NewCacheService(redisClient)

	// Setup routes
	router := routes.SetupRoutes(instanceManager, webhookService, cacheService, db, redisClient, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		zap.L().Info("Starting HTTP server", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown instance manager
	if err := instanceManager.Shutdown(ctx); err != nil {
		zap.L().Error("Failed to shutdown instance manager", zap.Error(err))
	}

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}

	zap.L().Info("Server exited")
}
