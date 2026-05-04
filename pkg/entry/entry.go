// Package entry defines the core data types used throughout logscope.
package entry

import "time"

// Level represents the severity of a log entry.
//
// The String() method is generated automatically by stringer.
// Regenerate with:
//
//	go generate ./pkg/entry/
//
//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=Level -output=level_string.go
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

// LogEntry represents a single parsed HTTP access log line.
//
// Log format (space-separated):
//
//	<timestamp> <level> <method> <path> <status> <latency_ms> <bytes>
//
// Example:
//
//	2026-04-26T10:00:00Z INFO GET /api/users 200 145 1234
type LogEntry struct {
	Timestamp time.Time
	Level     Level
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	Bytes     int64
}
