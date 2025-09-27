package services

import (
	"context"
	"fmt"
	"io"
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

// WhatsAppClient representa um cliente WhatsApp com gerenciamento completo de eventos
type WhatsAppClient struct {
	client         *whatsmeow.Client
	instance       *models.Instance
	store          *sqlstore.Device
	qrCode         string
	status         string
	webhookService *WebhookService
	logger         *zap.Logger
}

// NewWhatsAppClient cria um novo cliente WhatsApp com handlers de eventos
func NewWhatsAppClient(instance *models.Instance, container *sqlstore.Container, webhookService *WebhookService) (*WhatsAppClient, error) {
	// Criar device store para esta instância
	device := container.NewDevice()
	device.ID = instance.ID

	// Criar logger
	logger := waLog.Stdout("Client", "INFO", false)

	// Criar cliente WhatsApp
	client := whatsmeow.NewClient(device, logger)

	wc := &WhatsAppClient{
		client:         client,
		instance:       instance,
		store:          device,
		status:         "close",
		webhookService: webhookService,
		logger:         zap.L().Named("WhatsAppClient").With(zap.String("instance", instance.Name)),
	}

	// Registrar handler de eventos
	client.AddEventHandler(wc.eventHandler)

	wc.logger.Info("WhatsApp client created successfully")
	return wc, nil
}

// Connect conecta ao WhatsApp
func (wc *WhatsAppClient) Connect(ctx context.Context) error {
	wc.status = "connecting"
	wc.logger.Info("Connecting WhatsApp client")

	// Verificar se já está logado
	if wc.client.IsLoggedIn() {
		wc.status = "open"
		err := wc.client.Connect()
		if err == nil {
			wc.sendConnectionWebhook("open")
		}
		return err
	}

	// Gerar QR code se não estiver logado
	qrChan, err := wc.client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get QR channel: %w", err)
	}

	// Conectar cliente
	if err := wc.client.Connect(); err != nil {
		wc.status = "close"
		wc.sendConnectionWebhook("close")
		return fmt.Errorf("failed to connect client: %w", err)
	}

	// Aguardar QR code
	go func() {
		select {
		case evt := <-qrChan:
			if evt.Event == "code" {
				wc.qrCode = evt.Code
				wc.logger.Info("QR code generated")
				wc.sendConnectionWebhook("qrcode")
			}
		case <-time.After(30 * time.Second):
			wc.logger.Warn("QR code generation timeout")
		}
	}()

	return nil
}

// GetQRCode retorna o QR code atual
func (wc *WhatsAppClient) GetQRCode() string {
	return wc.qrCode
}

// GetStatus retorna o status da conexão
func (wc *WhatsAppClient) GetStatus() string {
	if wc.client.IsConnected() && wc.client.IsLoggedIn() {
		return "open"
	}
	if wc.status == "connecting" {
		return "connecting"
	}
	return "close"
}

