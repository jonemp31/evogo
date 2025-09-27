package routes

import (
	"github.com/evolution-api/evolution-go/internal/handlers"
	"github.com/evolution-api/evolution-go/internal/middleware"
	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(
	instanceManager *services.InstanceManager,
	messageService *services.MessageService,
	webhookService *services.WebhookService,
) *gin.Engine {
	// Configurar Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Middleware global
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())
	router.Use(middleware.Metrics())

	// Handlers
	instanceHandler := handlers.NewInstanceHandler(instanceManager, webhookService)
	messageHandler := handlers.NewMessageHandler(messageService)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "evolution-api-go",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1
	v1 := router.Group("/api/v1")
	v1.Use(middleware.Auth()) // Aplicar autenticação em todas as rotas da API

	// Instance routes
	instances := v1.Group("/instance")
	{
		instances.POST("/create", instanceHandler.CreateInstance)
		instances.POST("/connect/:instanceName", instanceHandler.ConnectInstance)
		instances.GET("/status/:instanceName", instanceHandler.GetInstanceStatus)
		instances.GET("/list", instanceHandler.ListInstances)
		instances.DELETE("/delete/:instanceName", instanceHandler.DeleteInstance)
		instances.POST("/logout/:instanceName", instanceHandler.LogoutInstance)
		instances.GET("/qrcode/:instanceName", instanceHandler.GetQRCode)
		instances.POST("/webhook/:instanceName", instanceHandler.UpdateInstanceWebhook)
		instances.POST("/settings/:instanceName", instanceHandler.UpdateInstanceSettings)
	}

	// Message routes
	messages := v1.Group("/message")
	{
		messages.POST("/sendText/:instanceName", messageHandler.SendText)
		messages.POST("/sendMedia/:instanceName", messageHandler.SendMedia)
		messages.POST("/sendImage/:instanceName", messageHandler.SendImage)
		messages.POST("/sendVideo/:instanceName", messageHandler.SendVideo)
		messages.POST("/sendAudio/:instanceName", messageHandler.SendAudio)
		messages.POST("/sendDocument/:instanceName", messageHandler.SendDocument)
		messages.GET("/status/:instanceName/:messageId", messageHandler.GetMessageStatus)
	}

	// Webhook routes (sem autenticação para receber webhooks externos)
	webhooks := router.Group("/webhook")
	{
		webhooks.POST("/:instanceName", func(c *gin.Context) {
			// Handler para receber webhooks externos
			instanceName := c.Param("instanceName")
			zap.L().Info("Received webhook", zap.String("instance", instanceName))
			c.JSON(200, gin.H{"status": "received"})
		})
	}

	// Documentação da API
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "Evolution API Go",
			"version":     "1.0.0",
			"description": "WhatsApp API built with Go",
			"endpoints": gin.H{
				"instances": gin.H{
					"POST /api/v1/instance/create":                  "Create new instance",
					"POST /api/v1/instance/connect/{instanceName}":  "Connect instance",
					"GET /api/v1/instance/status/{instanceName}":    "Get instance status",
					"GET /api/v1/instance/list":                     "List all instances",
					"DELETE /api/v1/instance/delete/{instanceName}": "Delete instance",
					"POST /api/v1/instance/logout/{instanceName}":   "Logout instance",
					"GET /api/v1/instance/qrcode/{instanceName}":    "Get QR code",
					"POST /api/v1/instance/webhook/{instanceName}":  "Update webhook URL",
					"POST /api/v1/instance/settings/{instanceName}": "Update instance settings",
				},
				"messages": gin.H{
					"POST /api/v1/message/sendText/{instanceName}":      "Send text message",
					"POST /api/v1/message/sendMedia/{instanceName}":     "Send media message",
					"POST /api/v1/message/sendImage/{instanceName}":     "Send image message",
					"POST /api/v1/message/sendVideo/{instanceName}":     "Send video message",
					"POST /api/v1/message/sendAudio/{instanceName}":     "Send audio message",
					"POST /api/v1/message/sendDocument/{instanceName}":  "Send document message",
					"GET /api/v1/message/status/{instanceName}/{msgId}": "Get message status",
				},
				"webhooks": gin.H{
					"POST /webhook/{instanceName}": "Receive external webhook",
				},
				"system": gin.H{
					"GET /health":  "Health check",
					"GET /metrics": "Prometheus metrics",
					"GET /docs":    "API documentation",
				},
			},
			"authentication": "Use 'apikey' header with your API key",
			"webhook_events": []string{
				"connection.update",
				"messages.upsert",
				"messages.update",
				"presence.update",
				"chat.presence",
			},
		})
	})

	return router
}
