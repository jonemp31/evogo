package models

import (
	"time"
)

// Instance representa a estrutura de uma instância do WhatsApp.
type Instance struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	ApiKey     string    `json:"api_key"` // Campo que estava faltando
	WebhookURL string    `json:"webhook_url"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WebhookEvent representa o payload padrão para todos os eventos de webhook.
type WebhookEvent struct {
	Event    string      `json:"event"`
	Instance string      `json:"instance"`
	Data     interface{} `json:"data"`
}