// SendText envia uma mensagem de texto
func (wc *WhatsAppClient) SendText(ctx context.Context, to, message string) (string, error) {
	if wc.GetStatus() != "open" {
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

	wc.logger.Info("Text message sent", zap.String("to", to), zap.String("messageId", resp.ID))
	return resp.ID, nil
}

// SendMedia envia uma mídia (imagem, vídeo, áudio, documento)
func (wc *WhatsAppClient) SendMedia(ctx context.Context, to, mediaURL, mediaType, caption string, viewOnce, ptt bool) (string, error) {
	if wc.GetStatus() != "open" {
		return "", fmt.Errorf("client not connected")
	}

	jid, err := types.ParseJID(to)
	if err != nil {
		return "", fmt.Errorf("invalid JID: %w", err)
	}

	// Baixar mídia da URL
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
			Ptt:      ptt,
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

	wc.logger.Info("Media message sent",
		zap.String("to", to),
		zap.String("mediaType", mediaType),
		zap.String("messageId", resp.ID),
		zap.Bool("viewOnce", viewOnce),
		zap.Bool("ptt", ptt))

	return resp.ID, nil
}

// Logout faz logout do cliente
func (wc *WhatsAppClient) Logout(ctx context.Context) error {
	wc.status = "close"
	err := wc.client.Logout(ctx)
	if err == nil {
		wc.sendConnectionWebhook("close")
	}
	return err
}

// Disconnect desconecta o cliente
func (wc *WhatsAppClient) Disconnect() {
	wc.status = "close"
	wc.client.Disconnect()
	wc.sendConnectionWebhook("close")
}

// eventHandler é o handler principal de eventos do WhatsApp
func (wc *WhatsAppClient) eventHandler(evt interface{}) {
	switch e := evt.(type) {
	case *events.Connected:
		wc.handleConnected(e)
	case *events.Disconnected:
		wc.handleDisconnected(e)
	case *events.Message:
		wc.handleMessage(e)
	case *events.MessageStatus:
		wc.handleMessageStatus(e)
	case *events.Presence:
		wc.handlePresence(e)
	case *events.ChatPresence:
		wc.handleChatPresence(e)
	default:
		wc.logger.Debug("Unhandled event", zap.String("type", fmt.Sprintf("%T", e)))
	}
}

// handleConnected processa eventos de conexão
func (wc *WhatsAppClient) handleConnected(evt *events.Connected) {
	wc.status = "open"
	wc.logger.Info("WhatsApp client connected successfully")
	wc.sendConnectionWebhook("open")
}

// handleDisconnected processa eventos de desconexão
func (wc *WhatsAppClient) handleDisconnected(evt *events.Disconnected) {
	wc.status = "close"
	wc.logger.Info("WhatsApp client disconnected", zap.String("reason", evt.Reason.String()))
	wc.sendConnectionWebhook("close")
}

// handleMessage processa mensagens recebidas
func (wc *WhatsAppClient) handleMessage(evt *events.Message) {
	wc.logger.Info("Processing incoming message",
		zap.String("id", evt.Info.ID),
		zap.String("chat", evt.Info.Chat.String()),
		zap.Bool("fromMe", evt.Info.IsFromMe))

	// Criar objeto de mensagem
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

	// Extrair conteúdo da mensagem baseado no tipo
	wc.extractMessageContent(evt.Message, message)

	// Enviar webhook de mensagem recebida
	wc.sendMessageWebhook("messages.upsert", message)
}

// handleMessageStatus processa atualizações de status de mensagem
func (wc *WhatsAppClient) handleMessageStatus(evt *events.MessageStatus) {
	wc.logger.Info("Processing message status update",
		zap.String("id", evt.Info.ID),
		zap.String("status", string(evt.Status)))

	// Enviar webhook de atualização de status
	wc.sendMessageWebhook("messages.update", map[string]interface{}{
		"id":     evt.Info.ID,
		"status": string(evt.Status),
	})
}

// handlePresence processa eventos de presença
func (wc *WhatsAppClient) handlePresence(evt *events.Presence) {
	wc.sendWebhookEvent("presence.update", map[string]interface{}{
		"jid":      evt.From.String(),
		"presence": evt.Presence,
	})
}

// handleChatPresence processa eventos de presença no chat
func (wc *WhatsAppClient) handleChatPresence(evt *events.ChatPresence) {
	wc.sendWebhookEvent("chat.presence", map[string]interface{}{
		"jid":      evt.MessageSource.Chat.String(),
		"presence": evt.State,
	})
}

// extractMessageContent extrai o conteúdo da mensagem baseado no tipo
func (wc *WhatsAppClient) extractMessageContent(msg *types.Message, message *models.Message) {
	switch {
	case msg.GetConversation() != "":
		message.Message = msg.GetConversation()
	case msg.GetExtendedTextMessage() != nil:
		message.Message = msg.GetExtendedTextMessage().GetText()
	case msg.GetImageMessage() != nil:
		imgMsg := msg.GetImageMessage()
		message.MediaURL = imgMsg.GetURL()
		message.Caption = imgMsg.GetCaption()
		message.ViewOnce = imgMsg.GetViewOnce()
	case msg.GetVideoMessage() != nil:
		vidMsg := msg.GetVideoMessage()
		message.MediaURL = vidMsg.GetURL()
		message.Caption = vidMsg.GetCaption()
		message.ViewOnce = vidMsg.GetViewOnce()
	case msg.GetAudioMessage() != nil:
		audioMsg := msg.GetAudioMessage()
		message.MediaURL = audioMsg.GetURL()
		message.ViewOnce = audioMsg.GetViewOnce()
		message.Ptt = audioMsg.GetPtt()
	case msg.GetDocumentMessage() != nil:
		docMsg := msg.GetDocumentMessage()
		message.MediaURL = docMsg.GetURL()
		message.Caption = docMsg.GetCaption()
		message.FileName = docMsg.GetFileName()
	case msg.GetStickerMessage() != nil:
		stickerMsg := msg.GetStickerMessage()
		message.MediaURL = stickerMsg.GetURL()
	case msg.GetLocationMessage() != nil:
		message.MessageType = "locationMessage"
		// Extrair dados de localização se necessário
	case msg.GetContactMessage() != nil:
		message.MessageType = "contactMessage"
		// Extrair dados de contato se necessário
	}
}

// getMessageType extrai o tipo da mensagem
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

// downloadMedia baixa mídia de uma URL
func (wc *WhatsAppClient) downloadMedia(url string) ([]byte, error) {
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

// sendConnectionWebhook envia webhook de status de conexão
func (wc *WhatsAppClient) sendConnectionWebhook(status string) {
	if wc.webhookService == nil {
		return
	}

	payload := map[string]interface{}{
		"event":    "connection.update",
		"instance": wc.instance.Name,
		"data": map[string]interface{}{
			"state": status,
		},
	}

	wc.webhookService.SendWebhook(wc.instance.WebhookURL, payload)
}

// sendMessageWebhook envia webhook de mensagem
func (wc *WhatsAppClient) sendMessageWebhook(event string, message *models.Message) {
	if wc.webhookService == nil {
		return
	}

	payload := map[string]interface{}{
		"event":    event,
		"instance": wc.instance.Name,
		"data":     message,
	}

	wc.webhookService.SendWebhook(wc.instance.WebhookURL, payload)
}

// sendWebhookEvent envia webhook genérico
func (wc *WhatsAppClient) sendWebhookEvent(event string, data interface{}) {
	if wc.webhookService == nil || wc.instance.WebhookURL == "" {
		return
	}

	payload := map[string]interface{}{
		"event":    event,
		"instance": wc.instance.Name,
		"data":     data,
	}

	wc.webhookService.SendWebhook(wc.instance.WebhookURL, payload)
}
