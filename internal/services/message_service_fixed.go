package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
)

// MessageService handles message-related operations
type MessageService struct {
	instanceManager *InstanceManager
	logger          *zap.Logger
}

// NewMessageService creates a new message service
func NewMessageService(instanceManager *InstanceManager) *MessageService {
	return &MessageService{
		instanceManager: instanceManager,
		logger:          zap.L().Named("MessageService"),
	}
}

// SendText sends a text message
func (ms *MessageService) SendText(ctx context.Context, instanceName, to, message string) (string, error) {
	ms.logger.Info("Sending text message",
		zap.String("instance", instanceName),
		zap.String("to", to))

	// Get WhatsApp client
	client, err := ms.instanceManager.getInstanceByName(instanceName)
	if err != nil {
		return "", fmt.Errorf("instance not found: %w", err)
	}

	// Get client from manager
	waClient := ms.instanceManager.getClient(instanceName)
	if waClient == nil {
		return "", fmt.Errorf("WhatsApp client not found for instance %s", instanceName)
	}

	// Send message
	messageID, err := waClient.SendText(ctx, to, message)
	if err != nil {
		return "", fmt.Errorf("failed to send text message: %w", err)
	}

	// Store message in database
	msg := &models.Message{
		ID:          messageID,
		InstanceID:  client.ID,
		RemoteJID:   to,
		FromMe:      true,
		MessageType: "conversation",
		Message:     message,
		Status:      "sent",
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}

	// Save to database (implementation depends on repository)
	ms.logger.Info("Text message sent successfully",
		zap.String("instance", instanceName),
		zap.String("messageId", messageID))

	return messageID, nil
}

// SendMedia sends a media message
func (ms *MessageService) SendMedia(ctx context.Context, instanceName, to, mediaURL, mediaType, caption string, viewOnce bool) (string, error) {
	ms.logger.Info("Sending media message",
		zap.String("instance", instanceName),
		zap.String("to", to),
		zap.String("mediaType", mediaType),
		zap.Bool("viewOnce", viewOnce))

	// Get WhatsApp client
	client, err := ms.instanceManager.getInstanceByName(instanceName)
	if err != nil {
		return "", fmt.Errorf("instance not found: %w", err)
	}

	// Get client from manager
	waClient := ms.instanceManager.getClient(instanceName)
	if waClient == nil {
		return "", fmt.Errorf("WhatsApp client not found for instance %s", instanceName)
	}

	// Send media message
	messageID, err := waClient.SendMedia(ctx, to, mediaURL, mediaType, caption, viewOnce)
	if err != nil {
		return "", fmt.Errorf("failed to send media message: %w", err)
	}

	// Store message in database
	msg := &models.Message{
		ID:          messageID,
		InstanceID:  client.ID,
		RemoteJID:   to,
		FromMe:      true,
		MessageType: ms.getMediaMessageType(mediaType),
		MediaURL:    mediaURL,
		Caption:     caption,
		ViewOnce:    viewOnce,
		Status:      "sent",
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}

	// Save to database (implementation depends on repository)
	ms.logger.Info("Media message sent successfully",
		zap.String("instance", instanceName),
		zap.String("messageId", messageID),
		zap.String("mediaType", mediaType))

	return messageID, nil
}

// ProcessIncomingMessage processes incoming WhatsApp messages
func (ms *MessageService) ProcessIncomingMessage(ctx context.Context, evt *events.Message) {
	ms.logger.Info("Processing incoming message",
		zap.String("id", evt.Info.ID),
		zap.String("chat", evt.Info.Chat.String()),
		zap.Bool("fromMe", evt.Info.IsFromMe))

	// Skip messages from me
	if evt.Info.IsFromMe {
		return
	}

	// Create message model
	message := &models.Message{
		ID:        evt.Info.ID,
		RemoteJID: evt.Info.Chat.String(),
		FromMe:    evt.Info.IsFromMe,
		Timestamp: evt.Info.Timestamp,
		Status:    "received",
		CreatedAt: time.Now(),
	}

	// Extract message content based on type
	switch {
	case evt.Message.GetConversation() != "":
		message.Message = evt.Message.GetConversation()
		message.MessageType = "conversation"
	case evt.Message.GetExtendedTextMessage() != nil:
		message.Message = evt.Message.GetExtendedTextMessage().GetText()
		message.MessageType = "extendedTextMessage"
	case evt.Message.GetImageMessage() != nil:
		imgMsg := evt.Message.GetImageMessage()
		message.MediaURL = imgMsg.GetURL()
		message.Caption = imgMsg.GetCaption()
		message.ViewOnce = imgMsg.GetViewOnce()
		message.MessageType = "imageMessage"
	case evt.Message.GetVideoMessage() != nil:
		vidMsg := evt.Message.GetVideoMessage()
		message.MediaURL = vidMsg.GetURL()
		message.Caption = vidMsg.GetCaption()
		message.ViewOnce = vidMsg.GetViewOnce()
		message.MessageType = "videoMessage"
	case evt.Message.GetAudioMessage() != nil:
		audioMsg := evt.Message.GetAudioMessage()
		message.MediaURL = audioMsg.GetURL()
		message.ViewOnce = audioMsg.GetViewOnce()
		message.MessageType = "audioMessage"
	case evt.Message.GetDocumentMessage() != nil:
		docMsg := evt.Message.GetDocumentMessage()
		message.MediaURL = docMsg.GetURL()
		message.Caption = docMsg.GetCaption()
		message.FileName = docMsg.GetFileName()
		message.MessageType = "documentMessage"
	case evt.Message.GetStickerMessage() != nil:
		stickerMsg := evt.Message.GetStickerMessage()
		message.MediaURL = stickerMsg.GetURL()
		message.MessageType = "stickerMessage"
	case evt.Message.GetLocationMessage() != nil:
		locMsg := evt.Message.GetLocationMessage()
		message.MessageType = "locationMessage"
		// Extract location data if needed
		_ = locMsg
	case evt.Message.GetContactMessage() != nil:
		contactMsg := evt.Message.GetContactMessage()
		message.MessageType = "contactMessage"
		// Extract contact data if needed
		_ = contactMsg
	}

	// Save message to database
	// Implementation depends on repository

	// Send webhook event
	ms.sendWebhookEvent("messages.upsert", message)
}

// ProcessMessageStatus processes message status updates
func (ms *MessageService) ProcessMessageStatus(ctx context.Context, evt *events.MessageStatus) {
	ms.logger.Info("Processing message status",
		zap.String("id", evt.Info.ID),
		zap.String("status", string(evt.Status)))

	// Update message status in database
	// Implementation depends on repository

	// Send webhook event
	ms.sendWebhookEvent("messages.update", map[string]interface{}{
		"id":     evt.Info.ID,
		"status": evt.Status,
	})
}

// getMediaMessageType returns the message type for media
func (ms *MessageService) getMediaMessageType(mediaType string) string {
	switch mediaType {
	case "image":
		return "imageMessage"
	case "video":
		return "videoMessage"
	case "audio":
		return "audioMessage"
	case "document":
		return "documentMessage"
	default:
		return "unknown"
	}
}

// sendWebhookEvent sends a webhook event
func (ms *MessageService) sendWebhookEvent(event string, data interface{}) {
	ms.logger.Debug("Sending webhook event",
		zap.String("event", event))

	// Implementation will be in webhook service
	// This is a placeholder for webhook sending logic
}

// downloadMedia downloads media from URL (helper function)
func (ms *MessageService) downloadMedia(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download media: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read media data: %w", err)
	}

	return data, nil
}
