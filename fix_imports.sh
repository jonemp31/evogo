#!/bin/bash
echo "🚀 Corrigindo caminhos de importação..."

# Encontra todos os arquivos .go e substitui o caminho do módulo antigo pelo novo
find . -type f -name "*.go" -exec sed -i 's|github.com/evolution-api/evolution-go|github.com/jonemp31/evogo|g' {} +

echo "✅ Caminhos de importação corrigidos."
echo "🧹 Limpando módulos e atualizando dependências..."

# Limpa o cache e garante que todas as dependências corretas sejam baixadas
go mod tidy

echo "🛠️ Tentando compilar o projeto..."
go build -o evolution-api-go ./cmd/server

# Verifica o resultado da compilação
if [ $? -eq 0 ]; then
  echo "✅ 🎉 SUCESSO! A compilação foi concluída. O executável 'evolution-api-go' foi criado."
  echo "👉 Próximo passo: execute ./evolution-api-go para iniciar a API!"
else
  echo "❌ FALHA. A compilação falhou. Verifique os erros acima."
fi
