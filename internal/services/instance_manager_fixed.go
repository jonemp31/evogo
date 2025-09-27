package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.uber.org/zap"
)

// InstanceManager manages WhatsApp instances
type InstanceManager struct {
	instances  map[string]*WhatsAppClient
	container  *sqlstore.Container
	repository InstanceRepository
	mutex      sync.RWMutex
	logger     *zap.Logger
}

// InstanceRepository interface for instance data persistence
type InstanceRepository interface {
	Create(instance *models.Instance) error
	GetByName(name string) (*models.Instance, error)
	GetByID(id string) (*models.Instance, error)
	List() ([]*models.Instance, error)
	Update(instance *models.Instance) error
	Delete(id string) error
}

// NewInstanceManager creates a new instance manager
func NewInstanceManager(container *sqlstore.Container, repository InstanceRepository) *InstanceManager {
	return &InstanceManager{
		instances:  make(map[string]*WhatsAppClient),
		container:  container,
		repository: repository,
		logger:     zap.L().Named("InstanceManager"),
	}
}

// CreateInstance creates a new WhatsApp instance
func (im *InstanceManager) CreateInstance(ctx context.Context, name string, webhookURL string) (*models.Instance, error) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.logger.Info("Creating new instance", zap.String("name", name))

	// Check if instance already exists
	if _, exists := im.instances[name]; exists {
		return nil, fmt.Errorf("instance %s already exists", name)
	}

	// Check if instance exists in database
	if existing, err := im.repository.GetByName(name); err == nil && existing != nil {
		return nil, fmt.Errorf("instance %s already exists in database", name)
	}

	// Create new instance model
	instance := &models.Instance{
		Name:       name,
		WebhookURL: webhookURL,
		Status:     "close",
		CreatedAt:  time.Now(),
	}

	// Create WhatsApp client
	client, err := NewWhatsAppClient(instance, im.container)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	// Store instance in memory
	im.instances[name] = client

	// Save to database
	if err := im.repository.Create(instance); err != nil {
		// Remove from memory if database save fails
		delete(im.instances, name)
		return nil, fmt.Errorf("failed to save instance to database: %w", err)
	}

	im.logger.Info("Instance created successfully", zap.String("name", name))
	return instance, nil
}

// ConnectInstance connects an instance to WhatsApp
func (im *InstanceManager) ConnectInstance(ctx context.Context, name string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.logger.Info("Connecting instance", zap.String("name", name))

	// Get instance from memory or database
	client, exists := im.instances[name]
	if !exists {
		// Try to load from database
		instance, err := im.repository.GetByName(name)
		if err != nil {
			return fmt.Errorf("instance %s not found", name)
		}

		// Create client
		client, err = NewWhatsAppClient(instance, im.container)
		if err != nil {
			return fmt.Errorf("failed to create WhatsApp client: %w", err)
		}

		im.instances[name] = client
	}

	// Connect client
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect instance: %w", err)
	}

	// Update status in database
	client.instance.Status = client.GetStatus()
	if err := im.repository.Update(client.instance); err != nil {
		im.logger.Warn("Failed to update instance status in database",
			zap.String("name", name),
			zap.Error(err))
	}

	im.logger.Info("Instance connected successfully", zap.String("name", name))
	return nil
}

// GetInstanceStatus returns the status of an instance
func (im *InstanceManager) GetInstanceStatus(ctx context.Context, name string) (*models.Instance, error) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	im.logger.Debug("Getting instance status", zap.String("name", name))

	// Try to get from memory first
	if client, exists := im.instances[name]; exists {
		client.instance.Status = client.GetStatus()
		return client.instance, nil
	}

	// Try to get from database
	instance, err := im.repository.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	return instance, nil
}

// ListInstances returns all instances
func (im *InstanceManager) ListInstances(ctx context.Context) ([]*models.Instance, error) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	im.logger.Debug("Listing all instances")

	// Get instances from database
	instances, err := im.repository.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	// Update status for instances in memory
	for _, instance := range instances {
		if client, exists := im.instances[instance.Name]; exists {
			instance.Status = client.GetStatus()
		}
	}

	return instances, nil
}

