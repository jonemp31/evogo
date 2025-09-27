# Evolution API Go

Uma implementação em Go da Evolution API, focada nos recursos essenciais do WhatsApp: envio de mensagens (texto, mídia, documentos), gerenciamento de instâncias, webhooks e status de conexão.

## 🚀 Características

- **Multi-tenant**: Suporte a múltiplas instâncias WhatsApp simultâneas
- **Alta Performance**: Construído em Go para máxima eficiência
- **Webhooks**: Sistema robusto de webhooks com retry automático
- **View Once**: Suporte a mensagens que se autodestróem
- **Redis Cache**: Cache distribuído para alta performance
- **PostgreSQL**: Banco de dados robusto e escalável
- **Docker**: Containerização completa com docker-compose

## 📋 Requisitos

- Go 1.21+
- PostgreSQL 13+
- Redis 6+
- Docker & Docker Compose (opcional)

## 🛠️ Instalação

### Opção 1: Docker Compose (Recomendado)

```bash
# Clone o repositório
git clone <repository-url>
cd evolution-go

# Configure as variáveis de ambiente
cp env.example .env
# Edite o .env com suas configurações

# Inicie todos os serviços
docker-compose up -d

# Verifique se está funcionando
curl http://localhost:8080/health
```

### Opção 2: Instalação Manual

```bash
# Clone o repositório
git clone <repository-url>
cd evolution-go

# Instale dependências
go mod download

# Configure o banco de dados
# Execute as migrações em migrations/001_create_instances_table.sql

# Configure as variáveis de ambiente
cp env.example .env
# Edite o .env

# Execute a aplicação
go run cmd/server/main.go
```

## 🔧 Configuração

### Variáveis de Ambiente

```env
# Servidor
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_URL=http://localhost:8080

# Banco de Dados
DATABASE_URL=postgres://user:password@localhost:5432/evolution_api?sslmode=disable
DATABASE_PROVIDER=postgresql

# Redis
REDIS_URL=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# API
API_KEY=your-secret-api-key-here
DEBUG_MODE=false

# Webhooks
WEBHOOK_TIMEOUT=30
WEBHOOK_RETRY_ATTEMPTS=3
```

## 📚 API Endpoints

### Instâncias

#### Criar Instância
```http
POST /api/instance
Content-Type: application/json
apikey: your-api-key

{
  "name": "minha-instancia",
  "token": "meu-token-opcional",
  "webhookUrl": "https://meu-site.com/webhook",
  "webhookEvents": ["messages.upsert", "connection.update"]
}
```

#### Conectar Instância
```http
POST /api/instance/minha-instancia/connect
apikey: your-api-key
```

#### Status da Instância
```http
GET /api/instance/minha-instancia/status
apikey: your-api-key
```

#### Listar Instâncias
```http
GET /api/instance
apikey: your-api-key
```

#### Deletar Instância
```http
DELETE /api/instance/minha-instancia
apikey: your-api-key
```

### Mensagens

#### Enviar Texto
```http
POST /api/message/minha-instancia/text
Content-Type: application/json
apikey: your-api-key

{
  "to": "5511999999999@s.whatsapp.net",
  "text": "Olá! Esta é uma mensagem de teste."
}
```

#### Enviar Imagem
```http
POST /api/message/minha-instancia/image
Content-Type: application/json
apikey: your-api-key

{
  "to": "5511999999999@s.whatsapp.net",
  "mediaUrl": "https://example.com/image.jpg",
  "caption": "Legenda da imagem",
  "viewOnce": true
}
```

#### Enviar Vídeo
```http
POST /api/message/minha-instancia/video
Content-Type: application/json
apikey: your-api-key

{
  "to": "5511999999999@s.whatsapp.net",
  "mediaUrl": "https://example.com/video.mp4",
  "caption": "Legenda do vídeo",
  "viewOnce": true
}
```

#### Enviar Áudio
```http
POST /api/message/minha-instancia/audio
Content-Type: application/json
apikey: your-api-key

{
  "to": "5511999999999@s.whatsapp.net",
  "mediaUrl": "https://example.com/audio.mp3",
  "viewOnce": true
}
```

#### Enviar Documento
```http
POST /api/message/minha-instancia/document
Content-Type: application/json
apikey: your-api-key

{
  "to": "5511999999999@s.whatsapp.net",
  "mediaUrl": "https://example.com/document.pdf",
  "caption": "Descrição do documento"
}
```

## 🔗 Webhooks

### Eventos Suportados

- `instance.create` - Instância criada
- `instance.delete` - Instância deletada
- `instance.logout` - Instância desconectada
- `connection.update` - Status de conexão alterado
- `qrcode.updated` - QR Code atualizado
- `messages.upsert` - Nova mensagem recebida
- `messages.update` - Status de mensagem alterado

### Exemplo de Webhook

```json
{
  "event": "messages.upsert",
  "instance": "minha-instancia",
  "data": {
    "id": "message-id",
    "instanceId": "instance-uuid",
    "remoteJid": "5511999999999@s.whatsapp.net",
    "fromMe": false,
    "message": "Olá!",
    "messageType": "conversation",
    "timestamp": "2024-01-01T12:00:00Z",
    "status": "received"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 🔄 Compatibilidade com Evolution API

A API mantém compatibilidade com os endpoints da Evolution API original:

- `POST /instance/create`
- `GET /instance/fetchInstances`
- `GET /instance/connect/:name`
- `GET /instance/connectionState/:name`
- `DELETE /instance/logout/:name`
- `DELETE /instance/delete/:name`
- `POST /message/sendText/:name`
- `POST /message/sendMedia/:name`
- `POST /message/sendImage/:name`
- `POST /message/sendVideo/:name`
- `POST /message/sendAudio/:name`
- `POST /message/sendDocument/:name`

## 🏗️ Arquitetura

```
evolution-go/
├── cmd/server/          # Aplicação principal
├── internal/
│   ├── config/         # Configurações
│   ├── handlers/       # Handlers HTTP
│   ├── middleware/     # Middlewares
│   ├── models/         # Modelos de dados
│   ├── repository/     # Camada de dados
│   ├── routes/         # Definição de rotas
│   └── services/       # Lógica de negócio
├── migrations/         # Migrações do banco
├── docker-compose.yml  # Orquestração Docker
└── Dockerfile         # Imagem Docker
```

## 🚀 Performance

- **~3-5x mais rápido** que Node.js
- **~50-70% menos consumo** de memória
- **Suporte a milhares** de instâncias simultâneas
- **Goroutines nativas** para alta concorrência

## 🔒 Segurança

- Autenticação via API Key
- Validação de entrada robusta
- Logs estruturados
- Containerização segura

## 📊 Monitoramento

- Health checks integrados
- Logs estruturados em JSON
- Métricas de performance
- Graceful shutdown

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a Licença Apache 2.0 - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

- **Issues**: Abra uma issue no GitHub
- **Documentação**: Consulte este README
- **Comunidade**: Entre em contato via GitHub Discussions

---

**Evolution API Go** - A próxima geração da API WhatsApp em Go! 🚀
