package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonemp31/evogo/internal/models"
	"github.com/jonemp31/evogo/internal/repository"
	"go.uber.org/zap"
)

// InstanceManager gerencia o ciclo de vida das instâncias do WhatsApp.
type InstanceManager struct {
	repo           repository.InstanceRepository
	webhookService WebhookSender
	logger         *zap.Logger
	instances      sync.Map // Armazena os clientes ativos: [instanceName]*WhatsAppClient
}

// NewInstanceManager cria um novo gerenciador de instâncias.
func NewInstanceManager(repo repository.InstanceRepository, webhookService WebhookSender, logger *zap.Logger) *InstanceManager {
	return &InstanceManager{
		repo:           repo,
		webhookService: webhookService,
		logger:         logger,
		instances:      sync.Map{},
	}
}

// CreateInstance cria uma nova instância no banco de dados.
func (m *InstanceManager) CreateInstance(ctx context.Context, instanceData *models.Instance) (*models.Instance, error) {
	existing, err := m.repo.FindByName(instanceData.Name)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar instância existente: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("instância com o nome '%s' já existe", instanceData.Name)
	}

	instanceData.Status = "created"
	if err := m.repo.Create(instanceData); err != nil {
		return nil, fmt.Errorf("falha ao guardar a instância: %w", err)
	}
	return instanceData, nil
}

// ConnectInstance conecta uma instância ao WhatsApp e retorna o QR Code se necessário.
func (m *InstanceManager) ConnectInstance(ctx context.Context, instanceName string) (string, error) {
	instanceData, err := m.repo.FindByName(instanceName)
	if err != nil {
		return "", fmt.Errorf("falha ao encontrar a instância: %w", err)
	}
	if instanceData == nil {
		return "", fmt.Errorf("instância não encontrada")
	}

	client := NewWhatsAppClient(instanceData, m.webhookService, m.logger)
	m.instances.Store(instanceName, client)

	qrCode, err := client.Connect(ctx)
	if err != nil {
		m.instances.Delete(instanceName) // Limpa se a conexão falhar
		return "", fmt.Errorf("falha ao conectar o cliente whatsapp: %w", err)
	}

	instanceData.Status = "connecting"
	m.repo.Update(instanceData)

	return qrCode, nil
}

// GetInstance retorna um cliente de instância ativo.
func (m *InstanceManager) GetInstance(instanceName string) (*WhatsAppClient, bool) {
	client, ok := m.instances.Load(instanceName)
	if !ok {
		return nil, false
	}
	return client.(*WhatsAppClient), true
}

// GetAllInstancesInfo retorna informações de todas as instâncias do repositório.
func (m *InstanceManager) GetAllInstancesInfo() ([]*models.Instance, error) {
	return m.repo.FindAll()
}

// DisconnectInstance desconecta uma instância do WhatsApp.
func (m *InstanceManager) DisconnectInstance(instanceName string) error {
	client, ok := m.GetInstance(instanceName)
	if !ok {
		return fmt.Errorf("instância não encontrada ou não conectada")
	}

	client.Disconnect()
	m.instances.Delete(instanceName)

	instanceData, err := m.repo.FindByName(instanceName)
	if err != nil {
		return err
	}
	if instanceData != nil {
		instanceData.Status = "disconnected"
		m.repo.Update(instanceData)
	}

	return nil
}

// DeleteInstance remove completamente uma instância.
func (m *InstanceManager) DeleteInstance(instanceName string) error {
	if client, ok := m.GetInstance(instanceName); ok {
		client.Disconnect()
		m.instances.Delete(instanceName)
	}
	return m.repo.Delete(instanceName)
}

// RestoreInstances tenta reconectar instâncias que não estavam 'disconnected' ou 'created'.
func (m *InstanceManager) RestoreInstances() {
	m.logger.Info("A restaurar instâncias...")
	instances, err := m.repo.FindAll()
	if err != nil {
		m.logger.Error("Falha ao buscar instâncias para restaurar", zap.Error(err))
		return
	}

	for _, instance := range instances {
		if instance.Status != "disconnected" && instance.Status != "created" {
			go func(inst *models.Instance) {
				m.logger.Info("A tentar restaurar a conexão", zap.String("instance", inst.Name))
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()
				_, err := m.ConnectInstance(ctx, inst.Name)
				if err != nil {
					m.logger.Error("Falha ao restaurar instância", zap.String("instance", inst.Name), zap.Error(err))
				}
			}(instance)
		}
	}
}
