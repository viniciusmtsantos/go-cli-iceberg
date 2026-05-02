// logscope analyzes structured HTTP access logs and reports latency percentiles,
// status code distributions, and top endpoints.
//
// Usage:
//
//	# Generate sample data
//	logscope -gen -lines 100000 -output access.log
//
//	# Analyze a log file
//	logscope -input access.log
//
//	# Pipeline mode
//	logscope -gen -lines 50000 | logscope
//
//	# With profiling (for go tool pprof demo)
//	logscope -input access.log -cpuprofile cpu.prof
//	go tool pprof -http=:8080 cpu.prof
//
// Build with version info:
//
//	go build -ldflags "-X main.version=1.2.0 -X main.commit=abc1234" ./cmd/logscope
//
// Build with JSON output:
//
//	go build -tags json -o logscope-json ./cmd/logscope
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/gopher/logscope/internal/generator"
	"github.com/gopher/logscope/internal/parser"
	"github.com/gopher/logscope/internal/processor"
	"github.com/gopher/logscope/internal/report"
)

// Build-time variables — injected via -ldflags.
//
// Example:
//
//	go build -ldflags "-X main.version=1.2.0 -X main.commit=$(git rev-parse --short HEAD)" ./cmd/logscope
//
// Inspect after build:
//
//	go version -m ./logscope
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	// ── flags ──────────────────────────────────────────────────────────────────
	var (
		flagVersion = flag.Bool("version", false, "print version info and exit")
		flagGen     = flag.Bool("gen", false, "generate a sample log file (use -lines to set size)")
		flagLines   = flag.Int("lines", 50_000, "number of log lines to generate (-gen mode)")
		flagInput   = flag.String("input", "", "log file to analyze (default: stdin)")
		flagOutput  = flag.String("output", "-", "output file for report or generated logs (default: stdout)")
		flagWorkers = flag.Int("workers", runtime.NumCPU(), "number of concurrent workers for processing")
		flagNaive   = flag.Bool("naive", false, "⚠️  use naive (racy) processor — for demo only, use with -race")

		// Observability flags — enables go tool pprof and go tool trace demos
		flagCPUProf = flag.String("cpuprofile", "", "write CPU profile to file (then: go tool pprof -http=:8080 <file>)")
		flagMemProf = flag.String("memprofile", "", "write heap profile to file (then: go tool pprof -http=:8080 <file>)")
		flagTrace   = flag.String("trace", "", "write execution trace to file (then: go tool trace <file>)")
	)
	flag.Parse()

	// ── version ────────────────────────────────────────────────────────────────
	if *flagVersion {
		fmt.Printf("logscope %s\n", version)
		fmt.Printf("  commit:     %s\n", commit)
		fmt.Printf("  built:      %s\n", buildTime)
		fmt.Printf("  go version: %s\n", runtime.Version())
		fmt.Printf("  platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("  workers:    %d (= NumCPU)\n", runtime.NumCPU())
		return
	}

	// ── output target ──────────────────────────────────────────────────────────
	out := os.Stdout
	if *flagOutput != "-" {
		f, err := os.Create(*flagOutput)
		must("create output file", err)
		defer f.Close()
		out = f
	}

	// ── profiling / tracing setup ──────────────────────────────────────────────
	if *flagCPUProf != "" {
		f, err := os.Create(*flagCPUProf)
		must("create cpu profile", err)
		defer f.Close()
		must("start cpu profile", pprof.StartCPUProfile(f))
		defer pprof.StopCPUProfile()
		fmt.Fprintf(os.Stderr, "📊 CPU profiling → %s\n", *flagCPUProf)
	}

	if *flagTrace != "" {
		f, err := os.Create(*flagTrace)
		must("create trace file", err)
		defer f.Close()
		must("start trace", trace.Start(f))
		defer trace.Stop()
		fmt.Fprintf(os.Stderr, "🔬 Execution tracing → %s\n", *flagTrace)
	}

	// ── generate mode ──────────────────────────────────────────────────────────
	if *flagGen {
		generator.Generate(out, *flagLines)
		fmt.Fprintf(os.Stderr, "✓ generated %d log lines\n", *flagLines)
		return
	}

	// ── analyze mode ───────────────────────────────────────────────────────────
	in := os.Stdin
	if *flagInput != "" {
		f, err := os.Open(*flagInput)
		must("open input file", err)
		defer f.Close()
		in = f
	}

	fmt.Fprintln(os.Stderr, "⏳ parsing...")
	entries, err := parser.ParseReader(in)
	must("parse logs", err)
	fmt.Fprintf(os.Stderr, "✓ parsed %d entries\n", len(entries))

	fmt.Fprintf(os.Stderr, "⏳ processing with %d worker(s)...\n", *flagWorkers)
	var stats *processor.Stats
	if *flagNaive {
		fmt.Fprintln(os.Stderr, "⚠️  using naive (racy) processor — run with -race to detect data races")
		stats = processor.ProcessNaive(entries)
	} else {
		stats = processor.ProcessSafe(entries, *flagWorkers)
	}
	fmt.Fprintln(os.Stderr, "✓ done")

	report.Write(out, stats)

	// Memory profile is written after all work completes (captures peak usage)
	if *flagMemProf != "" {
		f, err := os.Create(*flagMemProf)
		must("create mem profile", err)
		defer f.Close()
		runtime.GC() // force GC before heap snapshot
		must("write mem profile", pprof.WriteHeapProfile(f))
		fmt.Fprintf(os.Stderr, "📊 heap profile → %s\n", *flagMemProf)
	}
}

func must(op string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error %s: %v\n", op, err)
		os.Exit(1)
	}
}
