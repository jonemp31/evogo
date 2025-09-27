package services

import (
	"context"
	"fmt"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
)

// MessageService gerencia operações relacionadas a mensagens
type MessageService struct {
	instanceManager *InstanceManager
	logger          *zap.Logger
}

// NewMessageService cria um novo serviço de mensagens
func NewMessageService(instanceManager *InstanceManager) *MessageService {
	return &MessageService{
		instanceManager: instanceManager,
		logger:          zap.L().Named("MessageService"),
	}
}

// SendText envia uma mensagem de texto
func (ms *MessageService) SendText(ctx context.Context, instanceName, to, message string) (string, error) {
	ms.logger.Info("Sending text message",
		zap.String("instance", instanceName),
		zap.String("to", to))

	// Obter cliente WhatsApp
	waClient := ms.instanceManager.getClient(instanceName)
	if waClient == nil {
		return "", fmt.Errorf("WhatsApp client not found for instance %s", instanceName)
	}

	// Enviar mensagem
	messageID, err := waClient.SendText(ctx, to, message)
	if err != nil {
		return "", fmt.Errorf("failed to send text message: %w", err)
	}

	// Criar registro de mensagem
	msg := &models.Message{
		ID:          messageID,
		InstanceID:  waClient.instance.ID,
		RemoteJID:   to,
		FromMe:      true,
		MessageType: "conversation",
		Message:     message,
		Status:      "sent",
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}

	// Salvar no banco de dados (implementação depende do repository)
	_ = msg

	ms.logger.Info("Text message sent successfully",
		zap.String("instance", instanceName),
		zap.String("messageId", messageID))

	return messageID, nil
}

// SendMedia envia uma mídia (imagem, vídeo, áudio, documento)
func (ms *MessageService) SendMedia(ctx context.Context, instanceName, to, mediaURL, mediaType, caption string, viewOnce, ptt bool) (string, error) {
	ms.logger.Info("Sending media message",
		zap.String("instance", instanceName),
		zap.String("to", to),
		zap.String("mediaType", mediaType),
		zap.Bool("viewOnce", viewOnce),
		zap.Bool("ptt", ptt))

	// Obter cliente WhatsApp
	waClient := ms.instanceManager.getClient(instanceName)
	if waClient == nil {
		return "", fmt.Errorf("WhatsApp client not found for instance %s", instanceName)
	}

	// Enviar mídia
	messageID, err := waClient.SendMedia(ctx, to, mediaURL, mediaType, caption, viewOnce, ptt)
	if err != nil {
		return "", fmt.Errorf("failed to send media message: %w", err)
	}

	// Criar registro de mensagem
	msg := &models.Message{
		ID:          messageID,
		InstanceID:  waClient.instance.ID,
		RemoteJID:   to,
		FromMe:      true,
		MessageType: ms.getMediaMessageType(mediaType),
		MediaURL:    mediaURL,
		Caption:     caption,
		ViewOnce:    viewOnce,
		Ptt:         ptt,
		Status:      "sent",
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}

	// Salvar no banco de dados (implementação depende do repository)
	_ = msg

	ms.logger.Info("Media message sent successfully",
		zap.String("instance", instanceName),
		zap.String("messageId", messageID),
		zap.String("mediaType", mediaType))

	return messageID, nil
}

// ProcessIncomingMessage processa mensagens recebidas do WhatsApp
func (ms *MessageService) ProcessIncomingMessage(ctx context.Context, evt *events.Message) {
	ms.logger.Info("Processing incoming message",
		zap.String("id", evt.Info.ID),
		zap.String("chat", evt.Info.Chat.String()),
		zap.Bool("fromMe", evt.Info.IsFromMe))

	// Pular mensagens próprias
	if evt.Info.IsFromMe {
		return
	}

	// Criar modelo de mensagem
	message := &models.Message{
		ID:        evt.Info.ID,
		RemoteJID: evt.Info.Chat.String(),
		FromMe:    evt.Info.IsFromMe,
		Timestamp: evt.Info.Timestamp,
		Status:    "received",
		CreatedAt: time.Now(),
	}

	// Extrair conteúdo da mensagem baseado no tipo
	ms.extractMessageContent(evt.Message, message)

	// Salvar mensagem no banco de dados
	// Implementação depende do repository

	// Enviar webhook de evento
	ms.sendWebhookEvent("messages.upsert", message)
}

// ProcessMessageStatus processa atualizações de status de mensagem
func (ms *MessageService) ProcessMessageStatus(ctx context.Context, evt *events.MessageStatus) {
	ms.logger.Info("Processing message status",
		zap.String("id", evt.Info.ID),
		zap.String("status", string(evt.Status)))

	// Atualizar status da mensagem no banco de dados
	// Implementação depende do repository

	// Enviar webhook de evento
	ms.sendWebhookEvent("messages.update", map[string]interface{}{
		"id":     evt.Info.ID,
		"status": string(evt.Status),
	})
}

// GetMessageStatus obtém o status de uma mensagem
func (ms *MessageService) GetMessageStatus(ctx context.Context, instanceName, messageID string) (string, error) {
	ms.logger.Debug("Getting message status",
		zap.String("instance", instanceName),
		zap.String("messageId", messageID))

	// Buscar status da mensagem no banco de dados
	// Implementação depende do repository

	// Por enquanto, retornar status padrão
	return "sent", nil
}

// extractMessageContent extrai o conteúdo da mensagem baseado no tipo
func (ms *MessageService) extractMessageContent(msg interface{}, message *models.Message) {
	// Esta função seria implementada com base na biblioteca go-whatsapp
	// Por enquanto, é um placeholder
	message.MessageType = "unknown"
	message.Message = "Message content extraction not implemented"
}

// getMediaMessageType retorna o tipo de mensagem para mídia
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

// sendWebhookEvent envia um evento de webhook
func (ms *MessageService) sendWebhookEvent(event string, data interface{}) {
	ms.logger.Debug("Sending webhook event",
		zap.String("event", event))

	// Implementação será no webhook service
	// Este é um placeholder para a lógica de envio de webhook
}
