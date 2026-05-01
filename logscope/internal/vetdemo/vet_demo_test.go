package vetdemo

import "testing"

// TestVetDemo is skipped because this package contains intentional vet bugs
// for demonstration purposes. Run go vet ./... to see the bugs.
func TestVetDemo(t *testing.T) {
t.Skip("This package contains intentional vet bugs. Use 'go vet ./...' to see them.")
}
