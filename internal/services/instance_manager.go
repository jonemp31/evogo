package services

import (
	"fmt"
	"sync"

	"github.com/jonemp31/evogo/internal/models"
	"github.com/jonemp31/evogo/internal/repository"
	"go.uber.org/zap"
)

type InstanceManager struct {
	Instances      map[string]*WhatsAppClient
	mu             sync.Mutex
	repo           repository.InstanceRepository
	logger         *zap.Logger
	webhookService *WebhookService
}

func NewInstanceManager(repo repository.InstanceRepository, logger *zap.Logger, webhookService *WebhookService) *InstanceManager {
	return &InstanceManager{
		Instances:      make(map[string]*WhatsAppClient),
		repo:           repo,
		logger:         logger,
		webhookService: webhookService,
	}
}

func (m *InstanceManager) CreateInstance(instanceData models.Instance) (*models.Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Instances[instanceData.Name]; exists {
		return nil, fmt.Errorf("instância '%s' já existe", instanceData.Name)
	}

	instanceData.Status = "created"
	createdInstance, err := m.repo.Create(instanceData)
	if err != nil {
		return nil, err
	}

	m.logger.Info("Instância criada e registrada", zap.String("name", createdInstance.Name))
	return createdInstance, nil
}

func (m *InstanceManager) ConnectInstance(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.Instances[name]; exists && (client.Instance.Status == "open" || client.Instance.Status == "connecting") {
		return "", fmt.Errorf("instância '%s' já está conectada ou conectando", name)
	}

	instanceData, err := m.repo.FindByName(name)
	if err != nil {
		return "", fmt.Errorf("instância '%s' não encontrada no banco de dados", name)
	}

	// CORREÇÃO: Passando o webhookService para o NewWhatsAppClient
	client, err := NewWhatsAppClient(instanceData, m.logger, m.webhookService)
	if err != nil {
		return "", fmt.Errorf("erro ao criar cliente do WhatsApp para '%s': %v", name, err)
	}

	m.Instances[name] = client

	qrCode, err := client.Connect()
	if err != nil {
		delete(m.Instances, name)
		return "", fmt.Errorf("falha ao conectar instância '%s': %v", name, err)
	}

	m.logger.Info("Instância conectando...", zap.String("name", name))
	return qrCode, nil
}

func (m *InstanceManager) DisconnectInstance(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.Instances[name]
	if !exists {
		return fmt.Errorf("instância '%s' não está conectada ou não existe", name)
	}

	if err := client.Disconnect(); err != nil {
		m.logger.Error("Erro ao desconectar cliente WhatsApp, removendo da memória de qualquer forma", zap.Error(err), zap.String("instance", name))
	}

	delete(m.Instances, name)
	m.logger.Info("Instância desconectada", zap.String("name", name))
	return nil
}

func (m *InstanceManager) GetInstance(name string) (*WhatsAppClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.Instances[name]
	if !exists {
		return nil, fmt.Errorf("instância '%s' não encontrada ou não está ativa", name)
	}
	return client, nil
}

func (m *InstanceManager) ListAllInstances() ([]*models.Instance, error) {
	return m.repo.FindAll()
}

func (m *InstanceManager) GetInstanceStatus(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.Instances[name]; exists {
		return client.Instance.Status, nil
	}

	instance, err := m.repo.FindByName(name)
	if err != nil {
		return "not_found", err
	}

	return instance.Status, nil // Retorna o status do banco se não estiver ativo
}

func (m *InstanceManager) getInstanceByName(name string) (*models.Instance, error) {
	return m.repo.FindByName(name)
}

func (m *InstanceManager) UpdateInstanceWebhook(name, webhookURL string) error {
	instance, err := m.repo.FindByName(name)
	if err != nil {
		return fmt.Errorf("instância '%s' não encontrada", name)
	}

	instance.WebhookURL = webhookURL
	return m.repo.Update(instance)
}

func (m *InstanceManager) UpdateInstanceSettings(name string, settings *models.InstanceSettings) error {
	instance, err := m.repo.FindByName(name)
	if err != nil {
		return fmt.Errorf("instância '%s' não encontrada", name)
	}

	instance.Settings = *settings
	return m.repo.Update(instance)
}
