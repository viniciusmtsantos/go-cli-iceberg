#!/usr/bin/env bash
# reset.sh — restaura o logscope ao estado inicial para treino da apresentação.
# Introduz propositalmente: (1) formatação quebrada em processor.go e (2) bug de vet em report.go.
set -euo pipefail

echo "[reset] limpando cache do módulo para evitar conflitos de checksum..."
go clean -modcache 2>/dev/null || true

echo "[reset] removendo binários e arquivos gerados..."
rm -f logscope logscope.exe logscope-json logscope-arm64 logscope-safe

echo "[reset] removendo artefatos de profiling e análise..."
rm -f cpu.prof mem.prof trace.out coverage.out access.log

echo "[reset] limpando cache de testes..."
go clean -testcache 2>/dev/null || true

echo "[reset] restaurando go.mod com dependência fatih/color..."
cat > go.mod << 'EOF'
module github.com/gopher/logscope

go 1.25.0

require (
	github.com/fatih/color v1.16.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.14.0 // indirect
)
EOF

cat > go.sum << 'EOF'
github.com/fatih/color v1.16.0 h1:zmkK9Ngbjj+K0yRhTVONQh1p/HknKYSlNT+vZCzyokM=
github.com/fatih/color v1.16.0/go.mod h1:fL2Sau1YI5c0pdGEVCbKQbLXB6edEj1ZgiY4NijnWvE=
github.com/mattn/go-colorable v0.1.13 h1:fFA4WZxdEF4tXPZVKMLwD8oUnCTTo08duU7wxecdEvA=
github.com/mattn/go-colorable v0.1.13/go.mod h1:7S9/ev0klgBDR4GtXTXX8a3vIGJpMovkB8vQcUbaXHg=
github.com/mattn/go-isatty v0.0.16/go.mod h1:kYGgaQfpe5nmfYZH+SKPsOc2e4SrIfOl2e/yFXSvRLM=
github.com/mattn/go-isatty v0.0.20 h1:xfD0iDuEKnDkl03q4limB+vH+GxLEtL/jb4xVJSWWEY=
github.com/mattn/go-isatty v0.0.20/go.mod h1:W+V8PltTTMOvKvAeJH7IuucS94S2C6jfK/D7dTCTo3Y=
golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.6.0/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
golang.org/x/sys v0.14.0 h1:Vz7Qs629MkJkGyHxUlRHizWJRG2j8fbQKjELVSNhy7Q=
golang.org/x/sys v0.14.0/go.mod h1:/VUhepiaJMQUp4+oa/7Zr1D23ma6VTLIYjOOTFZPUcA=
EOF

echo "[reset] verificando e corrigindo checksums se necessário..."
go mod download 2>/dev/null || true

echo "[reset] restaurando arquivos de código ao estado limpo do git..."
REPO_ROOT=$(git rev-parse --show-toplevel)
git -C "$REPO_ROOT" checkout HEAD -- internal/processor/processor.go
git -C "$REPO_ROOT" checkout HEAD -- internal/parser/parser_test.go
git -C "$REPO_ROOT" checkout HEAD -- internal/processor/naive.go

echo "[reset] restaurando report.go com color integration..."
cat > internal/report/report.go << 'GOEOF'
//go:build !json

// Package report writes aggregated log statistics to an output stream.
//
// Two implementations exist, selected at build time via build tags:
//   - Text (default):  go build ./cmd/logscope
//   - JSON:            go build -tags json ./cmd/logscope
package report

import (
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/gopher/logscope/internal/processor"
)

// Write outputs a human-readable text report to w.
func Write(w io.Writer, stats *processor.Stats) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	sep := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	fmt.Fprintln(tw, sep)
	fmt.Fprintln(tw, "  LOGSCOPE — HTTP Access Log Report")
	fmt.Fprintln(tw, sep)
	fmt.Fprintln(tw)

	fmt.Fprintf(tw, "  Entries analyzed:\t%d\n", stats.Total)
	fmt.Fprintf(tw, "  Total traffic:\t%s\n", formatBytes(stats.TotalBytes))
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, "  Latency")
	fmt.Fprintf(tw, "    avg:\t%v\n", stats.Avg())
	fmt.Fprintf(tw, "    p50:\t%v\n", stats.P(50))
	fmt.Fprintf(tw, "    p95:\t%v\n", stats.P(95))
	fmt.Fprintf(tw, "    p99:\t%v\n", stats.P(99))
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, "  Status Codes")
	for _, s := range sortedIntKeys(stats.ByStatus) {
		bar := progressBar(stats.ByStatus[s], stats.Total, 20)
		statusColor := getStatusColor(s)
		fmt.Fprintf(tw, "    %s:\t%s  %d\n", statusColor.Sprint(s), bar, stats.ByStatus[s])
	}
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, "  Log Levels")
	for _, l := range sortedStringKeys(stats.ByLevel) {
		bar := progressBar(stats.ByLevel[l], stats.Total, 20)
		fmt.Fprintf(tw, "    %-5s:\t%s  %d\n", l, bar, stats.ByLevel[l])
	}
	fmt.Fprintln(tw)

	fmt.Fprintln(tw, "  Top Endpoints")
	type kv struct {
		path  string
		count int
	}
	var paths []kv
	for p, c := range stats.ByPath {
		paths = append(paths, kv{p, c})
	}
	sort.Slice(paths, func(i, j int) bool { return paths[i].count > paths[j].count })
	limit := 10
	if len(paths) < limit {
		limit = len(paths)
	}
	for _, p := range paths[:limit] {
		bar := progressBar(p.count, stats.Total, 20)
		fmt.Fprintf(tw, "    %-35s\t%s  %d\n", p.path, bar, p.count)
	}
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, sep)
}

