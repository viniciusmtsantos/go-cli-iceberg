// Package processor aggregates parsed log entries into statistical summaries.
package processor

import (
	"sort"
	"sync"
	"time"

	"github.com/gopher/logscope/pkg/entry"
)

// Stats holds aggregated statistics from a set of log entries.
type Stats struct {
	Total      int
	ByLevel    map[string]int
	ByStatus   map[int]int
	ByPath     map[string]int
	Latencies  []time.Duration // sorted ascending — used for percentile calculation
	TotalBytes int64
}

// P returns the nth percentile latency (0–100).
// P(50) = median, P(95) = p95, P(99) = p99.
func (s *Stats) P(percentile float64) time.Duration {
	if len(s.Latencies) == 0 {
		return 0
	}
	idx := int(float64(len(s.Latencies)-1) * percentile / 100.0)
	return s.Latencies[idx]
}

// Avg returns the mean latency across all entries.
func (s *Stats) Avg() time.Duration {
	if len(s.Latencies) == 0 {
		return 0
	}
	var total time.Duration
	for _, l := range s.Latencies {
		total += l
	}
	return total / time.Duration(len(s.Latencies))
}

// partialStats is the result from a single worker goroutine.
// Each worker has its own partialStats, so NO mutex is needed during processing.
// This is the key insight: eliminate shared state, merge at the end.
type partialStats struct {
	byLevel    map[string]int
	byStatus   map[int]int
	byPath     map[string]int
	latencies  []time.Duration
	totalBytes int64
}

// ProcessSafe aggregates log entries concurrently using a worker-per-chunk strategy.
//
// Design: each goroutine processes an exclusive slice of the entries and writes
// to its own local partialStats — no locks, no shared state during processing.
// The main goroutine merges all partial results after all workers finish.
//
// This pattern is idiomatic Go: communicate results via return values,
// not by sharing memory.
//
// Demonstrate race-free behaviour with:
//
//	go test -race ./internal/processor/
func ProcessSafe(entries []entry.LogEntry, workers int) *Stats {
	if len(entries) == 0 {
		return &Stats{
			ByLevel:  make(map[string]int),
			ByStatus: make(map[int]int),
			ByPath:   make(map[string]int),
		}
	}
	if workers <= 0 {
		workers = 1
	}
	if workers > len(entries) {
		workers = len(entries)
	}

	partials := make([]partialStats, workers)
	chunkSize := (len(entries) + workers - 1) / workers

	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)

		start := w * chunkSize
		end := start + chunkSize
		if end > len(entries) {
			end = len(entries)
		}
		chunk := entries[start:end]

		if start >= len(entries) {
			wg.Done()
			continue
		}

		go func(w int, chunk []entry.LogEntry) {
			defer wg.Done()

			p := partialStats{
				byLevel:   make(map[string]int, 5),
				byStatus:  make(map[int]int, 16),
				byPath:    make(map[string]int, 32),
				latencies: make([]time.Duration, 0, len(chunk)),
			}

			for _, e := range chunk {
				p.byLevel[e.Level.String()]++
				p.byStatus[e.Status]++
				p.byPath[e.Path]++
				p.latencies = append(p.latencies, e.Latency)
				p.totalBytes += e.Bytes
			}

			partials[w] = p
		}(w, chunk)
	}
	wg.Wait()

	// Merge: single goroutine, no synchronization needed
	stats := &Stats{
		Total:     len(entries),
		ByLevel:   make(map[string]int),
		ByStatus:  make(map[int]int),
		ByPath:    make(map[string]int),
		Latencies: make([]time.Duration, 0, len(entries)),
	}

	for _, p := range partials {
		for k, v := range p.byLevel {
			stats.ByLevel[k] += v
		}
		for k, v := range p.byStatus {
			stats.ByStatus[k] += v
		}
		for k, v := range p.byPath {
			stats.ByPath[k] += v
		}
		stats.Latencies = append(stats.Latencies, p.latencies...)
		stats.TotalBytes += p.totalBytes
	}

	sort.Slice(stats.Latencies, func(i, j int) bool {
		return stats.Latencies[i] < stats.Latencies[j]
	})

	return stats
}
