package parser_test

import (
	"strings"
	"testing"

	"github.com/gopher/logscope/internal/parser"
)

// FuzzParseReader verifies that the parser never panics on arbitrary input.
//
// Run with:
//
//	go test -fuzz=FuzzParseReader -fuzztime=10s ./internal/parser/
//
// The fuzzer mutates the seed corpus — inserting random bytes, truncating lines,
// injecting unicode — and reports any input that causes a panic.
func FuzzParseReader(f *testing.F) {
	// Seed corpus — representative log lines
	f.Add("2026-04-26T10:00:00Z INFO GET /api/users 200 145 1234\n")
	f.Add("2026-04-26T10:00:01Z WARN POST /api/auth/login 401 22 0\n")
	f.Add("2026-04-26T10:00:02Z ERROR DELETE /api/orders/99 500 310 0\n")
	f.Add("malformed line that should be skipped\n")
	f.Add("") // empty input — must return empty slice, not panic

	f.Fuzz(func(t *testing.T, data string) {
		// The parser must never panic regardless of input.
		// Errors are expected and acceptable; panics are bugs.
		_, _ = parser.ParseReader(strings.NewReader(data))
	})
}
