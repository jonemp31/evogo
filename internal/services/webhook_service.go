package services

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jonemp31/evogo/internal/models"
	"go.uber.org/zap"
)

// WebhookSender define a interface para enviar webhooks.
// Isso nos permite trocar a implementação ou usar um mock para testes.
type WebhookSender interface {
	SendWebhook(ctx context.Context, url string, event *models.WebhookEvent) error
}

type WebhookService struct {
	client *http.Client
	logger *zap.Logger
}

func NewWebhookService(logger *zap.Logger) WebhookSender {
	return &WebhookService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SendWebhook envia um evento para a URL especificada com retentativas.
func (s *WebhookService) SendWebhook(ctx context.Context, url string, event *models.WebhookEvent) error {
	log := s.logger.With(zap.String("event", event.Event), zap.String("instance", event.Instance))
	log.Info("Sending webhook", zap.String("url", url))

	payload, err := json.Marshal(event)
	if err != nil {
		log.Error("Failed to marshal webhook payload", zap.Error(err))
		return err // Não adianta tentar de novo se o payload é inválido
	}

	// Lógica de retentativa (Exponential Backoff)
	maxRetries := 3
	baseDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
		if err != nil {
			log.Error("Failed to create webhook request", zap.Error(err))
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		// Adicione aqui outros headers, como uma chave de autenticação, se necessário

		resp, err := s.client.Do(req)
		if err != nil {
			log.Warn("Failed to send webhook, retrying...", zap.Int("attempt", i+1), zap.Error(err))
			time.Sleep(baseDelay)
			baseDelay *= 2 // Aumenta o delay
			continue
		}

		// Se o status for 2xx, consideramos sucesso
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Info("Webhook sent successfully", zap.Int("status_code", resp.StatusCode))
			resp.Body.Close()
			return nil
		}

		log.Warn("Webhook received non-success status, retrying...",
			zap.Int("attempt", i+1),
			zap.Int("status_code", resp.StatusCode),
		)
		resp.Body.Close()
		time.Sleep(baseDelay)
		baseDelay *= 2
	}

	log.Error("Failed to send webhook after multiple retries")
	return nil // Retornamos nil para não bloquear o fluxo principal da aplicação
}
