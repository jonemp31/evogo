# 🚀 Evolution API Go - Quick Start

## ⚡ Início Rápido (5 minutos)

### 1. Pré-requisitos
```bash
# Instalar Docker e Docker Compose
# Ou Go 1.21+ + PostgreSQL + Redis
```

### 2. Executar com Docker (Recomendado)
```bash
# Clonar e entrar na pasta
cd evolution-go

# Configurar ambiente
cp env.example .env
# Editar .env se necessário

# Iniciar todos os serviços
docker-compose up -d

# Verificar se está funcionando
curl http://localhost:8080/health
```

### 3. Criar sua primeira instância
```bash
# Criar instância
curl -X POST http://localhost:8080/api/instance \
  -H "Content-Type: application/json" \
  -H "apikey: your-secret-api-key-here" \
  -d '{
    "name": "minha-instancia",
    "webhookUrl": "https://meu-site.com/webhook"
  }'

# Conectar instância (gerará QR code)
curl -X POST http://localhost:8080/api/instance/minha-instancia/connect \
  -H "apikey: your-secret-api-key-here"
```

### 4. Enviar mensagens
```bash
# Enviar texto
curl -X POST http://localhost:8080/api/message/minha-instancia/text \
  -H "Content-Type: application/json" \
  -H "apikey: your-secret-api-key-here" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "text": "Olá! Esta é uma mensagem de teste."
  }'

# Enviar imagem
curl -X POST http://localhost:8080/api/message/minha-instancia/image \
  -H "Content-Type: application/json" \
  -H "apikey: your-secret-api-key-here" \
  -d '{
    "to": "5511999999999@s.whatsapp.net",
    "mediaUrl": "https://example.com/image.jpg",
    "caption": "Legenda da imagem",
    "viewOnce": true
  }'
```

## 📊 Monitoramento

### Health Checks
- **Health**: `GET /health`
- **Readiness**: `GET /ready` 
- **Liveness**: `GET /live`

### Métricas Prometheus
- **Métricas**: `GET /metrics`

### Logs
```bash
# Ver logs da aplicação
docker-compose logs -f evolution-api

# Ver logs do banco
docker-compose logs -f postgres

# Ver logs do Redis
docker-compose logs -f redis
```

## 🔧 Desenvolvimento

### Instalação Manual (sem Docker)
```bash
# Instalar dependências
go mod download

# Configurar banco PostgreSQL
createdb evolution_api

# Executar migrações
psql evolution_api < migrations/001_create_instances_table.sql

# Configurar Redis
redis-server

# Executar aplicação
go run cmd/server/main.go
```

### Comandos Úteis
```bash
# Build
make build

# Testes
make test

# Formatar código
make fmt

# Docker build
make docker-build

# Parar serviços
make docker-stop
```

## 🎯 Funcionalidades Principais

### ✅ Implementado
- ✅ **Multi-instâncias WhatsApp** simultâneas
- ✅ **Envio de mensagens**: texto, imagem, vídeo, áudio, documento
- ✅ **View Once**: mensagens que se autodestróem
- ✅ **Webhooks**: sistema robusto com retry automático
- ✅ **Status de conexão**: tempo real
- ✅ **Cache Redis**: alta performance
- ✅ **Rate Limiting**: proteção contra abuso
- ✅ **Métricas Prometheus**: monitoramento completo
- ✅ **Health Checks**: verificação de saúde
- ✅ **Docker**: containerização completa
- ✅ **Logs estruturados**: JSON formatado
- ✅ **Graceful Shutdown**: parada segura

### 🔄 Compatibilidade Evolution API
- ✅ `POST /instance/create`
- ✅ `GET /instance/fetchInstances`
- ✅ `GET /instance/connect/:name`
- ✅ `GET /instance/connectionState/:name`
- ✅ `DELETE /instance/logout/:name`
- ✅ `DELETE /instance/delete/:name`
- ✅ `POST /message/sendText/:name`
- ✅ `POST /message/sendMedia/:name`
- ✅ `POST /message/sendImage/:name`
- ✅ `POST /message/sendVideo/:name`
- ✅ `POST /message/sendAudio/:name`
- ✅ `POST /message/sendDocument/:name`

## 📈 Performance

### Benchmarks Esperados
- **~3-5x mais rápido** que Node.js
- **~50-70% menos memória** que Node.js
- **Suporte a milhares** de instâncias simultâneas
- **Latência < 50ms** para envio de mensagens

### Escalabilidade
- **Horizontal**: múltiplas instâncias da API
- **Vertical**: otimizado para Go + Goroutines
- **Cache**: Redis distribuído
- **Banco**: PostgreSQL com pooling

## 🛡️ Segurança

- **Autenticação**: API Key obrigatória
- **Rate Limiting**: proteção contra DDoS
- **Validação**: entrada sanitizada
- **Logs**: sem dados sensíveis
- **Headers**: segurança HTTP

## 🚨 Troubleshooting

### Problemas Comuns

**1. Erro de conexão com banco**
```bash
# Verificar se PostgreSQL está rodando
docker-compose ps

# Ver logs do banco
docker-compose logs postgres
```

**2. Erro de conexão com Redis**
```bash
# Verificar Redis
docker-compose logs redis

# Testar conexão
redis-cli ping
```

**3. Instância não conecta**
```bash
# Verificar status
curl -H "apikey: your-key" http://localhost:8080/api/instance/minha-instancia/status

# Reconectar
curl -X POST -H "apikey: your-key" http://localhost:8080/api/instance/minha-instancia/connect
```

**4. Webhook não funciona**
```bash
# Verificar logs
docker-compose logs evolution-api | grep webhook

# Testar webhook manualmente
curl -X POST https://meu-site.com/webhook \
  -H "Content-Type: application/json" \
  -d '{"test": "webhook"}'
```

## 📞 Suporte

- **Issues**: Abra uma issue no GitHub
- **Documentação**: Consulte README.md
- **Logs**: Sempre verifique os logs primeiro

---

**🎉 Parabéns! Sua Evolution API Go está funcionando!**

Agora você tem uma API WhatsApp de alta performance, escalável e pronta para produção! 🚀
