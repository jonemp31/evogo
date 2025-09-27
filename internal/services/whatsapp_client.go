package services

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/jonemp31/evogo/internal/models"
	"go.uber.org/zap"
)

// WhatsAppClient gerencia a conexão e comunicação com o WhatsApp
type WhatsAppClient struct {
	Conn          *whatsapp.Conn
	Instance      *models.Instance
	Logger        *zap.Logger
	WebhookSender WebhookSender
	StartTime     time.Time
}

// WebhookSender define a interface para enviar webhooks.
type WebhookSender interface {
	SendWebhook(instanceName, eventType string, payload interface{})
}

// NewWhatsAppClient cria e inicializa um novo cliente WhatsApp
func NewWhatsAppClient(instance *models.Instance, logger *zap.Logger, webhookSender WebhookSender) (*WhatsAppClient, error) {
	conn, err := whatsapp.NewConn(20 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar conexão: %v", err)
	}

	client := &WhatsAppClient{
		Conn:          conn,
		Instance:      instance,
		Logger:        logger,
		WebhookSender: webhookSender,
	}

	client.Conn.AddHandler(client.eventHandler)
	return client, nil
}

func (wac *WhatsAppClient) eventHandler(event interface{}) {
	wac.Logger.Debug("Evento recebido", zap.Any("type", fmt.Sprintf("%T", event)))

	switch e := event.(type) {
	case *whatsapp.TextMessage:
		wac.handleMessage(e.Info, map[string]interface{}{"conversation": e.Text})
	case *whatsapp.ImageMessage:
		wac.handleMessage(e.Info, map[string]interface{}{"imageMessage": map[string]interface{}{"caption": e.Caption, "mimetype": e.Type}})
	case *whatsapp.VideoMessage:
		wac.handleMessage(e.Info, map[string]interface{}{"videoMessage": map[string]interface{}{"caption": e.Caption, "mimetype": e.Type}})
	case *whatsapp.AudioMessage:
		wac.handleMessage(e.Info, map[string]interface{}{"audioMessage": map[string]interface{}{"ptt": e.Ptt, "mimetype": e.Type}})
	case *whatsapp.DocumentMessage:
		wac.handleMessage(e.Info, map[string]interface{}{"documentMessage": map[string]interface{}{"title": e.Title, "mimetype": e.Type}})

	case *whatsapp.ConnectingEvent:
		wac.Instance.Status = "connecting"
		wac.WebhookSender.SendWebhook(wac.Instance.Name, "connection.update", map[string]string{"status": "connecting"})
	case *whatsapp.ConnectedEvent:
		wac.Instance.Status = "open"
		wac.StartTime = time.Now()
		wac.WebhookSender.SendWebhook(wac.Instance.Name, "connection.update", map[string]string{"status": "open"})
	case *whatsapp.DisconnectedEvent:
		wac.Instance.Status = "close"
		wac.WebhookSender.SendWebhook(wac.Instance.Name, "connection.update", map[string]string{"status": "close"})
	case *whatsapp.ErrorEvent:
		wac.Logger.Error("Erro recebido da conexão", zap.Error(e.Err))

	case whatsapp.Presence:
		payload := map[string]interface{}{"jid": e.Jid, "presence": e.Type, "t": e.Timestamp}
		wac.WebhookSender.SendWebhook(wac.Instance.Name, "presence.update", payload)

	case whatsapp.Receipt:
		if e.Type == whatsapp.ReceiptTypeRead || e.Type == whatsapp.ReceiptTypeDelivered {
			payload := map[string]interface{}{"id": e.MessageID, "remoteJid": e.Jid, "status": e.Type, "t": e.Timestamp}
			wac.WebhookSender.SendWebhook(wac.Instance.Name, "messages.update", payload)
		}
	}
}

func (wac *WhatsAppClient) handleMessage(info whatsapp.MessageInfo, messageContent map[string]interface{}) {
	if info.Timestamp < uint64(wac.StartTime.Unix()) {
		return
	}
	payload := map[string]interface{}{
		"event":    "messages.upsert",
		"instance": wac.Instance.Name,
		"data": map[string]interface{}{
			"key":              map[string]interface{}{"remoteJid": info.RemoteJid, "fromMe": info.FromMe, "id": info.Id},
			"pushName":         info.Sender.Name,
			"message":          messageContent,
			"messageTimestamp": info.Timestamp,
		},
	}
	wac.WebhookSender.SendWebhook(wac.Instance.Name, "messages.upsert", payload)
}

