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
// Esta é a única definição da interface no projeto.
type WebhookSender interface {
	SendWebhook(ctx context.Context, url string, event *models.WebhookEvent)
}

// WebhookService implementa a interface WebhookSender.
type WebhookService struct {
	client *http.Client
	logger *zap.Logger
}

// NewWebhookService cria um novo serviço de webhook.
func NewWebhookService(logger *zap.Logger) WebhookSender {
	return &WebhookService{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		logger: logger,
	}
}

// SendWebhook envia um evento para a URL especificada com lógica de retry.
// É executado em uma goroutine, por isso não retorna erros.
func (s *WebhookService) SendWebhook(ctx context.Context, url string, event *models.WebhookEvent) {
	log := s.logger.With(zap.String("event", event.Event), zap.String("instance", event.Instance))
	log.Info("A enviar webhook", zap.String("url", url))

	payload, err := json.Marshal(event)
	if err != nil {
		log.Error("Falha ao serializar o payload do webhook", zap.Error(err))
		return
	}

	maxRetries := 3
	baseDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
		if err != nil {
			log.Error("Falha ao criar o pedido de webhook", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			log.Warn("Falha ao enviar webhook, a tentar novamente...", zap.Int("tentativa", i+1), zap.Error(err))
			time.Sleep(baseDelay)
			baseDelay *= 2 // Exponential backoff
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Info("Webhook enviado com sucesso", zap.Int("status_code", resp.StatusCode))
			resp.Body.Close()
			return
		}

		log.Warn("Webhook recebeu um status de não sucesso, a tentar novamente...",
			zap.Int("tentativa", i+1),
			zap.Int("status_code", resp.StatusCode),
		)
		resp.Body.Close()
		time.Sleep(baseDelay)
		baseDelay *= 2
	}

	log.Error("Falha ao enviar webhook após múltiplas tentativas")
}
