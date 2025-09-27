package routes

import (
	"database/sql"
	"time"

	"github.com/evolution-api/evolution-go/internal/config"
	"github.com/evolution-api/evolution-go/internal/handlers"
	"github.com/evolution-api/evolution-go/internal/middleware"
	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	instanceManager *services.InstanceManager,
	webhookService *services.WebhookService,
	cacheService *services.CacheService,
	db *sql.DB,
	redis *redis.Client,
	cfg *config.Config,
) *gin.Engine {
	// Create Gin engine
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.Use(middleware.AuthMiddleware(cfg))
	r.Use(middleware.RateLimitMiddleware(cacheService, middleware.RateLimitConfig{
		RequestsPerMinute: 100,
		WindowSize:        time.Minute,
	}))

	// Health check endpoints
	healthHandler := handlers.NewHealthHandler(db, redis)
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)
	r.GET("/live", healthHandler.LivenessCheck)

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := r.Group("/api")
	{
		// Instance routes
		instances := api.Group("/instance")
		{
			instanceHandler := handlers.NewInstanceHandler(instanceManager, webhookService)

			instances.POST("", instanceHandler.CreateInstance)
			instances.GET("", instanceHandler.ListInstances)
			instances.POST("/:name/connect", instanceHandler.ConnectInstance)
			instances.GET("/:name/status", instanceHandler.GetInstanceStatus)
			instances.DELETE("/:name", instanceHandler.DeleteInstance)
			instances.POST("/:name/logout", instanceHandler.LogoutInstance)
		}

		// Message routes
		messages := api.Group("/message")
		{
			messageHandler := handlers.NewMessageHandler(instanceManager)

			messages.POST("/:name/text", messageHandler.SendText)
			messages.POST("/:name/image", messageHandler.SendImage)
			messages.POST("/:name/video", messageHandler.SendVideo)
			messages.POST("/:name/audio", messageHandler.SendAudio)
			messages.POST("/:name/document", messageHandler.SendDocument)
		}
	}

	// Evolution API compatibility routes
	evolution := r.Group("/")
	{
		instanceHandler := handlers.NewInstanceHandler(instanceManager, webhookService)
		messageHandler := handlers.NewMessageHandler(instanceManager)

		// Instance compatibility routes
		evolution.POST("/instance/create", instanceHandler.CreateInstance)
		evolution.GET("/instance/fetchInstances", instanceHandler.ListInstances)
		evolution.GET("/instance/connect/:name", instanceHandler.ConnectInstance)
		evolution.GET("/instance/connectionState/:name", instanceHandler.GetInstanceStatus)
		evolution.DELETE("/instance/logout/:name", instanceHandler.LogoutInstance)
		evolution.DELETE("/instance/delete/:name", instanceHandler.DeleteInstance)

		// Message compatibility routes
		evolution.POST("/message/sendText/:name", messageHandler.SendText)
		evolution.POST("/message/sendMedia/:name", messageHandler.SendImage)
		evolution.POST("/message/sendImage/:name", messageHandler.SendImage)
		evolution.POST("/message/sendVideo/:name", messageHandler.SendVideo)
		evolution.POST("/message/sendAudio/:name", messageHandler.SendAudio)
		evolution.POST("/message/sendDocument/:name", messageHandler.SendDocument)
	}

	return r
}
