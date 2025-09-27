#!/bin/bash

echo "🔧 Aplicando correções de compilação..."

# Fazer backup dos arquivos originais
echo "📦 Fazendo backup dos arquivos originais..."
mkdir -p backup/$(date +%Y%m%d_%H%M%S)
cp internal/services/instance_manager.go backup/$(date +%Y%m%d_%H%M%S)/ 2>/dev/null || true
cp internal/services/message_service.go backup/$(date +%Y%m%d_%H%M%S)/ 2>/dev/null || true
cp internal/services/whatsapp_client.go backup/$(date +%Y%m%d_%H%M%S)/ 2>/dev/null || true
cp go.mod backup/$(date +%Y%m%d_%H%M%S)/ 2>/dev/null || true

echo "✅ Backup criado em: backup/$(date +%Y%m%d_%H%M%S)"

# Limpar dependências antigas
echo "🧹 Limpando dependências antigas..."
go clean -cache
go clean -modcache
rm -f go.sum

# Atualizar go.mod para usar go-whatsapp
echo "📝 Atualizando go.mod..."
cat > go.mod << 'EOF'
module github.com/jonemp31/evogo

go 1.21

require (
	github.com/Rhymen/go-whatsapp v0.1.1
	github.com/gin-gonic/gin v1.9.1
	github.com/go-playground/validator/v10 v10.16.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.4.0
	github.com/joho/godotenv v1.4.0
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.17.0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	go.uber.org/zap v1.26.0
	golang.org/x/net v0.17.0
)
EOF

# Baixar dependências
echo "📥 Baixando dependências..."
go mod download
go mod tidy

# Tentar compilar
echo "🚀 Tentando compilar..."
if go build -o evolution-api-go ./cmd/server; then
    echo "✅ Compilação bem-sucedida!"
    echo "🎉 O executável 'evolution-api-go' foi criado com sucesso!"
else
    echo "❌ Erro na compilação. Verifique os logs acima."
    echo "📋 Possíveis soluções:"
    echo "1. Verificar se todas as dependências estão instaladas"
    echo "2. Executar: go mod tidy"
    echo "3. Executar: go mod download"
    echo "4. Verificar se a versão do Go é >= 1.21"
fi
