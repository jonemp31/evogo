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
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/prometheus/client_golang v1.19.1
	go.uber.org/zap v1.27.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
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
