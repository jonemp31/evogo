package models

import (
	"time"
)

// Instance represents a WhatsApp instance
type Instance struct {
	ID             string            `json:"id" db:"id"`
	Name           string            `json:"name" db:"name"`
	Token          string            `json:"token" db:"token"`
	Status         string            `json:"status" db:"status"` // open, close, connecting
	OwnerJID       string            `json:"ownerJid,omitempty" db:"owner_jid"`
	ProfileName    string            `json:"profileName,omitempty" db:"profile_name"`
	ProfilePicURL  string            `json:"profilePicUrl,omitempty" db:"profile_pic_url"`
	Number         string            `json:"number,omitempty" db:"number"`
	BusinessID     string            `json:"businessId,omitempty" db:"business_id"`
	Integration    string            `json:"integration" db:"integration"`
	WebhookURL     string            `json:"webhookUrl,omitempty" db:"webhook_url"`
	WebhookEvents  []string          `json:"webhookEvents,omitempty" db:"webhook_events"`
	WebhookHeaders map[string]string `json:"webhookHeaders,omitempty" db:"webhook_headers"`
	WebhookBase64  bool              `json:"webhookBase64" db:"webhook_base64"`
	Settings       InstanceSettings  `json:"settings" db:"settings"`
	CreatedAt      time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time         `json:"updatedAt" db:"updated_at"`
	LastSeen       time.Time         `json:"lastSeen,omitempty" db:"last_seen"`
}

// InstanceSettings holds instance-specific settings
type InstanceSettings struct {
	RejectCall       bool   `json:"rejectCall" db:"reject_call"`
	MsgCall          string `json:"msgCall,omitempty" db:"msg_call"`
	GroupsIgnore     bool   `json:"groupsIgnore" db:"groups_ignore"`
	AlwaysOnline     bool   `json:"alwaysOnline" db:"always_online"`
	ReadMessages     bool   `json:"readMessages" db:"read_messages"`
	ReadStatus       bool   `json:"readStatus" db:"read_status"`
	SyncFullHistory  bool   `json:"syncFullHistory" db:"sync_full_history"`
	WAVoipToken      string `json:"wavoipToken,omitempty" db:"wavoip_token"`
	AutoReadMessages bool   `json:"autoReadMessages" db:"auto_read_messages"`
	ReadDelay        int    `json:"readDelay" db:"read_delay"`
}

// Message represents a WhatsApp message
type Message struct {
	ID            string    `json:"id" db:"id"`
	InstanceID    string    `json:"instanceId" db:"instance_id"`
	RemoteJID     string    `json:"remoteJid" db:"remote_jid"`
	FromMe        bool      `json:"fromMe" db:"from_me"`
	Message       string    `json:"message" db:"message"`
	MessageType   string    `json:"messageType" db:"message_type"`
	MediaURL      string    `json:"mediaUrl,omitempty" db:"media_url"`
	Caption       string    `json:"caption,omitempty" db:"caption"`
	MimeType      string    `json:"mimeType,omitempty" db:"mime_type"`
	FileName      string    `json:"fileName,omitempty" db:"file_name"`
	QuotedID      string    `json:"quotedId,omitempty" db:"quoted_id"`
	QuotedMessage string    `json:"quotedMessage,omitempty" db:"quoted_message"`
	ViewOnce      bool      `json:"viewOnce" db:"view_once"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	Status        string    `json:"status" db:"status"` // sent, delivered, read, failed
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// Contact represents a WhatsApp contact
type Contact struct {
	ID            string    `json:"id" db:"id"`
	InstanceID    string    `json:"instanceId" db:"instance_id"`
	RemoteJID     string    `json:"remoteJid" db:"remote_jid"`
	PushName      string    `json:"pushName,omitempty" db:"push_name"`
	ProfilePicURL string    `json:"profilePicUrl,omitempty" db:"profile_pic_url"`
	IsOnWhatsApp  bool      `json:"isOnWhatsApp" db:"is_on_whatsapp"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// Chat represents a WhatsApp chat
type Chat struct {
	ID             string    `json:"id" db:"id"`
	InstanceID     string    `json:"instanceId" db:"instance_id"`
	RemoteJID      string    `json:"remoteJid" db:"remote_jid"`
	Name           string    `json:"name,omitempty" db:"name"`
	UnreadMessages int       `json:"unreadMessages" db:"unread_messages"`
	IsGroup        bool      `json:"isGroup" db:"is_group"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	Event     string      `json:"event"`
	Instance  string      `json:"instance"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// ConnectionStatus represents the connection status of an instance
type ConnectionStatus struct {
	InstanceName string `json:"instanceName"`
	State        string `json:"state"`
	RemoteJID    string `json:"remoteJid,omitempty"`
}

// QRCode represents a QR code for authentication
type QRCode struct {
	Base64 string `json:"base64"`
	Code   string `json:"code"`
}
