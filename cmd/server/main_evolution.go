package main

import (
	"context"
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
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
)

func main() {
	// Configurar logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Carregar configuração
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger.Info("Starting Evolution API Go",
		zap.String("version", "1.0.0"),
		zap.String("port", cfg.Server.Port))

	// Conectar ao banco de dados
	db, err := repository.NewDatabase(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Criar container do sqlstore para WhatsApp
	container, err := sqlstore.New("postgres", cfg.Database.DSN, nil)
	if err != nil {
		logger.Fatal("Failed to create sqlstore container", zap.Error(err))
	}

	// Criar repositório de instâncias
	instanceRepo := repository.NewInstanceRepository(db)

	// Criar serviços
	webhookService := services.NewWebhookService()
	instanceManager := services.NewInstanceManager(container, instanceRepo)
	messageService := services.NewMessageService(instanceManager)

	// Configurar rotas
	router := routes.SetupRoutes(instanceManager, messageService, webhookService)

	// Criar servidor HTTP
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Canal para capturar sinais do sistema
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor em goroutine
	go func() {
		logger.Info("Server starting", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Aguardar sinal de parada
	<-quit
	logger.Info("Shutting down server...")

	// Criar contexto com timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Desconectar todas as instâncias
	instanceManager.Shutdown()

	// Fechar servidor
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server shutdown completed")
	}
}
