#!/usr/bin/env bash
# reset.sh — restaura o projeto ao estado inicial para treino da apresentação
set -euo pipefail

echo "[reset] removendo binários compilados..."
rm -f goprobe goprobe.exe

echo "[reset] removendo go.sum..."
rm -f go.sum

echo "[reset] restaurando go.mod ao estado inicial (sem dependências)..."
cat > go.mod << 'EOF'
module github.com/casadebackend/goprobe

go 1.22
EOF

echo "[reset] limpando cache de build local..."
go clean -cache 2>/dev/null || true

echo ""
echo "✓ Pronto! Estado restaurado. Você pode começar do zero:"
echo ""
echo "  1.  go version"
echo "  2.  go version -m -json \$(which docker)"
echo "  3.  go mod init github.com/casadebackend/goprobe   # (já feito; apenas demonstre o go.mod)"
echo "  4.  go get github.com/fatih/color@latest"
echo "  5.  go mod tidy"
echo "  6.  go run . version"
echo "  7.  go run golang.org/x/vuln/cmd/govulncheck@latest ."
echo "  8.  GOOS=windows go build -o goprobe.exe .         # mostra build constraint"
echo "  9.  go build -o goprobe ."
echo "  10. go version -m ./goprobe"
echo "  11. go install -n .                                  # mostra o que seria feito"
echo "  11. go install .                                    # instala goprobe em \$GOPATH/bin"
echo "  12. go test ./...                                   # testes"
echo "  13. go test -bench=. ./internal/checker/"
echo ""
