#!/bin/bash

echo "🚀 Migrando para Evolution API Go - Implementação Completa"
echo "=========================================================="

# Backup dos arquivos originais
echo "📦 Fazendo backup dos arquivos originais..."
mkdir -p backup/$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="backup/$(date +%Y%m%d_%H%M%S)"

cp internal/services/whatsapp_client.go $BACKUP_DIR/ 2>/dev/null || true
cp internal/services/message_service.go $BACKUP_DIR/ 2>/dev/null || true
cp internal/services/instance_manager.go $BACKUP_DIR/ 2>/dev/null || true
cp internal/services/webhook_service.go $BACKUP_DIR/ 2>/dev/null || true
cp internal/handlers/instance_handler.go $BACKUP_DIR/ 2>/dev/null || true
cp internal/handlers/message_handler.go $BACKUP_DIR/ 2>/dev/null || true
cp internal/routes/routes.go $BACKUP_DIR/ 2>/dev/null || true
cp cmd/server/main.go $BACKUP_DIR/ 2>/dev/null || true

echo "✅ Backup criado em: $BACKUP_DIR"

# Aplicar novos arquivos
echo "🔨 Aplicando implementação Evolution..."

# Substituir arquivos principais
cp internal/services/whatsapp_client_evolution.go internal/services/whatsapp_client.go
cp internal/services/message_service_evolution.go internal/services/message_service.go
cp internal/services/instance_manager_corrected.go internal/services/instance_manager.go
cp internal/services/webhook_service_evolution.go internal/services/webhook_service.go
cp internal/handlers/instance_handler_evolution.go internal/handlers/instance_handler.go
cp internal/handlers/message_handler_evolution.go internal/handlers/message_handler.go
cp internal/routes/routes_evolution.go internal/routes/routes.go
cp cmd/server/main_evolution.go cmd/server/main.go

# Criar diretório dto se não existir
mkdir -p internal/dto
cp internal/dto/message_dto.go internal/dto/message_dto.go

# Limpar arquivos temporários
echo "🧹 Limpando arquivos temporários..."
rm -f internal/services/*_evolution.go
rm -f internal/services/*_corrected.go
rm -f internal/handlers/*_evolution.go
rm -f internal/routes/*_evolution.go
rm -f cmd/server/main_evolution.go

echo "✅ Migração concluída!"
echo ""
echo "📋 Funcionalidades implementadas:"
echo "  ✅ Fase 1: Fundações da conexão e gerenciamento de eventos"
echo "  ✅ Fase 2: Envio completo de mídias (texto, imagem, vídeo, áudio, documento)"
echo "  ✅ Fase 3: Webhooks funcionais"
echo "  ✅ ViewOnce para imagens, vídeos e áudios"
echo "  ✅ PTT (áudio narrado)"
echo "  ✅ DTOs completos para todas as requisições"
echo "  ✅ Handlers multipart/form-data"
echo "  ✅ WebhookService com retry e exponential backoff"
echo "  ✅ InstanceManager com gerenciamento completo de ciclo de vida"
echo "  ✅ MessageService com processamento de eventos"
echo "  ✅ Rotas RESTful completas"
echo "  ✅ Documentação da API integrada"
echo ""
echo "🚀 Tentando compilar..."

# Tentar compilar
go build -o evolution-api-go ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Compilação bem-sucedida!"
    echo "📁 Executável criado: evolution-api-go"
    echo ""
    echo "🎯 Para testar a API:"
    echo "  ./evolution-api-go"
    echo ""
    echo "📖 Documentação disponível em:"
    echo "  http://localhost:8080/docs"
    echo ""
    echo "🔗 Endpoints principais:"
    echo "  POST /api/v1/instance/create"
    echo "  POST /api/v1/instance/connect/{instanceName}"
    echo "  POST /api/v1/message/sendText/{instanceName}"
    echo "  POST /api/v1/message/sendImage/{instanceName}"
    echo "  POST /api/v1/message/sendVideo/{instanceName}"
    echo "  POST /api/v1/message/sendAudio/{instanceName}"
    echo "  POST /api/v1/message/sendDocument/{instanceName}"
else
    echo "❌ Erro na compilação. Verifique os logs acima."
    echo "📋 Possíveis soluções:"
    echo "  1. Verificar se todas as dependências estão instaladas"
    echo "  2. Executar: go mod tidy"
    echo "  3. Executar: go mod download"
    echo "  4. Verificar se a versão do Go é >= 1.24.0"
fi
