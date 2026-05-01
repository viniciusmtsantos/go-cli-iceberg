// Package vetdemo contains intentional bugs that go vet detects automatically.
// It exists for the "go vet" section of the presentation.
//
// Demonstrate with:
//
//	go vet ./...
//
// After the demo, fix the bugs (see the FIX: comments) and run again.
package vetdemo

import (
	"fmt"
	"sync"
)

// badCounter demonstrates the classic error of copying a sync.Mutex by value.
//
// go vet reports:
//
//	"passes lock by value: vetdemo.badCounter contains sync.Mutex"
//
// This happens because a value receiver (badCounter) makes a COPY of the struct,
// including the mutex — and each copy is a different, independent mutex,
// completely breaking synchronization.
//
// FIX: change the receiver from `c badCounter` to `c *badCounter`
type badCounter struct {
	mu    sync.Mutex
	count int
}

// BUG: value receiver copies the mutex. go vet detects this.
func (c badCounter) inc() { // ← should be (c *badCounter)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

// badFormat demonstrates a format string error that go vet detects.
//
// go vet reports:
//
//	"fmt.Sprintf format %s has arg n of wrong type int"
//
// FIX: change %s to %d
func badFormat(n int) string {
	return fmt.Sprintf("total de entradas: %s", n) // ← FIX: trocar %s por %d
}

// Avoid "declared and not used" errors for the types/functions above.
var (
	_ = badCounter{}
	_ = badFormat
)
