package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"go.uber.org/zap"
)

// WebhookService handles webhook operations
type WebhookService struct {
	httpClient    *http.Client
	retryAttempts int
	timeout       time.Duration
}

// NewWebhookService creates a new webhook service
func NewWebhookService(timeout time.Duration, retryAttempts int) *WebhookService {
	return &WebhookService{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		retryAttempts: retryAttempts,
		timeout:       timeout,
	}
}

// SendWebhook sends a webhook event
func (ws *WebhookService) SendWebhook(ctx context.Context, url string, event *models.WebhookEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Evolution-API-Go/1.0")

	// Send with retries
	for attempt := 1; attempt <= ws.retryAttempts; attempt++ {
		resp, err := ws.httpClient.Do(req)
		if err != nil {
			zap.L().Warn("Webhook send attempt failed",
				zap.Int("attempt", attempt),
				zap.String("url", url),
				zap.Error(err))

			if attempt == ws.retryAttempts {
				return fmt.Errorf("failed to send webhook after %d attempts: %w", ws.retryAttempts, err)
			}

			// Wait before retry
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			zap.L().Debug("Webhook sent successfully",
				zap.String("url", url),
				zap.String("event", event.Event),
				zap.Int("status", resp.StatusCode))
			return nil
		}

		zap.L().Warn("Webhook returned non-success status",
			zap.Int("attempt", attempt),
			zap.String("url", url),
			zap.Int("status", resp.StatusCode))

		if attempt == ws.retryAttempts {
			return fmt.Errorf("webhook returned status %d", resp.StatusCode)
		}

		// Wait before retry
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return nil
}

// SendInstanceEvent sends an instance-related webhook event
func (ws *WebhookService) SendInstanceEvent(ctx context.Context, instance *models.Instance, eventType string, data interface{}) {
	if instance.WebhookURL == "" {
		return
	}

	// Check if event should be sent based on webhook events filter
	if len(instance.WebhookEvents) > 0 {
		shouldSend := false
		for _, event := range instance.WebhookEvents {
			if event == eventType {
				shouldSend = true
				break
			}
		}
		if !shouldSend {
			return
		}
	}

	webhookEvent := &models.WebhookEvent{
		Event:     eventType,
		Instance:  instance.Name,
		Data:      data,
		Timestamp: time.Now(),
	}

	// Send webhook asynchronously
	go func() {
		if err := ws.SendWebhook(context.Background(), instance.WebhookURL, webhookEvent); err != nil {
			zap.L().Error("Failed to send webhook",
				zap.String("instance", instance.Name),
				zap.String("event", eventType),
				zap.Error(err))
		}
	}()
}

// SendMessageEvent sends a message-related webhook event
func (ws *WebhookService) SendMessageEvent(ctx context.Context, instance *models.Instance, eventType string, message *models.Message) {
	if instance.WebhookURL == "" {
		return
	}

	// Check if event should be sent based on webhook events filter
	if len(instance.WebhookEvents) > 0 {
		shouldSend := false
		for _, event := range instance.WebhookEvents {
			if event == eventType {
				shouldSend = true
				break
			}
		}
		if !shouldSend {
			return
		}
	}

	webhookEvent := &models.WebhookEvent{
		Event:     eventType,
		Instance:  instance.Name,
		Data:      message,
		Timestamp: time.Now(),
	}

	// Send webhook asynchronously
	go func() {
		if err := ws.SendWebhook(context.Background(), instance.WebhookURL, webhookEvent); err != nil {
			zap.L().Error("Failed to send message webhook",
				zap.String("instance", instance.Name),
				zap.String("event", eventType),
				zap.Error(err))
		}
	}()
}

// SendConnectionEvent sends a connection status webhook event
func (ws *WebhookService) SendConnectionEvent(ctx context.Context, instance *models.Instance, status string) {
	if instance.WebhookURL == "" {
		return
	}

	// Check if event should be sent based on webhook events filter
	if len(instance.WebhookEvents) > 0 {
		shouldSend := false
		for _, event := range instance.WebhookEvents {
			if event == "connection.update" {
				shouldSend = true
				break
			}
		}
		if !shouldSend {
			return
		}
	}

	connectionData := map[string]interface{}{
		"instance": map[string]interface{}{
			"instanceName": instance.Name,
			"state":        status,
		},
	}

	webhookEvent := &models.WebhookEvent{
		Event:     "connection.update",
		Instance:  instance.Name,
		Data:      connectionData,
		Timestamp: time.Now(),
	}

	// Send webhook asynchronously
	go func() {
		if err := ws.SendWebhook(context.Background(), instance.WebhookURL, webhookEvent); err != nil {
			zap.L().Error("Failed to send connection webhook",
				zap.String("instance", instance.Name),
				zap.Error(err))
		}
	}()
}

// SendQRCodeEvent sends a QR code webhook event
func (ws *WebhookService) SendQRCodeEvent(ctx context.Context, instance *models.Instance, qrCode string) {
	if instance.WebhookURL == "" {
		return
	}

	// Check if event should be sent based on webhook events filter
	if len(instance.WebhookEvents) > 0 {
		shouldSend := false
		for _, event := range instance.WebhookEvents {
			if event == "qrcode.updated" {
				shouldSend = true
				break
			}
		}
		if !shouldSend {
			return
		}
	}

	qrData := map[string]interface{}{
		"base64": fmt.Sprintf("data:image/png;base64,%s", qrCode),
		"code":   qrCode,
	}

	webhookEvent := &models.WebhookEvent{
		Event:     "qrcode.updated",
		Instance:  instance.Name,
		Data:      qrData,
		Timestamp: time.Now(),
	}

	// Send webhook asynchronously
	go func() {
		if err := ws.SendWebhook(context.Background(), instance.WebhookURL, webhookEvent); err != nil {
			zap.L().Error("Failed to send QR code webhook",
				zap.String("instance", instance.Name),
				zap.Error(err))
		}
	}()
}
