// Package vetdemo contains intentional bugs that go vet detects automatically.
// It exists for the "go vet" section of the presentation (Bloco 9).
//
// This package intentionally causes go test ./... to fail with [build failed]
// when the vet bugs are active — that failure is the teaching moment for Bloco 3.
// After fixing both bugs (see FIX: comments in vet_demo.go), go test ./... passes.
//
// To see the vet issues:
//
//	go vet ./...
package vetdemo
