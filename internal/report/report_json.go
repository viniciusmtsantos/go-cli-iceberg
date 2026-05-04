//go:build json

// This file replaces report.go when built with: go build -tags json ./cmd/logscope
//
// Build tags allow selecting entirely different implementations at compile time,
// with zero runtime overhead — the unused implementation is not in the binary.
//
// Demonstrate with:
//
//	go build ./cmd/logscope          # text output (default)
//	go build -tags json ./cmd/logscope  # JSON output
//	./logscope -input access.log | jq .
package report

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gopher/logscope/internal/processor"
)

// jsonReport is the JSON schema for the output.
type jsonReport struct {
	Total      int            `json:"total"`
	TotalBytes int64          `json:"total_bytes_kb"`
	Latency    latencyReport  `json:"latency"`
	ByStatus   map[int]int    `json:"by_status"`
	ByLevel    map[string]int `json:"by_level"`
	ByPath     map[string]int `json:"by_path"`
}

type latencyReport struct {
	AvgMs float64 `json:"avg_ms"`
	P50Ms float64 `json:"p50_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

// Write outputs a JSON-formatted report to w.
// Activate with: go build -tags json ./cmd/logscope
func Write(w io.Writer, stats *processor.Stats) {
	report := jsonReport{
		Total:      stats.Total,
		TotalBytes: stats.TotalBytes / 1024,
		Latency: latencyReport{
			AvgMs: msFloat(stats.Avg()),
			P50Ms: msFloat(stats.P(50)),
			P95Ms: msFloat(stats.P(95)),
			P99Ms: msFloat(stats.P(99)),
		},
		ByStatus: stats.ByStatus,
		ByLevel:  stats.ByLevel,
		ByPath:   stats.ByPath,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(report)
}

func msFloat(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