func progressBar(value, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := value * width / total
	bar := make([]rune, width)
	for i := range bar {
		if i < filled {
			bar[i] = '█'
		} else {
			bar[i] = '░'
		}
	}
	return string(bar)
}

func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func sortedIntKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func sortedStringKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getStatusColor(status int) *color.Color {
	switch {
	case status >= 200 && status < 300:
		return color.New(color.FgGreen)
	case status >= 300 && status < 400:
		return color.New(color.FgCyan)
	case status >= 400 && status < 500:
		return color.New(color.FgYellow)
	case status >= 500:
		return color.New(color.FgRed)
	default:
		return color.New(color.FgWhite)
	}
}
GOEOF

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
echo "✓ Pronto! Estado inicial restaurado:"
echo ""
echo "  ① processor.go → Avg() sem indentação (go fmt vai corrigir)"
echo "  ② report.go    → Com integração github.com/fatih/color nos status codes HTTP"
echo "  ③ go.mod       → Dependência fatih/color v1.16.0 incluída"
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
echo "  24. go test -short ./...                    # agora passa (vet bug já corrigido)"
echo "  25. go test -v ./internal/parser/"
echo "  26. go env && go env -json GOPATH GOCACHE GOMODCACHE"
echo "  27. go env -w GOFLAGS=-trimpath && cat \$(go env GOENV)"
echo "  28. go env -u GOFLAGS"
echo "  29. go list ./... && go list -m -u all"
echo "  30. go mod why golang.org/x/sys && go mod why github.com/fatih/color"
echo "  31. go test -short -cover ./internal/..."
echo "  32. go test -short -coverprofile=coverage.out ./internal/... && go tool cover -func=coverage.out"
echo "  33. go tool cover -html=coverage.out"
echo "  34. go mod tidy && cat go.mod    # mantém color (usada em report.go)"
echo "  35. go clean -testcache && go test -short ./internal/..."
echo "  36. go clean -cache && go clean -modcache"
echo ""
echo "⚙️  CAMADA 3 — BUILD AVANÇADO E GERAÇÃO"
echo "  37. cat pkg/entry/entry.go      # mostra //go:generate"
echo "  38. rm pkg/entry/level_string.go && go generate ./pkg/entry/ && cat pkg/entry/level_string.go"
echo "  39. cat internal/report/report.go && cat internal/report/report_json.go"
echo "  40. go build -tags json -o logscope-json ./cmd/logscope"
echo "  41. ./logscope-json -gen -lines 100 | ./logscope-json"
echo "  42. go build -ldflags \"-X main.version=1.2.0 -X main.commit=\$(git rev-parse --short HEAD)\" -o logscope ./cmd/logscope"
echo "  43. ./logscope -version && go version -m ./logscope"
echo "  44. GOOS=windows go build -o logscope.exe ./cmd/logscope && file logscope.exe"
echo "  45. GOOS=linux GOARCH=arm64 go build -o logscope-arm64 ./cmd/logscope && file logscope-arm64"
echo "  46. go build -trimpath -o logscope-safe ./cmd/logscope"
echo "  47. strings ./logscope-safe | grep \$(pwd) || echo '✓ nenhum path local'"
echo ""
echo "🔬 CAMADA 4 — TESTING PROFUNDO E OBSERVABILIDADE"
echo "  48. go test -bench=. -benchmem ./internal/parser/"
echo "  49. go test -run=^$ -bench=BenchmarkProcessSafe -benchmem ./internal/processor/"
echo "  50. go test -run TestNaiveRace ./internal/processor/"
echo "  51. go test -race -run TestNaiveRace ./internal/processor/"
echo "  52. go test -race -run TestProcessSafe_DeterministicWithRace ./internal/processor/"
echo "  53. go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/"
echo "  54. go test -shuffle=on -v ./internal/parser/"
echo "  55. go test -count=3 ./internal/..."
echo "  56. logscope -gen -lines 200000 -output access.log"
echo "  57. logscope -input access.log -cpuprofile cpu.prof && go tool pprof -http=:8080 cpu.prof"
echo "  58. logscope -input access.log -memprofile mem.prof && go tool pprof -http=:8081 mem.prof"
echo "  59. logscope -input access.log -trace trace.out && go tool trace trace.out"
echo "  60. go tool nm ./logscope | head -20"
echo "  61. go tool nm ./logscope | wc -l"
echo ""
