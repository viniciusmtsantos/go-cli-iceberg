#!/usr/bin/env bash
# reset.sh — restaura o logscope ao estado inicial para treino da apresentação.
# Introduz propositalmente: (1) formatação quebrada em processor.go e (2) bug de vet em report.go.
set -euo pipefail

echo "[reset] removendo binários e arquivos gerados..."
rm -f logscope logscope.exe logscope-json logscope-arm64 logscope-safe

echo "[reset] removendo artefatos de profiling e análise..."
rm -f cpu.prof mem.prof trace.out coverage.out access.log

echo "[reset] limpando cache de testes..."
go clean -testcache 2>/dev/null || true

echo "[reset] restaurando go.mod ao estado sem dependências externas..."
cat > go.mod << 'EOF'
module github.com/gopher/logscope

go 1.24
EOF
rm -f go.sum

echo "[reset] restaurando arquivos de código ao estado limpo do git..."
REPO_ROOT=$(git rev-parse --show-toplevel)
git -C "$REPO_ROOT" checkout HEAD -- internal/processor/processor.go
git -C "$REPO_ROOT" checkout HEAD -- internal/report/report.go
git -C "$REPO_ROOT" checkout HEAD -- internal/parser/parser_test.go
git -C "$REPO_ROOT" checkout HEAD -- internal/processor/naive.go

echo "[reset] introduzindo bug de vet em report.go (Entries analyzed: %d → %s)..."
python3 - << 'PYEOF'
content = open("internal/report/report.go").read()
old = 'fmt.Fprintf(tw, "  Entries analyzed:\\t%d\\n", stats.Total)'
new = 'fmt.Fprintf(tw, "  Entries analyzed:\\t%s\\n", stats.Total)'
result = content.replace(old, new)
if result == content:
    print("AVISO: substituição do vet bug não encontrou o trecho esperado — report.go pode já estar com bug ou formatação diferente")
else:
    open("internal/report/report.go", "w").write(result)
    print("ok")
PYEOF

echo "[reset] introduzindo formatação quebrada em processor.go (função Avg)..."
python3 - << 'PYEOF'
content = open("internal/processor/processor.go").read()
old = '''// Avg returns the mean latency across all entries.
func (s *Stats) Avg() time.Duration {
\tif len(s.Latencies) == 0 {
\t\treturn 0
\t}
\tvar total time.Duration
\tfor _, l := range s.Latencies {
\t\ttotal += l
\t}
\treturn total / time.Duration(len(s.Latencies))
}'''
new = '''// Avg returns the mean latency across all entries.
func (s *Stats) Avg() time.Duration {
if len(s.Latencies) == 0 {return 0}
var total time.Duration
for _,l := range s.Latencies {total += l}
return total/time.Duration(len(s.Latencies))
}'''
result = content.replace(old, new)
if result == content:
    print("AVISO: substituição do fmt bug não encontrou o trecho esperado — processor.go pode já estar com bug ou formatação diferente")
else:
    open("internal/processor/processor.go", "w").write(result)
    print("ok")
PYEOF

