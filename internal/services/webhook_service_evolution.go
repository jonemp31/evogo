package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// WebhookService gerencia o envio de webhooks
type WebhookService struct {
	httpClient *http.Client
	logger     *zap.Logger
}

// NewWebhookService cria um novo serviço de webhook
func NewWebhookService() *WebhookService {
	return &WebhookService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: zap.L().Named("WebhookService"),
	}
}

// WebhookPayload representa o payload de um webhook
type WebhookPayload struct {
	Event     string      `json:"event"`
	Instance  string      `json:"instance"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// SendWebhook envia um webhook para a URL especificada
func (ws *WebhookService) SendWebhook(webhookURL string, payload interface{}) {
	if webhookURL == "" {
		ws.logger.Debug("Webhook URL is empty, skipping webhook")
		return
	}

	// Criar payload completo
	webhookData := &WebhookPayload{
		Event:     ws.extractEvent(payload),
		Instance:  ws.extractInstance(payload),
		Data:      payload,
		Timestamp: time.Now(),
	}

	// Serializar payload para JSON
	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		ws.logger.Error("Failed to marshal webhook payload", zap.Error(err))
		return
	}

	// Enviar webhook de forma assíncrona
	go ws.sendWebhookAsync(webhookURL, jsonData)
}

// sendWebhookAsync envia o webhook de forma assíncrona com retry
func (ws *WebhookService) sendWebhookAsync(webhookURL string, jsonData []byte) {
	maxRetries := 3
	retryDelay := time.Second * 2

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := ws.sendHTTPRequest(webhookURL, jsonData)
		if err == nil {
			ws.logger.Debug("Webhook sent successfully",
				zap.String("url", webhookURL),
				zap.Int("attempt", attempt))
			return
		}

		ws.logger.Warn("Failed to send webhook",
			zap.String("url", webhookURL),
			zap.Int("attempt", attempt),
			zap.Int("maxRetries", maxRetries),
			zap.Error(err))

		if attempt < maxRetries {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	ws.logger.Error("Failed to send webhook after all retries",
		zap.String("url", webhookURL),
		zap.Int("maxRetries", maxRetries))
}

// sendHTTPRequest envia a requisição HTTP
func (ws *WebhookService) sendHTTPRequest(webhookURL string, jsonData []byte) error {
	req, err := http.NewRequestWithContext(context.Background(), "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Evolution-API-Go/1.0")

	resp, err := ws.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// extractEvent extrai o evento do payload
func (ws *WebhookService) extractEvent(payload interface{}) string {
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if event, exists := payloadMap["event"]; exists {
			if eventStr, ok := event.(string); ok {
				return eventStr
			}
		}
	}
	return "unknown"
}

// extractInstance extrai a instância do payload
func (ws *WebhookService) extractInstance(payload interface{}) string {
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if instance, exists := payloadMap["instance"]; exists {
			if instanceStr, ok := instance.(string); ok {
				return instanceStr
			}
		}
	}
	return "unknown"
}

// SendConnectionWebhook envia webhook de status de conexão
func (ws *WebhookService) SendConnectionWebhook(instanceName, webhookURL, status string) {
	payload := map[string]interface{}{
		"event":    "connection.update",
		"instance": instanceName,
		"data": map[string]interface{}{
			"state": status,
		},
	}

	ws.SendWebhook(webhookURL, payload)
}

// SendMessageWebhook envia webhook de mensagem
func (ws *WebhookService) SendMessageWebhook(instanceName, webhookURL, event string, message interface{}) {
	payload := map[string]interface{}{
		"event":    event,
		"instance": instanceName,
		"data":     message,
	}

	ws.SendWebhook(webhookURL, payload)
}

// SendQRCodeWebhook envia webhook de QR code
func (ws *WebhookService) SendQRCodeWebhook(instanceName, webhookURL, qrCode string) {
	payload := map[string]interface{}{
		"event":    "connection.update",
		"instance": instanceName,
		"data": map[string]interface{}{
			"state":  "qrcode",
			"qrCode": qrCode,
		},
	}

	ws.SendWebhook(webhookURL, payload)
}
