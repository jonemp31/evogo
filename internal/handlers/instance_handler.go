package handlers

import (
	"net/http"

	"github.com/evolution-api/evolution-go/internal/models"
	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InstanceHandler handles instance-related HTTP requests
type InstanceHandler struct {
	instanceManager *services.InstanceManager
	webhookService  *services.WebhookService
}

// NewInstanceHandler creates a new instance handler
func NewInstanceHandler(instanceManager *services.InstanceManager, webhookService *services.WebhookService) *InstanceHandler {
	return &InstanceHandler{
		instanceManager: instanceManager,
		webhookService:  webhookService,
	}
}

// CreateInstance creates a new instance
func (h *InstanceHandler) CreateInstance(c *gin.Context) {
	var req services.CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.instanceManager.CreateInstance(c.Request.Context(), &req)
	if err != nil {
		zap.L().Error("Failed to create instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create instance",
			"details": err.Error(),
		})
		return
	}

	// Send webhook event
	h.webhookService.SendInstanceEvent(c.Request.Context(), resp.Instance, "instance.create", resp.Instance)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    resp,
	})
}

// ConnectInstance connects an instance to WhatsApp
func (h *InstanceHandler) ConnectInstance(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	resp, err := h.instanceManager.ConnectInstance(c.Request.Context(), instanceName)
	if err != nil {
		zap.L().Error("Failed to connect instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to connect instance",
			"details": err.Error(),
		})
		return
	}

	// Send webhook event for connection status
	if resp.Connected {
		h.webhookService.SendConnectionEvent(c.Request.Context(), &models.Instance{Name: instanceName}, "open")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// GetInstanceStatus returns the status of an instance
func (h *InstanceHandler) GetInstanceStatus(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	status, err := h.instanceManager.GetInstanceStatus(instanceName)
	if err != nil {
		zap.L().Error("Failed to get instance status", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Instance not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// ListInstances returns all instances
func (h *InstanceHandler) ListInstances(c *gin.Context) {
	instances, err := h.instanceManager.ListInstances()
	if err != nil {
		zap.L().Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list instances",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    instances,
	})
}

// DeleteInstance deletes an instance
func (h *InstanceHandler) DeleteInstance(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	err := h.instanceManager.DeleteInstance(c.Request.Context(), instanceName)
	if err != nil {
		zap.L().Error("Failed to delete instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete instance",
			"details": err.Error(),
		})
		return
	}

	// Send webhook event
	h.webhookService.SendInstanceEvent(c.Request.Context(), &models.Instance{Name: instanceName}, "instance.delete", nil)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Instance deleted successfully",
	})
}

// LogoutInstance logs out an instance
func (h *InstanceHandler) LogoutInstance(c *gin.Context) {
	instanceName := c.Param("name")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Instance name is required",
		})
		return
	}

	err := h.instanceManager.LogoutInstance(c.Request.Context(), instanceName)
	if err != nil {
		zap.L().Error("Failed to logout instance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to logout instance",
			"details": err.Error(),
		})
		return
	}

	// Send webhook event
	h.webhookService.SendInstanceEvent(c.Request.Context(), &models.Instance{Name: instanceName}, "instance.logout", nil)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Instance logged out successfully",
	})
}
