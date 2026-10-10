package testutil

import (
	"cmp"
	"os"

	"openmars/internal/logger"
)

// Tests log errors only, so failures are not buried. Set OPENMARS_TEST_LOG_LEVEL to
// debug, info or warn to see more while investigating one.
func init() {
	logger.Init(os.Stderr, cmp.Or(os.Getenv("OPENMARS_TEST_LOG_LEVEL"), "error"))
}
