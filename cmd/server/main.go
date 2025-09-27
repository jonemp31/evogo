package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jonemp31/evogo/internal/config"
	"github.com/jonemp31/evogo/internal/handlers"
	"github.com/jonemp31/evogo/internal/repository"
	"github.com/jonemp31/evogo/internal/routes"
	"github.com/jonemp31/evogo/internal/services"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	// 1. Carregar Configurações
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// 2. Configurar Logger
	logger, err := zap.NewProduction()
	if cfg.Server.Env == "development" {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// 3. Conexão com o Banco de Dados
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("Cannot connect to database", zap.Error(err))
	}
	if err := db.Ping(); err != nil {
		logger.Fatal("Cannot ping database", zap.Error(err))
	}

	// 4. Inicializar Camadas da Aplicação (na ordem correta)
	// Repositório
	instanceRepo := repository.NewPostgresInstanceRepository(db)

	// Serviços
	webhookService := services.NewWebhookService(logger)
	instanceManager := services.NewInstanceManager(instanceRepo, webhookService, logger)

	// Handlers
	instanceHandler := handlers.NewInstanceHandler(instanceManager)
	messageHandler := handlers.NewMessageHandler(instanceManager)
	healthHandler := handlers.NewHealthHandler(db)

	// Restaurar instâncias em background
	go instanceManager.RestoreInstances()

	// 5. Configurar Roteador
	router := gin.Default()
	routes.SetupRoutes(router, instanceHandler, messageHandler, healthHandler)

	// 6. Iniciar Servidor
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Starting server", zap.String("address", serverAddr))
	if err := http.ListenAndServe(serverAddr, router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
