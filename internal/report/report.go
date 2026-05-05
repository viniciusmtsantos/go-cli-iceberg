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
		fmt.Fprintf(tw, "    %d:\t%s  %d\n", s, bar, stats.ByStatus[s])
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
