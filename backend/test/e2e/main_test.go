package e2e_test

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain fails the run if any server goroutine outlives the tests: every server is
// closed by its test's cleanup, so nothing it started may still be running.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
