package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
)

// WhatsAppClient represents a WhatsApp client instance
type WhatsAppClient struct {
	client   *whatsmeow.Client
	instance *models.Instance
	store    *sqlstore.Device
	qrCode   string
	status   string
	eventCh  chan *events.Event
}

// NewWhatsAppClient creates a new WhatsApp client
func NewWhatsAppClient(instance *models.Instance, container *sqlstore.Container) (*WhatsAppClient, error) {
	// Create device store for this instance
	device, err := container.GetDevice(instance.ID)
	if err != nil {
		// Create new device if doesn't exist
		device = container.NewDevice()
		device.ID = instance.ID
	}

	// Create logger
	logger := waLog.Stdout("Client", "INFO", false)

	// Create WhatsApp client
	client := whatsmeow.NewClient(device, logger)

	wc := &WhatsAppClient{
		client:   client,
		instance: instance,
		store:    device,
		status:   "close",
		eventCh:  make(chan *events.Event, 100),
	}

	// Add event handler
	client.AddEventHandler(wc.eventHandler)

	return wc, nil
}

// Connect connects to WhatsApp
func (wc *WhatsAppClient) Connect(ctx context.Context) error {
	wc.status = "connecting"
	zap.L().Info("Connecting WhatsApp client", zap.String("instance", wc.instance.Name))

	// Check if already logged in
	if wc.client.IsLoggedIn() {
		wc.status = "open"
		return wc.client.Connect()
	}

	// Generate QR code if not logged in
	qrChan, err := wc.client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get QR channel: %w", err)
	}

	// Connect client
	if err := wc.client.Connect(); err != nil {
		wc.status = "close"
		return fmt.Errorf("failed to connect client: %w", err)
	}

	// Wait for QR code
	select {
	case evt := <-qrChan:
		if evt.Event == "code" {
			wc.qrCode = evt.Code
			zap.L().Info("QR code generated", zap.String("instance", wc.instance.Name))
			return nil
		}
	case <-time.After(30 * time.Second):
		return fmt.Errorf("QR code generation timeout")
	}

	return nil
}

// GetQRCode returns the current QR code
func (wc *WhatsAppClient) GetQRCode() string {
	return wc.qrCode
}

// GetStatus returns the connection status
func (wc *WhatsAppClient) GetStatus() string {
	if wc.client.IsConnected() && wc.client.IsLoggedIn() {
		return "open"
	}
	if wc.status == "connecting" {
		return "connecting"
	}
	return "close"
}

// SendText sends a text message
func (wc *WhatsAppClient) SendText(ctx context.Context, to, message string) (string, error) {
	if wc.status != "open" {
		return "", fmt.Errorf("client not connected")
	}

	jid, err := types.ParseJID(to)
	if err != nil {
		return "", fmt.Errorf("invalid JID: %w", err)
	}

	msg := &types.Message{
		Conversation: &message,
	}

	resp, err := wc.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return resp.ID, nil
}

// SendMedia sends a media message (image, video, audio, document)
func (wc *WhatsAppClient) SendMedia(ctx context.Context, to, mediaURL, mediaType, caption string, viewOnce bool) (string, error) {
	if wc.status != "open" {
		return "", fmt.Errorf("client not connected")
	}

	jid, err := types.ParseJID(to)
	if err != nil {
		return "", fmt.Errorf("invalid JID: %w", err)
	}

	// Download media from URL
	mediaData, err := wc.downloadMedia(mediaURL)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %w", err)
	}

	var msg *types.Message

	switch mediaType {
	case "image":
		imageMsg := &types.ImageMessage{
			Caption:  &caption,
			ViewOnce: viewOnce,
		}
		imageMsg.URL, err = wc.client.Upload(ctx, mediaData, whatsmeow.MediaImage)
		if err != nil {
			return "", fmt.Errorf("failed to upload image: %w", err)
		}
		msg = &types.Message{ImageMessage: imageMsg}

	case "video":
		videoMsg := &types.VideoMessage{
			Caption:  &caption,
			ViewOnce: viewOnce,
		}
		videoMsg.URL, err = wc.client.Upload(ctx, mediaData, whatsmeow.MediaVideo)
		if err != nil {
			return "", fmt.Errorf("failed to upload video: %w", err)
		}
		msg = &types.Message{VideoMessage: videoMsg}

	case "audio":
		audioMsg := &types.AudioMessage{
			ViewOnce: viewOnce,
		}
		audioMsg.URL, err = wc.client.Upload(ctx, mediaData, whatsmeow.MediaAudio)
		if err != nil {
			return "", fmt.Errorf("failed to upload audio: %w", err)
		}
		msg = &types.Message{AudioMessage: audioMsg}

	case "document":
		docMsg := &types.DocumentMessage{
			Caption: &caption,
		}
		docMsg.URL, err = wc.client.Upload(ctx, mediaData, whatsmeow.MediaDocument)
		if err != nil {
			return "", fmt.Errorf("failed to upload document: %w", err)
		}
		msg = &types.Message{DocumentMessage: docMsg}

	default:
		return "", fmt.Errorf("unsupported media type: %s", mediaType)
	}

	resp, err := wc.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send media message: %w", err)
	}

	return resp.ID, nil
}

