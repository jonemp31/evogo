package dto

import "time"

// SendTextRequest representa a requisição para enviar texto
type SendTextRequest struct {
	Number  string `json:"number" binding:"required"`
	Text    string `json:"text" binding:"required"`
	Options struct {
		Delay    int    `json:"delay"`
		Presence string `json:"presence"`
	} `json:"options"`
}

// SendMediaRequest representa a requisição para enviar mídia
type SendMediaRequest struct {
	Number    string `form:"number" binding:"required"`
	MediaURL  string `form:"mediaUrl"`
	MediaType string `form:"mediaType" binding:"required"`
	Caption   string `form:"caption"`
	Options   struct {
		Delay    int    `form:"delay"`
		Presence string `form:"presence"`
		ViewOnce bool   `form:"viewOnce"`
		Ptt      bool   `form:"ptt"` // Para áudios narrados
	} `form:"options"`
}

// MessageResponse representa a resposta de envio de mensagem
type MessageResponse struct {
	Success   bool      `json:"success"`
	MessageID string    `json:"messageId"`
	Instance  string    `json:"instance"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorResponse representa uma resposta de erro
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// InstanceWebhookRequest representa a requisição para configurar webhook
type InstanceWebhookRequest struct {
	WebhookURL string   `json:"webhookUrl" binding:"required"`
	Events     []string `json:"events"`
}

// InstanceSettingsRequest representa as configurações da instância
type InstanceSettingsRequest struct {
	RejectCall           bool   `json:"rejectCall"`
	MsgRetryCounterCache string `json:"msgRetryCounterCache"`
	UserAgent            string `json:"userAgent"`
	AlwaysOnline         bool   `json:"alwaysOnline"`
	ReadMessages         bool   `json:"readMessages"`
	ReadStatus           bool   `json:"readStatus"`
	SyncFullHistory      bool   `json:"syncFullHistory"`
	MarkOnlineOnConnect  bool   `json:"markOnlineOnConnect"`
	DefaultReactionEmoji string `json:"defaultReactionEmoji"`
}

// CreateInstanceRequest representa a requisição para criar instância
type CreateInstanceRequest struct {
	InstanceName    string                   `json:"instanceName" binding:"required"`
	Token           string                   `json:"token"`
	WebhookURL      string                   `json:"webhookUrl"`
	WebhookByEvents bool                     `json:"webhookByEvents"`
	Events          []string                 `json:"events"`
	Settings        *InstanceSettingsRequest `json:"settings"`
}

// InstanceStatusResponse representa o status da instância
type InstanceStatusResponse struct {
	Instance struct {
		Name       string    `json:"name"`
		Status     string    `json:"status"`
		WebhookURL string    `json:"webhookUrl"`
		CreatedAt  time.Time `json:"createdAt"`
		UpdatedAt  time.Time `json:"updatedAt"`
	} `json:"instance"`
}

// InstanceListResponse representa a lista de instâncias
type InstanceListResponse struct {
	Instances []InstanceStatusResponse `json:"instances"`
	Total     int                      `json:"total"`
}

// QRCodeResponse representa a resposta do QR code
type QRCodeResponse struct {
	Instance string `json:"instance"`
	QRCode   string `json:"qrCode"`
	Status   string `json:"status"`
}

// ConnectionStatusResponse representa o status de conexão
type ConnectionStatusResponse struct {
	Instance string `json:"instance"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}
