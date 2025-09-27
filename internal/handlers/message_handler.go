package handlers

import (
	"net/http"

	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MessageHandler handles message-related HTTP requests
type MessageHandler struct {
	instanceManager *services.InstanceManager
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(instanceManager *services.InstanceManager) *MessageHandler {
	return &MessageHandler{
		instanceManager: instanceManager,
	}
}

// SendText sends a text message
func (h *MessageHandler) SendText(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	var req struct {
		To   string `json:"to" binding:"required"`
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	messageReq := &services.SendMessageRequest{
		To:   req.To,
		Type: "text",
		Text: req.Text,
	}

	resp, err := h.instanceManager.SendMessage(c.Request.Context(), instanceName, messageReq)
	if err != nil {
		zap.L().Error("Failed to send text message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// SendImage sends an image message
func (h *MessageHandler) SendImage(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	var req struct {
		To       string `json:"to" binding:"required"`
		MediaURL string `json:"mediaUrl" binding:"required"`
		Caption  string `json:"caption"`
		ViewOnce bool   `json:"viewOnce"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	messageReq := &services.SendMessageRequest{
		To:       req.To,
		Type:     "image",
		MediaURL: req.MediaURL,
		Caption:  req.Caption,
		ViewOnce: req.ViewOnce,
	}

	resp, err := h.instanceManager.SendMessage(c.Request.Context(), instanceName, messageReq)
	if err != nil {
		zap.L().Error("Failed to send image message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// SendVideo sends a video message
func (h *MessageHandler) SendVideo(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	var req struct {
		To       string `json:"to" binding:"required"`
		MediaURL string `json:"mediaUrl" binding:"required"`
		Caption  string `json:"caption"`
		ViewOnce bool   `json:"viewOnce"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	messageReq := &services.SendMessageRequest{
		To:       req.To,
		Type:     "video",
		MediaURL: req.MediaURL,
		Caption:  req.Caption,
		ViewOnce: req.ViewOnce,
	}

	resp, err := h.instanceManager.SendMessage(c.Request.Context(), instanceName, messageReq)
	if err != nil {
		zap.L().Error("Failed to send video message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// SendAudio sends an audio message
func (h *MessageHandler) SendAudio(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	var req struct {
		To       string `json:"to" binding:"required"`
		MediaURL string `json:"mediaUrl" binding:"required"`
		ViewOnce bool   `json:"viewOnce"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	messageReq := &services.SendMessageRequest{
		To:       req.To,
		Type:     "audio",
		MediaURL: req.MediaURL,
		ViewOnce: req.ViewOnce,
	}

	resp, err := h.instanceManager.SendMessage(c.Request.Context(), instanceName, messageReq)
	if err != nil {
		zap.L().Error("Failed to send audio message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// SendDocument sends a document message
func (h *MessageHandler) SendDocument(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	var req struct {
		To       string `json:"to" binding:"required"`
		MediaURL string `json:"mediaUrl" binding:"required"`
		Caption  string `json:"caption"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	messageReq := &services.SendMessageRequest{
		To:       req.To,
		Type:     "document",
		MediaURL: req.MediaURL,
		Caption:  req.Caption,
	}

	resp, err := h.instanceManager.SendMessage(c.Request.Context(), instanceName, messageReq)
	if err != nil {
		zap.L().Error("Failed to send document message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}
