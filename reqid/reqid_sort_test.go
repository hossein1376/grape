package reqid

import (
	"testing"
	"time"
)

// TestNewRequestIDSortable generates a sequence of IDs, sleeping a few
// milliseconds between each generation to ensure the embedded millisecond
// timestamp advances. It then verifies the generated IDs are lexicographically
// increasing.
func TestNewRequestIDSortable(t *testing.T) {
	const n = 50
	ids := make([]string, n)
	for i := range n {
		ids[i] = string(NewRequestID())
		// Sleep a few milliseconds to make sure millisecond timestamp advances.
		time.Sleep(2 * time.Millisecond)
	}

	for i := 1; i < n; i++ {
		if ids[i-1] >= ids[i] {
			t.Fatalf(
				"IDs not lexicographically increasing at index %d: %q >= %q",
				i-1,
				ids[i-1],
				ids[i],
			)
		}
	}
}
