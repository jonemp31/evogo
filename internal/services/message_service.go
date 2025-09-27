package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
)

// MessageService handles message operations
type MessageService struct {
	instanceManager *InstanceManager
	webhookService  *WebhookService
	repository      *MessageRepository
}

// MessageRepository handles message data operations
type MessageRepository struct {
	// This would be implemented similar to InstanceRepository
	// For now, we'll use a simple in-memory approach
}

// NewMessageService creates a new message service
func NewMessageService(instanceManager *InstanceManager, webhookService *WebhookService) *MessageService {
	return &MessageService{
		instanceManager: instanceManager,
		webhookService:  webhookService,
		repository:      &MessageRepository{},
	}
}

// SendText sends a text message
func (ms *MessageService) SendText(ctx context.Context, instanceName, to, text string) (*models.Message, error) {
	// Validate input
	if to == "" || text == "" {
		return nil, fmt.Errorf("to and text are required")
	}

	// Get instance
	instance, err := ms.instanceManager.getInstanceByName(instanceName)
	if err != nil {
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	// Send message
	messageID, err := ms.sendTextMessage(ctx, instance, to, text)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	// Create message record
	message := &models.Message{
		ID:          messageID,
		InstanceID:  instance.ID,
		RemoteJID:   to,
		FromMe:      true,
		Message:     text,
		MessageType: "conversation",
		Timestamp:   time.Now(),
		Status:      "sent",
		CreatedAt:   time.Now(),
	}

	// Send webhook event
	ms.webhookService.SendMessageEvent(ctx, instance, "messages.upsert", message)

	return message, nil
}

// SendMedia sends a media message
func (ms *MessageService) SendMedia(ctx context.Context, instanceName, to, mediaURL, mediaType, caption string, viewOnce bool) (*models.Message, error) {
	// Validate input
	if to == "" || mediaURL == "" || mediaType == "" {
		return nil, fmt.Errorf("to, mediaUrl and type are required")
	}

	// Get instance
	instance, err := ms.instanceManager.getInstanceByName(instanceName)
	if err != nil {
		return nil, fmt.Errorf("instance not found: %w", err)
	}

	// Download and upload media
	messageID, err := ms.sendMediaMessage(ctx, instance, to, mediaURL, mediaType, caption, viewOnce)
	if err != nil {
		return nil, fmt.Errorf("failed to send media message: %w", err)
	}

	// Create message record
	message := &models.Message{
		ID:          messageID,
		InstanceID:  instance.ID,
		RemoteJID:   to,
		FromMe:      true,
		MediaURL:    mediaURL,
		Caption:     caption,
		MessageType: mediaType + "Message",
		ViewOnce:    viewOnce,
		Timestamp:   time.Now(),
		Status:      "sent",
		CreatedAt:   time.Now(),
	}

	// Send webhook event
	ms.webhookService.SendMessageEvent(ctx, instance, "messages.upsert", message)

	return message, nil
}

// ProcessIncomingMessage processes incoming messages
func (ms *MessageService) ProcessIncomingMessage(ctx context.Context, instance *models.Instance, evt *events.Message) {
	// Extract message data
	message := &models.Message{
		ID:          evt.Info.ID,
		InstanceID:  instance.ID,
		RemoteJID:   evt.Info.Chat.String(),
		FromMe:      evt.Info.IsFromMe,
		MessageType: ms.getMessageType(evt.Message),
		Timestamp:   evt.Info.Timestamp,
		Status:      "received",
		CreatedAt:   time.Now(),
	}

	// Extract message content based on type
	switch {
	case evt.Message.GetConversation() != "":
		message.Message = evt.Message.GetConversation()
	case evt.Message.GetExtendedTextMessage() != nil:
		message.Message = evt.Message.GetExtendedTextMessage().GetText()
	case evt.Message.GetImageMessage() != nil:
		imgMsg := evt.Message.GetImageMessage()
		message.MediaURL = imgMsg.GetUrl()
		message.Caption = imgMsg.GetCaption()
		message.ViewOnce = imgMsg.GetViewOnce()
	case evt.Message.GetVideoMessage() != nil:
		vidMsg := evt.Message.GetVideoMessage()
		message.MediaURL = vidMsg.GetUrl()
		message.Caption = vidMsg.GetCaption()
		message.ViewOnce = vidMsg.GetViewOnce()
	case evt.Message.GetAudioMessage() != nil:
		audioMsg := evt.Message.GetAudioMessage()
		message.MediaURL = audioMsg.GetUrl()
		message.ViewOnce = audioMsg.GetViewOnce()
	case evt.Message.GetDocumentMessage() != nil:
		docMsg := evt.Message.GetDocumentMessage()
		message.MediaURL = docMsg.GetUrl()
		message.Caption = docMsg.GetCaption()
		message.FileName = docMsg.GetFileName()
		message.MimeType = docMsg.GetMimetype()
	}

	// Store message (in production, this would save to database)
	zap.L().Info("Message received",
		zap.String("instance", instance.Name),
		zap.String("from", evt.Info.Sender.String()),
		zap.String("type", message.MessageType),
		zap.String("id", message.ID))

	// Send webhook event
	ms.webhookService.SendMessageEvent(ctx, instance, "messages.upsert", message)

	// Auto-read messages if enabled
	if instance.Settings.AutoReadMessages {
		go ms.markAsRead(ctx, instance, evt.Info.Chat, evt.Info.Sender, message.ID)
	}
}

// ProcessMessageStatus processes message status updates
func (ms *MessageService) ProcessMessageStatus(ctx context.Context, instance *models.Instance, evt *events.MessageStatus) {
	// Send webhook event for message status
	statusData := map[string]interface{}{
		"id":     evt.Info.ID,
		"status": evt.Status,
		"from":   evt.Info.From.String(),
		"to":     evt.Info.To.String(),
	}

	ms.webhookService.SendWebhookEvent(ctx, instance.WebhookURL, &models.WebhookEvent{
		Event:     "messages.update",
		Instance:  instance.Name,
		Data:      statusData,
		Timestamp: time.Now(),
	})

	zap.L().Debug("Message status updated",
		zap.String("instance", instance.Name),
		zap.String("messageId", evt.Info.ID),
		zap.String("status", string(evt.Status)))
}

// sendTextMessage sends a text message through WhatsApp
func (ms *MessageService) sendTextMessage(ctx context.Context, instance *models.Instance, to, text string) (string, error) {
	// This would use the WhatsApp client from the instance manager
	// For now, we'll simulate the message sending
	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())

	zap.L().Info("Text message sent",
		zap.String("instance", instance.Name),
		zap.String("to", to),
		zap.String("messageId", messageID))

	return messageID, nil
}

