// Package settings keeps app-level state and learner settings in
// settings.json. Without the file the app starts with defaults, so its errors
// are logged and never shown to the learner.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/olymp-software/chiron/internal/platform"
)

const (
	fileName = "settings.json"
	// formatVersion is the settings.json format this app reads and writes.
	formatVersion = 1
)

// Settings is the content of settings.json.
type Settings struct {
	Version int `json:"version"`
	// LastCourse is the storage key of the course entered last.
	LastCourse string `json:"lastCourse,omitempty"`
}

// Store is settings.json in a data folder. Its methods are safe for concurrent
// use.
type Store struct {
	path string
	log  *slog.Logger
	// mu serializes reads and writes of settings.json.
	mu sync.Mutex
}

// Open returns the settings store in dataDir.
func Open(dataDir string, log *slog.Logger) *Store {
	return &Store{path: filepath.Join(dataDir, fileName), log: log}
}

// Get returns the settings. An unreadable file reads as empty settings.
func (s *Store) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

// Update applies change to the settings and saves them. Settings in a newer
// format are not written.
func (s *Store) Update(change func(st *Settings)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.load()
	if st.Version > formatVersion {
		return fmt.Errorf("settings in newer format %d not written", st.Version)
	}
	change(&st)
	st.Version = formatVersion
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return platform.WriteFileAtomic(s.path, append(data, '\n'), 0o644)
}

func (s *Store) load() Settings {
	empty := Settings{Version: formatVersion}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return empty
	}
	if err != nil {
		s.log.Error("read settings", "err", err)
		return empty
	}
	var st Settings
	if err := json.Unmarshal(data, &st); err != nil {
		s.log.Error("parse settings", "err", err)
		return empty
	}
	return st
}