// Logout logs out the client
func (wc *WhatsAppClient) Logout(ctx context.Context) error {
	wc.status = "close"
	return wc.client.Logout(ctx)
}

// Disconnect disconnects the client
func (wc *WhatsAppClient) Disconnect() {
	wc.status = "close"
	wc.client.Disconnect()
}

// downloadMedia downloads media from URL
func (wc *WhatsAppClient) downloadMedia(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download media: status %d", resp.StatusCode)
	}

	// Read response body
	mediaData := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(mediaData)
	if err != nil {
		return nil, err
	}

	return mediaData, nil
}

// eventHandler handles WhatsApp events
func (wc *WhatsAppClient) eventHandler(evt interface{}) {
	switch e := evt.(type) {
	case *events.Connected:
		wc.status = "open"
		zap.L().Info("WhatsApp client connected", zap.String("instance", wc.instance.Name))

	case *events.Disconnected:
		wc.status = "close"
		zap.L().Info("WhatsApp client disconnected", zap.String("instance", wc.instance.Name))

	case *events.Message:
		wc.handleMessage(e)

	case *events.MessageStatus:
		wc.handleMessageStatus(e)

	case *events.Presence:
		wc.handlePresence(e)

	case *events.ChatPresence:
		wc.handleChatPresence(e)
	}
}

// handleMessage handles incoming messages
func (wc *WhatsAppClient) handleMessage(evt *events.Message) {
	message := &models.Message{
		ID:          evt.Info.ID,
		InstanceID:  wc.instance.ID,
		RemoteJID:   evt.Info.Chat.String(),
		FromMe:      evt.Info.IsFromMe,
		MessageType: wc.getMessageType(evt.Message),
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
		message.MediaURL = evt.Message.GetImageMessage().GetURL()
		message.Caption = evt.Message.GetImageMessage().GetCaption()
	case evt.Message.GetVideoMessage() != nil:
		message.MediaURL = evt.Message.GetVideoMessage().GetURL()
		message.Caption = evt.Message.GetVideoMessage().GetCaption()
	case evt.Message.GetAudioMessage() != nil:
		message.MediaURL = evt.Message.GetAudioMessage().GetURL()
	case evt.Message.GetDocumentMessage() != nil:
		message.MediaURL = evt.Message.GetDocumentMessage().GetURL()
		message.Caption = evt.Message.GetDocumentMessage().GetCaption()
		message.FileName = evt.Message.GetDocumentMessage().GetFileName()
	}

	// Send webhook event
	wc.sendWebhookEvent("messages.upsert", message)
}

// handleMessageStatus handles message status updates
func (wc *WhatsAppClient) handleMessageStatus(evt *events.MessageStatus) {
	// Send webhook event for message status
	wc.sendWebhookEvent("messages.update", map[string]interface{}{
		"id":     evt.Info.ID,
		"status": evt.Status,
	})
}

// handlePresence handles presence updates
func (wc *WhatsAppClient) handlePresence(evt *events.Presence) {
	wc.sendWebhookEvent("presence.update", map[string]interface{}{
		"jid":      evt.From.String(),
		"presence": evt.Presence,
	})
}

// handleChatPresence handles chat presence updates
func (wc *WhatsAppClient) handleChatPresence(evt *events.ChatPresence) {
	wc.sendWebhookEvent("chat.presence", map[string]interface{}{
		"jid":      evt.MessageSource.Chat.String(),
		"presence": evt.State,
	})
}

// getMessageType extracts message type from WhatsApp message
func (wc *WhatsAppClient) getMessageType(msg *types.Message) string {
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

// sendWebhookEvent sends a webhook event
func (wc *WhatsAppClient) sendWebhookEvent(event string, data interface{}) {
	if wc.instance.WebhookURL == "" {
		return
	}

	webhookEvent := &models.WebhookEvent{
		Event:     event,
		Instance:  wc.instance.Name,
		Data:      data,
		Timestamp: time.Now(),
	}

	// Send webhook asynchronously
	go func() {
		// Implementation will be in webhook service
		zap.L().Debug("Sending webhook event",
			zap.String("instance", wc.instance.Name),
			zap.String("event", event))
	}()
}
