package library

import "time"

// Antivirus and indexer handles make a folder rename fail for a moment on
// Windows, so it is retried.
const (
	renameAttempts = 10
	renameDelay    = 100 * time.Millisecond
)
