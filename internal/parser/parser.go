// Package parser reads and parses HTTP access log lines into LogEntry values.
package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gopher/logscope/pkg/entry"
)

// Parse parses a single log line into a LogEntry.
//
// Expected format (space-separated, 7 fields):
//
//	<timestamp> <level> <method> <path> <status> <latency_ms> <bytes>
//
// Returns an error for empty, malformed, or semantically invalid lines.
func Parse(line string) (entry.LogEntry, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return entry.LogEntry{}, fmt.Errorf("empty line")
	}

	fields := strings.Fields(line)
	if len(fields) != 7 {
		return entry.LogEntry{}, fmt.Errorf("expected 7 fields, got %d", len(fields))
	}

	ts, err := time.Parse(time.RFC3339, fields[0])
	if err != nil {
		return entry.LogEntry{}, fmt.Errorf("invalid timestamp %q: %w", fields[0], err)
	}

	lvl, err := parseLevel(fields[1])
	if err != nil {
		return entry.LogEntry{}, err
	}

	status, err := strconv.Atoi(fields[4])
	if err != nil {
		return entry.LogEntry{}, fmt.Errorf("invalid status %q: %w", fields[4], err)
	}
	if status < 100 || status > 599 {
		return entry.LogEntry{}, fmt.Errorf("status out of range: %d", status)
	}

	latencyMs, err := strconv.Atoi(fields[5])
	if err != nil {
		return entry.LogEntry{}, fmt.Errorf("invalid latency %q: %w", fields[5], err)
	}

	bytes, err := strconv.ParseInt(fields[6], 10, 64)
	if err != nil {
		return entry.LogEntry{}, fmt.Errorf("invalid bytes %q: %w", fields[6], err)
	}

	return entry.LogEntry{
		Timestamp: ts,
		Level:     lvl,
		Method:    fields[2],
		Path:      fields[3],
		Status:    status,
		Latency:   time.Duration(latencyMs) * time.Millisecond,
		Bytes:     bytes,
	}, nil
}

// ParseReader reads all log lines from r and returns valid entries.
// Malformed lines are silently skipped to handle real-world noisy logs.
func ParseReader(r io.Reader) ([]entry.LogEntry, error) {
	const maxLineSize = 1024 * 1024 // 1MB per line limit

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, maxLineSize), maxLineSize)

	var entries []entry.LogEntry
	for scanner.Scan() {
		e, err := Parse(scanner.Text())
		if err != nil {
			continue // skip malformed lines gracefully
		}
		entries = append(entries, e)
	}
	return entries, scanner.Err()
}

func parseLevel(s string) (entry.Level, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return entry.DEBUG, nil
	case "INFO":
		return entry.INFO, nil
	case "WARN":
		return entry.WARN, nil
	case "ERROR":
		return entry.ERROR, nil
	case "FATAL":
		return entry.FATAL, nil
	default:
		return 0, fmt.Errorf("unknown log level: %q", s)
	}
}
