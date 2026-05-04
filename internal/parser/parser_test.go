package parser_test

import (
	"strings"
	"testing"
	"time"

	"github.com/gopher/logscope/internal/parser"
	"github.com/gopher/logscope/pkg/entry"
)

// ─── fixtures ─────────────────────────────────────────────────────────────────

const validLine = "2026-04-26T10:00:00Z INFO GET /api/users 200 145 1234"

var multiLineInput = strings.TrimSpace(`
2026-04-26T10:00:00Z INFO GET /api/users 200 145 1234
2026-04-26T10:00:01Z WARN POST /api/auth/login 401 22 0
this is a malformed line that should be skipped
2026-04-26T10:00:02Z ERROR DELETE /api/orders/99 500 310 0

2026-04-26T10:00:03Z INFO GET /health 200 3 42
`)

// ─── unit tests ───────────────────────────────────────────────────────────────
// Run with: go test -v -run TestParse ./internal/parser/

func TestParse_ValidLine(t *testing.T) {
	e, err := parser.Parse(validLine)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.Method != "GET" {
		t.Errorf("method: got %q, want %q", e.Method, "GET")
	}
	if e.Path != "/api/users" {
		t.Errorf("path: got %q, want %q", e.Path, "/api/users")
	}
	if e.Status != 200 {
		t.Errorf("status: got %d, want 200", e.Status)
	}
	if e.Latency != 145*time.Millisecond {
		t.Errorf("latency: got %v, want 145ms", e.Latency)
	}
	if e.Bytes != 1234 {
		t.Errorf("bytes: got %d, want 1234", e.Bytes)
	}
	if e.Level != entry.INFO {
		t.Errorf("level: got %v, want INFO", e.Level)
	}
}

func TestParse_EmptyLine(t *testing.T) {
	_, err := parser.Parse("")
	if err == nil {
		t.Error("expected error for empty line, got nil")
	}
}

func TestParse_MalformedLine(t *testing.T) {
	cases := []struct {
		name string
		line string
	}{
		{"too few fields", "2026-04-26T10:00:00Z INFO GET /api/users 200"},
		{"invalid timestamp", "not-a-time INFO GET /api/users 200 10 100"},
		{"invalid level", "2026-04-26T10:00:00Z UNKNOWN GET /api/users 200 10 100"},
		{"invalid status", "2026-04-26T10:00:00Z INFO GET /api/users abc 10 100"},
		{"status out of range", "2026-04-26T10:00:00Z INFO GET /api/users 999 10 100"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parser.Parse(tc.line)
			if err == nil {
				t.Errorf("expected error for %q, got nil", tc.line)
			}
		})
	}
}

func TestParse_AllLevels(t *testing.T) {
	levels := []struct {
		str string
		lvl entry.Level
	}{
		{"DEBUG", entry.DEBUG},
		{"INFO", entry.INFO},
		{"WARN", entry.WARN},
		{"ERROR", entry.ERROR},
		{"FATAL", entry.FATAL},
	}

	for _, tc := range levels {
		t.Run(tc.str, func(t *testing.T) {
			line := "2026-04-26T10:00:00Z " + tc.str + " GET /api/test 200 10 0"
			e, err := parser.Parse(line)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if e.Level != tc.lvl {
				t.Errorf("got level %v, want %v", e.Level, tc.lvl)
			}
		})
	}
}

func TestParseReader_SkipsMalformed(t *testing.T) {
	entries, err := parser.ParseReader(strings.NewReader(multiLineInput))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4 valid lines (1 malformed + 1 empty line are skipped)
	if len(entries) != 4 {
		t.Errorf("got %d entries, want 4", len(entries))
	}
}

func TestParseReader_Empty(t *testing.T) {
	entries, err := parser.ParseReader(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

// TestParseReader_Parallel demonstrates the -parallel flag.
// Run with: go test -parallel=8 -run TestParseReader_Parallel ./internal/parser/
func TestParseReader_Parallel(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			entries, err := parser.ParseReader(strings.NewReader(multiLineInput))
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if len(entries) != 4 {
				t.Errorf("got %d entries, want 4", len(entries))
			}
		})
	}
}

// ─── benchmarks ───────────────────────────────────────────────────────────────
// Run with: go test -bench=. -benchmem ./internal/parser/
// Run with: go test -bench=BenchmarkParse -benchmem -count=5 ./internal/parser/

// BenchmarkParse measures single-line parsing throughput.
func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(validLine)
	}
}

// BenchmarkParseReader measures bulk parsing throughput.
// This is the hot path in logscope — worth profiling with pprof.
func BenchmarkParseReader(b *testing.B) {
	// Build a 1000-line input once
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(validLine)
		sb.WriteByte('\n')
	}
	input := sb.String()

	b.ResetTimer()
	b.SetBytes(int64(len(input)))

	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseReader(strings.NewReader(input))
	}
}

// BenchmarkParse_Parallel shows per-core parsing throughput.
// Run with: go test -bench=BenchmarkParse_Parallel -benchmem -cpu=1,2,4,8 ./internal/parser/
func BenchmarkParse_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = parser.Parse(validLine)
		}
	})
}