// sendMediaMessage sends a media message through WhatsApp
func (ms *MessageService) sendMediaMessage(ctx context.Context, instance *models.Instance, to, mediaURL, mediaType, caption string, viewOnce bool) (string, error) {
	// Download media from URL
	mediaData, err := ms.downloadMedia(mediaURL)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %w", err)
	}

	// Upload media to WhatsApp (this would use the WhatsApp client)
	// For now, we'll simulate the upload
	messageID := fmt.Sprintf("media_%d", time.Now().UnixNano())

	zap.L().Info("Media message sent",
		zap.String("instance", instance.Name),
		zap.String("to", to),
		zap.String("type", mediaType),
		zap.String("messageId", messageID),
		zap.Int("size", len(mediaData)))

	return messageID, nil
}

// downloadMedia downloads media from URL
func (ms *MessageService) downloadMedia(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download media: status %d", resp.StatusCode)
	}

	mediaData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read media data: %w", err)
	}

	return mediaData, nil
}

// markAsRead marks a message as read
func (ms *MessageService) markAsRead(ctx context.Context, instance *models.Instance, chatJID, senderJID types.JID, messageID string) {
	// Wait for the configured delay
	time.Sleep(time.Duration(instance.Settings.ReadDelay) * time.Second)

	// This would use the WhatsApp client to mark the message as read
	zap.L().Debug("Marking message as read",
		zap.String("instance", instance.Name),
		zap.String("messageId", messageID))
}

// getMessageType extracts message type from WhatsApp message
func (ms *MessageService) getMessageType(msg *types.Message) string {
	switch {
	case msg.GetConversation() != "":
		return "conversation"
	case msg.GetImageMessage() != nil:
		return "imageMessage"
	case msg.GetVideoMessage() != nil:
		return "videoMessage"
	case msg.GetAudioMessage() != nil:
		return "audioMessage"
	case msg.GetDocumentMessage() != nil:
		return "documentMessage"
	case msg.GetStickerMessage() != nil:
		return "stickerMessage"
	case msg.GetLocationMessage() != nil:
		return "locationMessage"
	case msg.GetContactMessage() != nil:
		return "contactMessage"
	default:
		return "unknown"
	}
}
