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