echo ""
echo "✓ Pronto! Estado inicial restaurado com dois bugs intencionais:"
echo ""
echo "  ① processor.go → Avg() sem indentação (go fmt vai corrigir)"
echo "  ② report.go    → Entries analyzed: %s para int stats.Total (go vet vai detectar)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "SEQUÊNCIA DA APRESENTAÇÃO"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🌊 SUPERFÍCIE"
echo "  1.  go help / go help <cmd>                         # abertura"
echo "  2.  go version"
echo "  3.  go version -m -json \$(which docker)"
echo "  4.  cat go.mod"
echo "  5.  go get github.com/fatih/color@latest"
echo "  6.  cat go.mod && cat go.sum | head -5"
echo "  7.  go run ./cmd/logscope -gen -lines 1000 -output testdata/access.log"
echo "  8.  go run ./cmd/logscope -input testdata/access.log"
echo "  9.  go run golang.org/x/vuln/cmd/govulncheck@latest ./..."
echo "  10. go build -o logscope ./cmd/logscope && ls -lh logscope"
echo "  11. ./logscope -gen -lines 100 | ./logscope"
echo "  12. go version -m ./logscope"
echo "  13. go env GOPATH && go install -n ./cmd/logscope"
echo "  14. go install ./cmd/logscope"
echo "  15. \$(go env GOPATH)/bin/logscope -version"
echo ""
echo "🔎 CAMADA 2 — QUALIDADE"
echo "  16. go doc fmt.Printf"
echo "  17. go doc ./pkg/entry LogEntry"
echo "  18. go doc ./internal/parser ParseReader"
echo "  19. go doc ./internal/processor ProcessSafe"
echo "  20. gofmt -l ."
echo "  21. gofmt -d internal/processor/processor.go"
echo "  22. go fmt ./... && gofmt -l ."
echo "  23. go vet ./..."
echo "  24. # corrige report.go: %s → %d em Entries analyzed"
echo "  25. go vet ./..."
echo "  26. go test -short ./...                    # agora passa (vet bug corrigido no passo 25)"
echo "  27. go test -v ./internal/parser/"
echo "  28. go env && go env -json GOPATH GOCACHE GOMODCACHE"
echo "  29. go env -w GOFLAGS=-trimpath && cat \$(go env GOENV)"
echo "  30. go env -u GOFLAGS"
echo "  31. go list ./... && go list -m -u all | grep '\\['"
echo "  32. go mod why golang.org/x/sys && go mod why github.com/fatih/color"
echo "  33. go test -short -cover ./internal/..."
echo "  34. go test -short -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out"
echo "  35. go tool cover -html=coverage.out"
echo "  36. go mod tidy && cat go.mod    # remove color (não usada no código)"
echo "  37. go clean -testcache && go test -short ./internal/..."
echo "  38. go clean -cache && go clean -modcache"
echo ""
echo "⚙️  CAMADA 3 — BUILD AVANÇADO E GERAÇÃO"
echo "  39. cat pkg/entry/entry.go      # mostra //go:generate"
echo "  40. rm pkg/entry/level_string.go && go generate ./pkg/entry/ && cat pkg/entry/level_string.go"
echo "  41. cat internal/report/report.go && cat internal/report/report_json.go"
echo "  42. go build -tags json -o logscope-json ./cmd/logscope"
echo "  43. ./logscope-json -gen -lines 100 | ./logscope-json"
echo "  44. go build -ldflags \"-X main.version=1.2.0 -X main.commit=\$(git rev-parse --short HEAD)\" -o logscope ./cmd/logscope"
echo "  45. ./logscope -version && go version -m ./logscope"
echo "  46. GOOS=windows go build -o logscope.exe ./cmd/logscope && file logscope.exe"
echo "  47. GOOS=linux GOARCH=arm64 go build -o logscope-arm64 ./cmd/logscope && file logscope-arm64"
echo "  48. go build -trimpath -o logscope-safe ./cmd/logscope"
echo "  49. strings ./logscope-safe | grep \$(pwd) || echo '✓ nenhum path local'"
echo ""
echo "🔬 CAMADA 4 — TESTING PROFUNDO E OBSERVABILIDADE"
echo "  50. go test -bench=. -benchmem ./internal/parser/"
echo "  51. go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/"
echo "  52. go test -run TestNaiveRace ./internal/processor/"
echo "  53. go test -race -run TestNaiveRace ./internal/processor/"
echo "  54. go test -race -run TestProcessSafe_DeterministicWithRace ./internal/processor/"
echo "  55. go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/"
echo "  56. go test -shuffle=on -v ./internal/parser/"
echo "  57. go test -count=3 ./internal/..."
echo "  58. logscope -gen -lines 200000 -output access.log"
echo "  59. logscope -input access.log -cpuprofile cpu.prof && go tool pprof -http=:8080 cpu.prof"
echo "  60. logscope -input access.log -memprofile mem.prof && go tool pprof -http=:8081 mem.prof"
echo "  61. logscope -input access.log -trace trace.out && go tool trace trace.out"
echo "  62. go tool nm ./logscope | head -20"
echo "  63. go tool nm ./logscope | wc -l"
echo ""
