// Package generator creates realistic HTTP access log data for testing and demos.
package generator

import (
	"fmt"
	"io"
	"math/rand"
	"time"
)

var (
	methods = []string{"GET", "GET", "GET", "POST", "PUT", "DELETE", "PATCH"}

	paths = []string{
		"/api/users",
		"/api/users/{id}",
		"/api/products",
		"/api/products/{id}",
		"/api/orders",
		"/api/orders/{id}",
		"/api/auth/login",
		"/api/auth/logout",
		"/api/auth/refresh",
		"/api/payments",
		"/health",
		"/metrics",
		"/api/v2/events",
		"/api/v2/reports",
	}

	// Weighted toward 200 to simulate realistic traffic
	statuses = []int{
		200, 200, 200, 200, 200,
		201, 201,
		204,
		400, 401, 403, 404, 404,
		500, 502, 503,
	}

	// Latency distribution in ms: most fast, some slow, rare outliers
	latencyBands = []struct {
		min, max int
		weight   int
	}{
		{1, 50, 60},    // fast: 60% of requests
		{51, 200, 25},  // medium: 25%
		{201, 500, 10}, // slow: 10%
		{501, 2000, 5}, // very slow: 5%
	}
)

// Generate writes n realistic HTTP access log lines to w.
//
// Each line follows the format:
//
//	<timestamp> <level> <method> <path> <status> <latency_ms> <bytes>
//
// Example usage:
//
//	logscope -gen -lines 100000 -output access.log
func Generate(w io.Writer, n int) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	base := time.Now().Add(-time.Duration(n) * time.Second)

	for i := range n {
		ts := base.Add(time.Duration(i) * time.Second)
		status := statuses[rng.Intn(len(statuses))]
		method := methods[rng.Intn(len(methods))]
		path := paths[rng.Intn(len(paths))]
		latency := sampleLatency(rng)
		bytes := rng.Intn(50000) + 64

		level := levelFor(status)

		fmt.Fprintf(w, "%s %s %s %s %d %d %d\n",
			ts.UTC().Format(time.RFC3339),
			level,
			method,
			path,
			status,
			latency,
			bytes,
		)
	}
}

func levelFor(status int) string {
	switch {
	case status >= 500:
		return "ERROR"
	case status >= 400:
		return "WARN"
	default:
		return "INFO"
	}
}

func sampleLatency(rng *rand.Rand) int {
	// Pick a band by weight
	total := 0
	for _, b := range latencyBands {
		total += b.weight
	}
	r := rng.Intn(total)
	for _, b := range latencyBands {
		r -= b.weight
		if r < 0 {
			return b.min + rng.Intn(b.max-b.min+1)
		}
	}
	return 100
}
