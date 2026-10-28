//go:build !windows

package library

import "time"

// A folder rename does not fail transiently on Linux and macOS.
const (
	renameAttempts = 1
	renameDelay    = time.Duration(0)
)