func (wac *WhatsAppClient) Connect() (string, error) {
	session, err := wac.readSession()
	if err == nil {
		wac.Logger.Info("Restaurando sessão existente...")
		session, err = wac.Conn.RestoreWithSession(session)
		if err != nil {
			wac.Logger.Warn("Falha ao restaurar sessão, será necessário um novo QR Code", zap.Error(err))
			return wac.loginWithNewQRCode()
		}
		return "", nil // Conexão restaurada, sem QR code
	}

	wac.Logger.Info("Nenhuma sessão encontrada ou falha ao ler.")
	return wac.loginWithNewQRCode()
}

func (wac *WhatsAppClient) loginWithNewQRCode() (string, error) {
	qr := make(chan string)
	var loginErr error
	go func() {
		_, loginErr = wac.Conn.Login(qr)
		if loginErr != nil {
			close(qr) // Fecha o canal se o login falhar
		}
	}()

	qrCode, ok := <-qr
	if !ok {
		return "", fmt.Errorf("falha no login com QR Code: %w", loginErr)
	}
	return qrCode, nil
}

func (wac *WhatsAppClient) Disconnect() error {
	if wac.Conn != nil {
		defer wac.Conn.Disconnect()
		return wac.saveSession()
	}
	return nil
}

func (wac *WhatsAppClient) SendTextMessage(remoteJid, text string) (string, error) {
	msg := whatsapp.TextMessage{Info: whatsapp.MessageInfo{RemoteJid: remoteJid}, Text: text}
	return wac.Conn.Send(msg)
}

func (wac *WhatsAppClient) SendMediaMessage(remoteJid, mediaType string, data []byte, mimeType, caption string, viewOnce, ptt bool) (string, error) {
	var msg interface{}
	switch mediaType {
	case "image":
		msg = whatsapp.ImageMessage{Info: whatsapp.MessageInfo{RemoteJid: remoteJid}, Type: mimeType, Content: data, Caption: caption, ViewOnce: viewOnce}
	case "video":
		msg = whatsapp.VideoMessage{Info: whatsapp.MessageInfo{RemoteJid: remoteJid}, Type: mimeType, Content: data, Caption: caption, ViewOnce: viewOnce}
	case "audio":
		msg = whatsapp.AudioMessage{Info: whatsapp.MessageInfo{RemoteJid: remoteJid}, Type: mimeType, Content: data, Ptt: ptt}
	case "document":
		msg = whatsapp.DocumentMessage{Info: whatsapp.MessageInfo{RemoteJid: remoteJid}, Type: mimeType, Content: data, Title: caption}
	default:
		return "", fmt.Errorf("tipo de mídia não suportado: %s", mediaType)
	}
	return wac.Conn.Send(msg)
}

func (wac *WhatsAppClient) saveSession() error {
	session, err := wac.Conn.Store()
	if err != nil {
		return fmt.Errorf("erro ao obter sessão para salvar: %w", err)
	}
	_ = os.MkdirAll("./sessions", os.ModePerm)
	filePath := fmt.Sprintf("./sessions/%s.gob", wac.Instance.Name)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de sessão: %w", err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(session); err != nil {
		return fmt.Errorf("erro ao salvar sessão: %w", err)
	}
	wac.Logger.Info("Sessão salva com sucesso", zap.String("instance", wac.Instance.Name))
	return nil
}

func (wac *WhatsAppClient) readSession() (whatsapp.Session, error) {
	var session whatsapp.Session
	filePath := fmt.Sprintf("./sessions/%s.gob", wac.Instance.Name)
	file, err := os.Open(filePath)
	if err != nil {
		return session, fmt.Errorf("erro ao abrir arquivo de sessão: %w", err)
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&session); err != nil {
		return session, fmt.Errorf("erro ao decodificar sessão: %w", err)
	}
	return session, nil
}
