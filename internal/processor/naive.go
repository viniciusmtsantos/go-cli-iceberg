package processor

import (
	"sort"
	"sync"

	"github.com/gopher/logscope/pkg/entry"
)

// ProcessNaive aggregates log entries concurrently WITHOUT proper synchronization.
//
// WARNING: This implementation contains a deliberate race condition.
// Multiple goroutines write concurrently to shared maps and fields
// with no mutex protection — this is undefined behaviour in Go.
//
// This function exists solely to demonstrate the Go race detector.
//
// Detect the race with:
//
//	go test -race -run TestNaiveRace ./internal/processor/
//
// Or build the full binary with race detection enabled:
//
//	go build -race -o logscope-race ./cmd/logscope
//	./logscope-race -naive -input access.log
//
// Fix: see ProcessSafe in processor.go for the correct approach.
func ProcessNaive(entries []entry.LogEntry) *Stats {
	stats := &Stats{
		ByLevel:  make(map[string]int),
		ByStatus: make(map[int]int),
		ByPath:   make(map[string]int),
	}

	var wg sync.WaitGroup
	for _, e := range entries {
		wg.Add(1)
		e := e
		go func() {
			defer wg.Done()

			// RACE CONDITION: all goroutines write to the same maps
			// and fields with no synchronization. The race detector will
			// report "DATA RACE" with the conflicting goroutine stack traces.
			stats.Total++
			stats.ByLevel[e.Level.String()]++
			stats.ByStatus[e.Status]++
			stats.ByPath[e.Path]++
			stats.Latencies = append(stats.Latencies, e.Latency)
			stats.TotalBytes += e.Bytes
		}()
	}
	wg.Wait()

	sort.Slice(stats.Latencies, func(i, j int) bool {
		return stats.Latencies[i] < stats.Latencies[j]
	})

	return stats
}