// DeleteInstance deletes an instance
func (im *InstanceManager) DeleteInstance(ctx context.Context, name string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.logger.Info("Deleting instance", zap.String("name", name))

	// Get instance from memory or database
	instance, err := im.getInstanceByName(name)
	if err != nil {
		return fmt.Errorf("instance %s not found", name)
	}

	// Disconnect client if exists in memory
	if client, exists := im.instances[name]; exists {
		client.Disconnect()
		delete(im.instances, name)
	}

	// Delete from database
	if err := im.repository.Delete(instance.ID); err != nil {
		return fmt.Errorf("failed to delete instance from database: %w", err)
	}

	im.logger.Info("Instance deleted successfully", zap.String("name", name))
	return nil
}

// LogoutInstance logs out an instance
func (im *InstanceManager) LogoutInstance(ctx context.Context, name string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.logger.Info("Logging out instance", zap.String("name", name))

	// Get client from memory
	client, exists := im.instances[name]
	if !exists {
		return fmt.Errorf("instance %s not found in memory", name)
	}

	// Logout client
	if err := client.Logout(ctx); err != nil {
		return fmt.Errorf("failed to logout instance: %w", err)
	}

	// Update status
	client.instance.Status = "close"
	if err := im.repository.Update(client.instance); err != nil {
		im.logger.Warn("Failed to update instance status",
			zap.String("name", name),
			zap.Error(err))
	}

	im.logger.Info("Instance logged out successfully", zap.String("name", name))
	return nil
}

// UpdateInstanceWebhook updates the webhook URL for an instance
func (im *InstanceManager) UpdateInstanceWebhook(ctx context.Context, name, webhookURL string) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.logger.Info("Updating instance webhook", zap.String("name", name))

	// Get instance
	instance, err := im.getInstanceByName(name)
	if err != nil {
		return fmt.Errorf("instance %s not found", name)
	}

	// Update webhook URL
	instance.WebhookURL = webhookURL
	instance.UpdatedAt = time.Now()

	// Save to database
	if err := im.repository.Update(instance); err != nil {
		return fmt.Errorf("failed to update instance webhook: %w", err)
	}

	// Update in memory if exists
	if client, exists := im.instances[name]; exists {
		client.instance.WebhookURL = webhookURL
	}

	im.logger.Info("Instance webhook updated successfully", zap.String("name", name))
	return nil
}

// UpdateInstanceSettings updates settings for an instance
func (im *InstanceManager) UpdateInstanceSettings(ctx context.Context, name string, settings *models.InstanceSettings) error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.logger.Info("Updating instance settings", zap.String("name", name))

	// Get instance
	instance, err := im.getInstanceByName(name)
	if err != nil {
		return fmt.Errorf("instance %s not found", name)
	}

	// Update settings
	instance.Settings = settings
	instance.UpdatedAt = time.Now()

	// Save to database
	if err := im.repository.Update(instance); err != nil {
		return fmt.Errorf("failed to update instance settings: %w", err)
	}

	// Update in memory if exists
	if client, exists := im.instances[name]; exists {
		client.instance.Settings = settings
	}

	im.logger.Info("Instance settings updated successfully", zap.String("name", name))
	return nil
}

// GetQRCode returns the QR code for an instance
func (im *InstanceManager) GetQRCode(ctx context.Context, name string) (string, error) {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	// Get client from memory
	client, exists := im.instances[name]
	if !exists {
		return "", fmt.Errorf("instance %s not found in memory", name)
	}

	// Get QR code
	qrCode := client.GetQRCode()
	if qrCode == "" {
		return "", fmt.Errorf("no QR code available for instance %s", name)
	}

	return qrCode, nil
}

// getInstanceByName gets an instance by name (internal method)
func (im *InstanceManager) getInstanceByName(name string) (*models.Instance, error) {
	// Try to get from memory first
	if client, exists := im.instances[name]; exists {
		return client.instance, nil
	}

	// Try to get from database
	instance, err := im.repository.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found", name)
	}

	return instance, nil
}

// getClient returns a WhatsApp client by name (internal method)
func (im *InstanceManager) getClient(name string) *WhatsAppClient {
	if client, exists := im.instances[name]; exists {
		return client
	}
	return nil
}
