#!/bin/bash
echo "🧹 Limpando arquivos duplicados..."
# PASSO 1: Remove todos os arquivos de serviço com sufixos conflitantes.
rm -f internal/services/*_corrected.go
rm -f internal/services/*_fixed.go
rm -f internal/services/*_evolution.go
echo "✅ Arquivos duplicados removidos."

echo "🚀 Corrigindo caminhos de importação..."
# PASSO 2: Encontra todos os arquivos .go e substitui o caminho do módulo antigo pelo novo
find . -type f -name "*.go" -exec sed -i 's|github.com/evolution-api/evolution-go|github.com/jonemp31/evogo|g' {} +

echo "✅ Caminhos de importação corrigidos."
echo "🧹 Limpando módulos e atualizando dependências..."

# PASSO 3: Limpa o cache e garante que todas as dependências corretas sejam baixadas
go mod tidy

echo "🛠️ Tentando compilar o projeto..."
# PASSO 4: Compila o projeto
go build -o evolution-api-go ./cmd/server

# PASSO 5: Verifica o resultado da compilação
if [ $? -eq 0 ]; then
  echo "✅ 🎉 SUCESSO! A compilação foi concluída. O executável 'evolution-api-go' foi criado."
  echo "👉 Próximo passo: execute ./evolution-api-go para iniciar a API!"
else
  echo "❌ FALHA. A compilação falhou. Verifique os erros acima."
fi
