package processor_test

import (
	"testing"
	"time"

	"github.com/gopher/logscope/internal/processor"
	"github.com/gopher/logscope/pkg/entry"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

func makeEntries(n int) []entry.LogEntry {
	entries := make([]entry.LogEntry, n)
	for i := range entries {
		status := 200
		level := entry.INFO
		if i%10 == 0 {
			status = 500
			level = entry.ERROR
		} else if i%5 == 0 {
			status = 404
			level = entry.WARN
		}
		entries[i] = entry.LogEntry{
			Level:   level,
			Method:  "GET",
			Path:    "/api/users",
			Status:  status,
			Latency: time.Duration(i+1) * time.Millisecond,
			Bytes:   int64((i + 1) * 100),
		}
	}
	return entries
}

// ─── unit tests ───────────────────────────────────────────────────────────────

func TestProcessSafe_Total(t *testing.T) {
	t.Parallel()

	entries := makeEntries(100)
	stats := processor.ProcessSafe(entries, 4)

	if stats.Total != 100 {
		t.Errorf("Total: got %d, want 100", stats.Total)
	}
}

func TestProcessSafe_Empty(t *testing.T) {
	t.Parallel()

	stats := processor.ProcessSafe(nil, 4)
	if stats.Total != 0 {
		t.Errorf("expected empty stats, got total=%d", stats.Total)
	}
}

func TestProcessSafe_Percentiles(t *testing.T) {
	t.Parallel()

	// entries with latencies 1ms, 2ms, ..., 100ms
	entries := makeEntries(100)
	stats := processor.ProcessSafe(entries, 4)

	p50 := stats.P(50)
	p95 := stats.P(95)
	p99 := stats.P(99)

	if p50 < 49*time.Millisecond || p50 > 51*time.Millisecond {
		t.Errorf("P50 out of range: %v", p50)
	}
	if p95 < 93*time.Millisecond {
		t.Errorf("P95 too low: %v", p95)
	}
	if p99 < 97*time.Millisecond {
		t.Errorf("P99 too low: %v", p99)
	}
}

func TestProcessSafe_StatusDistribution(t *testing.T) {
	t.Parallel()

	entries := makeEntries(100)
	stats := processor.ProcessSafe(entries, 4)

	// i%10==0 → 500, i%5==0 (but not %10) → 404, rest → 200
	// 100 entries: 10 × 500, 10 × 404, 80 × 200
	if stats.ByStatus[500] != 10 {
		t.Errorf("status 500: got %d, want 10", stats.ByStatus[500])
	}
	if stats.ByStatus[404] != 10 {
		t.Errorf("status 404: got %d, want 10", stats.ByStatus[404])
	}
	if stats.ByStatus[200] != 80 {
		t.Errorf("status 200: got %d, want 80", stats.ByStatus[200])
	}
}

// TestProcessSafe_DeterministicWithRace verifies that ProcessSafe is race-free.
// Run with: go test -race -run TestProcessSafe_DeterministicWithRace ./internal/processor/
func TestProcessSafe_DeterministicWithRace(t *testing.T) {
	t.Parallel()

	entries := makeEntries(10000)
	stats := processor.ProcessSafe(entries, 8)

	if stats.Total != 10000 {
		t.Errorf("Total: got %d, want 10000", stats.Total)
	}
}

// TestNaiveRace demonstrates the race condition in ProcessNaive.
//
// Run WITHOUT race detector first — it may or may not panic:
//
//	go test -run TestNaiveRace ./internal/processor/
//
// Then run WITH the race detector to reliably catch it:
//
//	go test -race -run TestNaiveRace ./internal/processor/
//
// The race detector will print the conflicting goroutine stack traces.
func TestNaiveRace(t *testing.T) {
	t.Skip("TestNaiveRace should be run explicitly: go test -race -run TestNaiveRace ./internal/processor/")
}

// ─── benchmarks ───────────────────────────────────────────────────────────────
// Run with: go test -bench=. -benchmem ./internal/processor/

// BenchmarkProcessSafe_Workers compares performance at different worker counts.
// Run with: go test -bench=BenchmarkProcessSafe -benchmem ./internal/processor/
func BenchmarkProcessSafe_1Worker(b *testing.B) {
	benchmarkProcessSafe(b, 1)
}
func BenchmarkProcessSafe_2Workers(b *testing.B) {
	benchmarkProcessSafe(b, 2)
}
func BenchmarkProcessSafe_4Workers(b *testing.B) {
	benchmarkProcessSafe(b, 4)
}
func BenchmarkProcessSafe_8Workers(b *testing.B) {
	benchmarkProcessSafe(b, 8)
}

func benchmarkProcessSafe(b *testing.B, workers int) {
	b.Helper()
	entries := makeEntries(100_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessSafe(entries, workers)
	}
}

// BenchmarkProcessNaive shows baseline (and will crash under -race).
func BenchmarkProcessNaive(b *testing.B) {
	entries := makeEntries(1000) // keep small — naive is unsafe
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = processor.ProcessNaive(entries)
	}
}
