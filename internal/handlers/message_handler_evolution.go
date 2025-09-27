package handlers

import (
	"net/http"
	"time"

	"github.com/evolution-api/evolution-go/internal/dto"
	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MessageHandler gerencia os endpoints de mensagens
type MessageHandler struct {
	messageService *services.MessageService
	logger         *zap.Logger
}

// NewMessageHandler cria um novo handler de mensagens
func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		logger:         zap.L().Named("MessageHandler"),
	}
}

// SendText envia uma mensagem de texto
// POST /message/sendText/{instanceName}
func (mh *MessageHandler) SendText(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.SendTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		mh.logger.Error("Failed to bind send text request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar número
	if req.Number == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_NUMBER",
			Message:   "Number is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar texto
	if req.Text == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_TEXT",
			Message:   "Text is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Enviar mensagem
	messageID, err := mh.messageService.SendText(c.Request.Context(), instanceName, req.Number, req.Text)
	if err != nil {
		mh.logger.Error("Failed to send text message",
			zap.String("instance", instanceName),
			zap.String("number", req.Number),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "SEND_FAILED",
			Message:   "Failed to send message: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Success:   true,
		MessageID: messageID,
		Instance:  instanceName,
		Timestamp: time.Now(),
	})
}

// SendMedia envia uma mídia (imagem, vídeo, áudio, documento)
// POST /message/sendMedia/{instanceName}
func (mh *MessageHandler) SendMedia(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.SendMediaRequest
	if err := c.ShouldBind(&req); err != nil {
		mh.logger.Error("Failed to bind send media request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar número
	if req.Number == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_NUMBER",
			Message:   "Number is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar URL da mídia
	if req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_MEDIA_URL",
			Message:   "Media URL is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar tipo de mídia
	validMediaTypes := map[string]bool{
		"image":    true,
		"video":    true,
		"audio":    true,
		"document": true,
	}
	if !validMediaTypes[req.MediaType] {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_MEDIA_TYPE",
			Message:   "Media type must be one of: image, video, audio, document",
			Timestamp: time.Now(),
		})
		return
	}

	// Enviar mídia
	messageID, err := mh.messageService.SendMedia(
		c.Request.Context(),
		instanceName,
		req.Number,
		req.MediaURL,
		req.MediaType,
		req.Caption,
		req.Options.ViewOnce,
		req.Options.Ptt,
	)
	if err != nil {
		mh.logger.Error("Failed to send media message",
			zap.String("instance", instanceName),
			zap.String("number", req.Number),
			zap.String("mediaType", req.MediaType),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "SEND_FAILED",
			Message:   "Failed to send media: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Success:   true,
		MessageID: messageID,
		Instance:  instanceName,
		Timestamp: time.Now(),
	})
}

// SendImage envia uma imagem
// POST /message/sendImage/{instanceName}
func (mh *MessageHandler) SendImage(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.SendMediaRequest
	if err := c.ShouldBind(&req); err != nil {
		mh.logger.Error("Failed to bind send image request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Forçar tipo de mídia para imagem
	req.MediaType = "image"

	// Validar número
	if req.Number == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_NUMBER",
			Message:   "Number is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar URL da mídia
	if req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_MEDIA_URL",
			Message:   "Media URL is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Enviar imagem
	messageID, err := mh.messageService.SendMedia(
		c.Request.Context(),
		instanceName,
		req.Number,
		req.MediaURL,
		req.MediaType,
		req.Caption,
		req.Options.ViewOnce,
		false, // PTT não se aplica a imagens
	)
	if err != nil {
		mh.logger.Error("Failed to send image",
			zap.String("instance", instanceName),
			zap.String("number", req.Number),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "SEND_FAILED",
			Message:   "Failed to send image: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Success:   true,
		MessageID: messageID,
		Instance:  instanceName,
		Timestamp: time.Now(),
	})
}

// SendVideo envia um vídeo
// POST /message/sendVideo/{instanceName}
func (mh *MessageHandler) SendVideo(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.SendMediaRequest
	if err := c.ShouldBind(&req); err != nil {
		mh.logger.Error("Failed to bind send video request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Forçar tipo de mídia para vídeo
	req.MediaType = "video"

	// Validar número
	if req.Number == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_NUMBER",
			Message:   "Number is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar URL da mídia
	if req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_MEDIA_URL",
			Message:   "Media URL is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Enviar vídeo
	messageID, err := mh.messageService.SendMedia(
		c.Request.Context(),
		instanceName,
		req.Number,
		req.MediaURL,
		req.MediaType,
		req.Caption,
		req.Options.ViewOnce,
		false, // PTT não se aplica a vídeos
	)
	if err != nil {
		mh.logger.Error("Failed to send video",
			zap.String("instance", instanceName),
			zap.String("number", req.Number),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "SEND_FAILED",
			Message:   "Failed to send video: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Success:   true,
		MessageID: messageID,
		Instance:  instanceName,
		Timestamp: time.Now(),
	})
}

// SendAudio envia um áudio
// POST /message/sendAudio/{instanceName}
func (mh *MessageHandler) SendAudio(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.SendMediaRequest
	if err := c.ShouldBind(&req); err != nil {
		mh.logger.Error("Failed to bind send audio request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Forçar tipo de mídia para áudio
	req.MediaType = "audio"

	// Validar número
	if req.Number == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_NUMBER",
			Message:   "Number is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar URL da mídia
	if req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_MEDIA_URL",
			Message:   "Media URL is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Enviar áudio
	messageID, err := mh.messageService.SendMedia(
		c.Request.Context(),
		instanceName,
		req.Number,
		req.MediaURL,
		req.MediaType,
		req.Caption,
		req.Options.ViewOnce,
		req.Options.Ptt, // PTT se aplica a áudios
	)
	if err != nil {
		mh.logger.Error("Failed to send audio",
			zap.String("instance", instanceName),
			zap.String("number", req.Number),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "SEND_FAILED",
			Message:   "Failed to send audio: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Success:   true,
		MessageID: messageID,
		Instance:  instanceName,
		Timestamp: time.Now(),
	})
}

// SendDocument envia um documento
// POST /message/sendDocument/{instanceName}
func (mh *MessageHandler) SendDocument(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.SendMediaRequest
	if err := c.ShouldBind(&req); err != nil {
		mh.logger.Error("Failed to bind send document request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Forçar tipo de mídia para documento
	req.MediaType = "document"

	// Validar número
	if req.Number == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_NUMBER",
			Message:   "Number is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar URL da mídia
	if req.MediaURL == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_MEDIA_URL",
			Message:   "Media URL is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Enviar documento
	messageID, err := mh.messageService.SendMedia(
		c.Request.Context(),
		instanceName,
		req.Number,
		req.MediaURL,
		req.MediaType,
		req.Caption,
		false, // ViewOnce não se aplica a documentos
		false, // PTT não se aplica a documentos
	)
	if err != nil {
		mh.logger.Error("Failed to send document",
			zap.String("instance", instanceName),
			zap.String("number", req.Number),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "SEND_FAILED",
			Message:   "Failed to send document: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Success:   true,
		MessageID: messageID,
		Instance:  instanceName,
		Timestamp: time.Now(),
	})
}

// GetMessageStatus obtém o status de uma mensagem
// GET /message/status/{instanceName}/{messageId}
func (mh *MessageHandler) GetMessageStatus(c *gin.Context) {
	instanceName := c.Param("instanceName")
	messageID := c.Param("messageId")

	if instanceName == "" || messageID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_PARAMS",
			Message:   "Instance name and message ID are required",
			Timestamp: time.Now(),
		})
		return
	}

	status, err := mh.messageService.GetMessageStatus(c.Request.Context(), instanceName, messageID)
	if err != nil {
		mh.logger.Error("Failed to get message status",
			zap.String("instance", instanceName),
			zap.String("messageId", messageID),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "STATUS_FAILED",
			Message:   "Failed to get message status: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"messageId": messageID,
		"status":    status,
		"instance":  instanceName,
		"timestamp": time.Now(),
	})
}
