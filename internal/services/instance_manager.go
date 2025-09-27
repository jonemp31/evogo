package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"github.com/evolution-api/evolution-go/internal/repository"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
)

// InstanceManager manages WhatsApp instances
type InstanceManager struct {
	instances  map[string]*WhatsAppClient
	repository *repository.InstanceRepository
	container  *sqlstore.Container
	mutex      sync.RWMutex
	eventCh    chan InstanceEvent
	shutdownCh chan struct{}
}

// InstanceEvent represents an instance event
type InstanceEvent struct {
	Type      string      `json:"type"`
	Instance  string      `json:"instance"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewInstanceManager creates a new instance manager
func NewInstanceManager(repo *repository.InstanceRepository, container *sqlstore.Container) *InstanceManager {
	im := &InstanceManager{
		instances:  make(map[string]*WhatsAppClient),
		repository: repo,
		container:  container,
		eventCh:    make(chan InstanceEvent, 1000),
		shutdownCh: make(chan struct{}),
	}

	// Start event processor
	go im.processEvents()

	// Start cleanup routine
	go im.cleanupRoutine()

	return im
}

// CreateInstance creates a new WhatsApp instance
func (im *InstanceManager) CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*CreateInstanceResponse, error) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	// Check if instance already exists
	if _, exists := im.instances[req.Name]; exists {
		return nil, fmt.Errorf("instance %s already exists", req.Name)
	}

	// Check if instance exists in database
	if _, err := im.repository.GetByName(req.Name); err == nil {
		return nil, fmt.Errorf("instance %s already exists in database", req.Name)
	}

	// Create instance model
	instance := &models.Instance{
		Name:        req.Name,
		Token:       req.Token,
		Status:      "close",
		Integration: "WHATSAPP-BAILEYS",
		Settings: models.InstanceSettings{
			RejectCall:       req.RejectCall,
			MsgCall:          req.MsgCall,
			GroupsIgnore:     req.GroupsIgnore,
			AlwaysOnline:     req.AlwaysOnline,
			ReadMessages:     req.ReadMessages,
			ReadStatus:       req.ReadStatus,
			SyncFullHistory:  req.SyncFullHistory,
			WAVoipToken:      req.WAVoipToken,
			AutoReadMessages: req.AutoReadMessages,
			ReadDelay:        req.ReadDelay,
		},
		WebhookURL:    req.WebhookURL,
		WebhookEvents: req.WebhookEvents,
		WebhookBase64: req.WebhookBase64,
	}

	// Save to database
	if err := im.repository.Create(instance); err != nil {
		return nil, fmt.Errorf("failed to create instance in database: %w", err)
	}

	// Create WhatsApp client
	client, err := NewWhatsAppClient(instance, im.container)
	if err != nil {
		// Cleanup database entry
		im.repository.Delete(instance.ID)
		return nil, fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	// Store client
	im.instances[req.Name] = client

	// Emit event
	im.emitEvent("instance.created", req.Name, instance)

	zap.L().Info("Instance created successfully", zap.String("name", req.Name))

	return &CreateInstanceResponse{
		Instance: instance,
	}, nil
}

// ConnectInstance connects an instance to WhatsApp
func (im *InstanceManager) ConnectInstance(ctx context.Context, name string) (*ConnectInstanceResponse, error) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	client, exists := im.instances[name]
	if !exists {
		// Try to load from database
		instance, err := im.repository.GetByName(name)
		if err != nil {
			return nil, fmt.Errorf("instance %s not found", name)
		}

		// Create client
		client, err = NewWhatsAppClient(instance, im.container)
		if err != nil {
			return nil, fmt.Errorf("failed to create WhatsApp client: %w", err)
		}

		im.instances[name] = client
	}

	// Check if already connected
	if client.GetStatus() == "open" {
		return &ConnectInstanceResponse{
			Connected: true,
			Message:   "Instance already connected",
		}, nil
	}

	// Connect
	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Update status in database
	im.repository.UpdateStatus(client.instance.ID, "connecting")

	qrCode := client.GetQRCode()
	if qrCode != "" {
		return &ConnectInstanceResponse{
			Connected: false,
			Message:   "QR code generated",
			QRCode: &models.QRCode{
				Code: qrCode,
			},
		}, nil
	}

	return &ConnectInstanceResponse{
		Connected: true,
		Message:   "Instance connected successfully",
	}, nil
}

// GetInstanceStatus returns the status of an instance
func (im *InstanceManager) GetInstanceStatus(name string) (*models.ConnectionStatus, error) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	client, exists := im.instances[name]
	if !exists {
		// Try to load from database
		instance, err := im.repository.GetByName(name)
		if err != nil {
			return nil, fmt.Errorf("instance %s not found", name)
		}

		return &models.ConnectionStatus{
			InstanceName: name,
			State:        instance.Status,
			RemoteJID:    instance.OwnerJID,
		}, nil
	}

	status := client.GetStatus()

	// Update last seen
	im.repository.UpdateLastSeen(client.instance.ID)

	return &models.ConnectionStatus{
		InstanceName: name,
		State:        status,
		RemoteJID:    client.instance.OwnerJID,
	}, nil
}

// ListInstances returns all instances
func (im *InstanceManager) ListInstances() ([]*models.Instance, error) {
	return im.repository.GetAll()
}

// DeleteInstance deletes an instance
func (im *InstanceManager) DeleteInstance(ctx context.Context, name string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	client, exists := im.instances[name]
	if exists {
		// Disconnect client
		client.Disconnect()
		delete(im.instances, name)
	}

	// Delete from database
	instance, err := im.repository.GetByName(name)
	if err != nil {
		return fmt.Errorf("instance %s not found", name)
	}

	if err := im.repository.Delete(instance.ID); err != nil {
		return fmt.Errorf("failed to delete instance from database: %w", err)
	}

	// Emit event
	im.emitEvent("instance.deleted", name, nil)

	zap.L().Info("Instance deleted successfully", zap.String("name", name))

	return nil
}

// LogoutInstance logs out an instance
func (im *InstanceManager) LogoutInstance(ctx context.Context, name string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	client, exists := im.instances[name]
	if !exists {
		return fmt.Errorf("instance %s not found", name)
	}

	// Logout client
	if err := client.Logout(ctx); err != nil {
		return fmt.Errorf("failed to logout client: %w", err)
	}

	// Update status in database
	im.repository.UpdateStatus(client.instance.ID, "close")

	// Emit event
	im.emitEvent("instance.logout", name, nil)

	zap.L().Info("Instance logged out successfully", zap.String("name", name))

	return nil
}

// SendMessage sends a message through an instance
func (im *InstanceManager) SendMessage(ctx context.Context, instanceName string, req *SendMessageRequest) (*SendMessageResponse, error) {
	im.mutex.RLock()
	client, exists := im.instances[instanceName]
	im.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("instance %s not found or not connected", instanceName)
	}

	if client.GetStatus() != "open" {
		return nil, fmt.Errorf("instance %s is not connected", instanceName)
	}

	var messageID string
	var err error

	switch req.Type {
	case "text":
		messageID, err = client.SendText(ctx, req.To, req.Text)
	case "image", "video", "audio", "document":
		messageID, err = client.SendMedia(ctx, req.To, req.MediaURL, req.Type, req.Caption, req.ViewOnce)
	default:
		return nil, fmt.Errorf("unsupported message type: %s", req.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &SendMessageResponse{
		MessageID: messageID,
		Status:    "sent",
	}, nil
}

// emitEvent emits an instance event
func (im *InstanceManager) emitEvent(eventType, instanceName string, data interface{}) {
	event := InstanceEvent{
		Type:      eventType,
		Instance:  instanceName,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case im.eventCh <- event:
	default:
		zap.L().Warn("Event channel full, dropping event", zap.String("type", eventType))
	}
}

// processEvents processes instance events
func (im *InstanceManager) processEvents() {
	for {
		select {
		case event := <-im.eventCh:
			im.handleEvent(event)
		case <-im.shutdownCh:
			return
		}
	}
}

// handleEvent handles an instance event
func (im *InstanceManager) handleEvent(event InstanceEvent) {
	// Here you can add webhook sending, logging, etc.
	zap.L().Debug("Processing instance event",
		zap.String("type", event.Type),
		zap.String("instance", event.Instance),
		zap.Time("timestamp", event.Timestamp))
}

// cleanupRoutine periodically cleans up inactive instances
func (im *InstanceManager) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			im.cleanupInactiveInstances()
		case <-im.shutdownCh:
			return
		}
	}
}

// cleanupInactiveInstances removes inactive instances
func (im *InstanceManager) cleanupInactiveInstances() {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	now := time.Now()
	for name, client := range im.instances {
		if client.GetStatus() == "close" {
			// Check if instance has been inactive for too long
			if now.Sub(client.instance.LastSeen) > 1*time.Hour {
				zap.L().Info("Removing inactive instance", zap.String("name", name))
				client.Disconnect()
				delete(im.instances, name)
			}
		}
	}
}

// Shutdown gracefully shuts down the instance manager
func (im *InstanceManager) Shutdown(ctx context.Context) error {
	close(im.shutdownCh)

	im.mutex.Lock()
	defer im.mutex.Unlock()

	// Disconnect all clients
	for name, client := range im.instances {
		zap.L().Info("Disconnecting instance", zap.String("name", name))
		client.Disconnect()
	}

	return nil
}

// Request/Response types

type CreateInstanceRequest struct {
	Name             string   `json:"name" binding:"required"`
	Token            string   `json:"token"`
	RejectCall       bool     `json:"rejectCall"`
	MsgCall          string   `json:"msgCall"`
	GroupsIgnore     bool     `json:"groupsIgnore"`
	AlwaysOnline     bool     `json:"alwaysOnline"`
	ReadMessages     bool     `json:"readMessages"`
	ReadStatus       bool     `json:"readStatus"`
	SyncFullHistory  bool     `json:"syncFullHistory"`
	WAVoipToken      string   `json:"wavoipToken"`
	AutoReadMessages bool     `json:"autoReadMessages"`
	ReadDelay        int      `json:"readDelay"`
	WebhookURL       string   `json:"webhookUrl"`
	WebhookEvents    []string `json:"webhookEvents"`
	WebhookBase64    bool     `json:"webhookBase64"`
}

type CreateInstanceResponse struct {
	Instance *models.Instance `json:"instance"`
}

type ConnectInstanceResponse struct {
	Connected bool           `json:"connected"`
	Message   string         `json:"message"`
	QRCode    *models.QRCode `json:"qrcode,omitempty"`
}

type SendMessageRequest struct {
	To       string `json:"to" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Text     string `json:"text,omitempty"`
	MediaURL string `json:"mediaUrl,omitempty"`
	Caption  string `json:"caption,omitempty"`
	ViewOnce bool   `json:"viewOnce"`
}

type SendMessageResponse struct {
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
}
