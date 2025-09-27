#!/bin/bash

echo "🔧 Aplicando correções nos arquivos..."

# Backup dos arquivos originais
echo "📦 Fazendo backup dos arquivos originais..."
cp internal/services/whatsapp_client.go internal/services/whatsapp_client.go.bak
cp internal/services/message_service.go internal/services/message_service.go.bak
cp internal/services/instance_manager.go internal/services/instance_manager.go.bak

# Aplicar correções
echo "🔨 Aplicando correções..."

# Substituir arquivos corrigidos
cp internal/services/whatsapp_client_fixed.go internal/services/whatsapp_client.go
cp internal/services/message_service_fixed.go internal/services/message_service.go
cp internal/services/instance_manager_fixed.go internal/services/instance_manager.go

# Remover arquivos temporários
rm internal/services/whatsapp_client_fixed.go
rm internal/services/message_service_fixed.go
rm internal/services/instance_manager_fixed.go

echo "✅ Correções aplicadas com sucesso!"
echo "🚀 Tentando compilar..."

# Tentar compilar
go build -o evolution-api-go ./cmd/server

if [ $? -eq 0 ]; then
    echo "✅ Compilação bem-sucedida!"
    echo "📁 Executável criado: evolution-api-go"
else
    echo "❌ Erro na compilação. Verifique os logs acima."
fi
