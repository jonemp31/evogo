package services

import (
	"fmt"

	"go.uber.org/zap"
)

type MessageService struct {
	instanceManager *InstanceManager
	logger          *zap.Logger
}

func NewMessageService(instanceManager *InstanceManager, logger *zap.Logger) *MessageService {
	return &MessageService{
		instanceManager: instanceManager,
		logger:          logger,
	}
}

// SendText envia uma mensagem de texto
func (s *MessageService) SendText(instanceName, remoteJid, text string) (string, error) {
	client, err := s.instanceManager.GetInstance(instanceName)
	if err != nil {
		return "", err
	}

	if client.Instance.Status != "open" {
		return "", fmt.Errorf("instância '%s' não está conectada", instanceName)
	}

	return client.SendTextMessage(remoteJid, text)
}

// SendMedia envia uma mensagem de mídia
func (s *MessageService) SendMedia(instanceName, remoteJid, mediaType string, data []byte, mimeType, caption string, viewOnce, ptt bool) (string, error) {
	client, err := s.instanceManager.GetInstance(instanceName)
	if err != nil {
		return "", err
	}

	if client.Instance.Status != "open" {
		return "", fmt.Errorf("instância '%s' não está conectada", instanceName)
	}

	return client.SendMediaMessage(remoteJid, mediaType, data, mimeType, caption, viewOnce, ptt)
}
