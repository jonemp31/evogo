package handlers

import (
	"net/http"
	"time"

	"github.com/evolution-api/evolution-go/internal/dto"
	"github.com/evolution-api/evolution-go/internal/models"
	"github.com/evolution-api/evolution-go/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InstanceHandler gerencia os endpoints de instâncias
type InstanceHandler struct {
	instanceManager *services.InstanceManager
	webhookService  *services.WebhookService
	logger          *zap.Logger
}

// NewInstanceHandler cria um novo handler de instâncias
func NewInstanceHandler(instanceManager *services.InstanceManager, webhookService *services.WebhookService) *InstanceHandler {
	return &InstanceHandler{
		instanceManager: instanceManager,
		webhookService:  webhookService,
		logger:          zap.L().Named("InstanceHandler"),
	}
}

// CreateInstance cria uma nova instância
// POST /instance/create
func (ih *InstanceHandler) CreateInstance(c *gin.Context) {
	var req dto.CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ih.logger.Error("Failed to bind create instance request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Validar nome da instância
	if req.InstanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Criar instância
	instance, err := ih.instanceManager.CreateInstance(c.Request.Context(), req.InstanceName, req.WebhookURL)
	if err != nil {
		ih.logger.Error("Failed to create instance",
			zap.String("instanceName", req.InstanceName),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "CREATE_FAILED",
			Message:   "Failed to create instance: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.InstanceStatusResponse{
		Instance: struct {
			Name       string    `json:"name"`
			Status     string    `json:"status"`
			WebhookURL string    `json:"webhookUrl"`
			CreatedAt  time.Time `json:"createdAt"`
			UpdatedAt  time.Time `json:"updatedAt"`
		}{
			Name:       instance.Name,
			Status:     instance.Status,
			WebhookURL: instance.WebhookURL,
			CreatedAt:  instance.CreatedAt,
			UpdatedAt:  instance.UpdatedAt,
		},
	})
}

// ConnectInstance conecta uma instância
// POST /instance/connect/{instanceName}
func (ih *InstanceHandler) ConnectInstance(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Conectar instância
	err := ih.instanceManager.ConnectInstance(c.Request.Context(), instanceName)
	if err != nil {
		ih.logger.Error("Failed to connect instance",
			zap.String("instanceName", instanceName),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "CONNECT_FAILED",
			Message:   "Failed to connect instance: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ConnectionStatusResponse{
		Instance: instanceName,
		Status:   "connecting",
		Message:  "Instance connection initiated",
	})
}

// GetInstanceStatus obtém o status de uma instância
// GET /instance/status/{instanceName}
func (ih *InstanceHandler) GetInstanceStatus(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Obter status da instância
	instance, err := ih.instanceManager.GetInstanceStatus(c.Request.Context(), instanceName)
	if err != nil {
		ih.logger.Error("Failed to get instance status",
			zap.String("instanceName", instanceName),
			zap.Error(err))

		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Success:   false,
			Error:     "INSTANCE_NOT_FOUND",
			Message:   "Instance not found: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.InstanceStatusResponse{
		Instance: struct {
			Name       string    `json:"name"`
			Status     string    `json:"status"`
			WebhookURL string    `json:"webhookUrl"`
			CreatedAt  time.Time `json:"createdAt"`
			UpdatedAt  time.Time `json:"updatedAt"`
		}{
			Name:       instance.Name,
			Status:     instance.Status,
			WebhookURL: instance.WebhookURL,
			CreatedAt:  instance.CreatedAt,
			UpdatedAt:  instance.UpdatedAt,
		},
	})
}

// ListInstances lista todas as instâncias
// GET /instance/list
func (ih *InstanceHandler) ListInstances(c *gin.Context) {
	// Listar instâncias
	instances, err := ih.instanceManager.ListInstances(c.Request.Context())
	if err != nil {
		ih.logger.Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "LIST_FAILED",
			Message:   "Failed to list instances: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// Converter para formato de resposta
	var instanceList []dto.InstanceStatusResponse
	for _, instance := range instances {
		instanceList = append(instanceList, dto.InstanceStatusResponse{
			Instance: struct {
				Name       string    `json:"name"`
				Status     string    `json:"status"`
				WebhookURL string    `json:"webhookUrl"`
				CreatedAt  time.Time `json:"createdAt"`
				UpdatedAt  time.Time `json:"updatedAt"`
			}{
				Name:       instance.Name,
				Status:     instance.Status,
				WebhookURL: instance.WebhookURL,
				CreatedAt:  instance.CreatedAt,
				UpdatedAt:  instance.UpdatedAt,
			},
		})
	}

	c.JSON(http.StatusOK, dto.InstanceListResponse{
		Instances: instanceList,
		Total:     len(instanceList),
	})
}

// DeleteInstance deleta uma instância
// DELETE /instance/delete/{instanceName}
func (ih *InstanceHandler) DeleteInstance(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Deletar instância
	err := ih.instanceManager.DeleteInstance(c.Request.Context(), instanceName)
	if err != nil {
		ih.logger.Error("Failed to delete instance",
			zap.String("instanceName", instanceName),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "DELETE_FAILED",
			Message:   "Failed to delete instance: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "Instance deleted successfully",
		"instance":  instanceName,
		"timestamp": time.Now(),
	})
}

// LogoutInstance faz logout de uma instância
// POST /instance/logout/{instanceName}
func (ih *InstanceHandler) LogoutInstance(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Fazer logout da instância
	err := ih.instanceManager.LogoutInstance(c.Request.Context(), instanceName)
	if err != nil {
		ih.logger.Error("Failed to logout instance",
			zap.String("instanceName", instanceName),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "LOGOUT_FAILED",
			Message:   "Failed to logout instance: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ConnectionStatusResponse{
		Instance: instanceName,
		Status:   "close",
		Message:  "Instance logged out successfully",
	})
}

// GetQRCode obtém o QR code de uma instância
// GET /instance/qrcode/{instanceName}
func (ih *InstanceHandler) GetQRCode(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	// Obter QR code
	qrCode, err := ih.instanceManager.GetQRCode(c.Request.Context(), instanceName)
	if err != nil {
		ih.logger.Error("Failed to get QR code",
			zap.String("instanceName", instanceName),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "QRCODE_FAILED",
			Message:   "Failed to get QR code: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.QRCodeResponse{
		Instance: instanceName,
		QRCode:   qrCode,
		Status:   "qrcode",
	})
}

// UpdateInstanceWebhook atualiza a URL do webhook de uma instância
// POST /instance/webhook/{instanceName}
func (ih *InstanceHandler) UpdateInstanceWebhook(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.InstanceWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ih.logger.Error("Failed to bind webhook request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Atualizar webhook da instância
	err := ih.instanceManager.UpdateInstanceWebhook(c.Request.Context(), instanceName, req.WebhookURL)
	if err != nil {
		ih.logger.Error("Failed to update instance webhook",
			zap.String("instanceName", instanceName),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "WEBHOOK_UPDATE_FAILED",
			Message:   "Failed to update webhook: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":    true,
		"message":    "Webhook updated successfully",
		"instance":   instanceName,
		"webhookUrl": req.WebhookURL,
		"timestamp":  time.Now(),
	})
}

// UpdateInstanceSettings atualiza as configurações de uma instância
// POST /instance/settings/{instanceName}
func (ih *InstanceHandler) UpdateInstanceSettings(c *gin.Context) {
	instanceName := c.Param("instanceName")
	if instanceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_INSTANCE_NAME",
			Message:   "Instance name is required",
			Timestamp: time.Now(),
		})
		return
	}

	var req dto.InstanceSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ih.logger.Error("Failed to bind settings request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Success:   false,
			Error:     "INVALID_REQUEST",
			Message:   "Invalid request body",
			Timestamp: time.Now(),
		})
		return
	}

	// Converter para modelo de configurações
	settings := &models.InstanceSettings{
		RejectCall:           req.RejectCall,
		MsgRetryCounterCache: req.MsgRetryCounterCache,
		UserAgent:            req.UserAgent,
		AlwaysOnline:         req.AlwaysOnline,
		ReadMessages:         req.ReadMessages,
		ReadStatus:           req.ReadStatus,
		SyncFullHistory:      req.SyncFullHistory,
		MarkOnlineOnConnect:  req.MarkOnlineOnConnect,
		DefaultReactionEmoji: req.DefaultReactionEmoji,
	}

	// Atualizar configurações da instância
	err := ih.instanceManager.UpdateInstanceSettings(c.Request.Context(), instanceName, settings)
	if err != nil {
		ih.logger.Error("Failed to update instance settings",
			zap.String("instanceName", instanceName),
			zap.Error(err))

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Success:   false,
			Error:     "SETTINGS_UPDATE_FAILED",
			Message:   "Failed to update settings: " + err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "Settings updated successfully",
		"instance":  instanceName,
		"settings":  req,
		"timestamp": time.Now(),
	})
}
