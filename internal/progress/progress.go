// Package progress keeps the learner's progress of each enrolled course in
// <state>/<key>.json. A course is enrolled exactly when its progress file
// exists.
package progress

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/platform"
)

// formatVersion is the progress file format this app reads and writes.
const formatVersion = 1

// Progress is the progress of one course. Current and Completed hold activity
// slugs. CourseCompleted is stored, so the library can show it without reading
// the course; Resolve keeps it in line with the course order.
type Progress struct {
	Version         int      `json:"version"`
	Current         string   `json:"current"`
	Completed       []string `json:"completed"`
	CourseCompleted bool     `json:"courseCompleted"`
}

// Store is the progress of all courses in one folder. Its methods are safe for
// concurrent use; errors are *apperr.Error.
type Store struct {
	dir string
	log *slog.Logger
	// locks serializes changes of one progress file, keyed by storage key.
	locks sync.Map
}

// Open opens the progress store in dir, creating the folder.
func Open(dir string, log *slog.Logger) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Error("create state folder", "err", err)
		return nil, apperr.New(apperr.ProgressWriteFailed)
	}
	return &Store{dir: dir, log: log}, nil
}

// Exists reports whether the course with key is enrolled.
func (s *Store) Exists(key string) (bool, error) {
	_, err := os.Stat(s.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		s.log.Error("stat progress", "key", key, "err", err)
		return false, apperr.New(apperr.ProgressReadFailed)
	}
	return true, nil
}

// Load returns the progress of the course with key; a course without progress
// is not enrolled. A file in a newer format is read as far as this app
// understands it.
func (s *Store) Load(key string) (Progress, error) {
	data, err := os.ReadFile(s.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return Progress{}, apperr.New(apperr.NotEnrolled)
	}
	if err != nil {
		s.log.Error("read progress", "key", key, "err", err)
		return Progress{}, apperr.New(apperr.ProgressReadFailed)
	}
	var p Progress
	if err := json.Unmarshal(data, &p); err != nil {
		s.log.Error("parse progress", "key", key, "err", err)
		return Progress{}, apperr.New(apperr.ProgressReadFailed)
	}
	if p.Completed == nil {
		p.Completed = []string{}
	}
	return p, nil
}

// Enroll creates the progress of the course with key with first as the current
// activity. An enrolled course keeps its progress.
func (s *Store) Enroll(key, first string) error {
	unlock := s.lock(key)
	defer unlock()
	if ok, err := s.Exists(key); err != nil || ok {
		return err
	}
	return s.save(key, Progress{Version: formatVersion, Current: first, Completed: []string{}})
}

// Update applies change to the progress of the course with key under its lock
// and saves the result if it differs. It returns the resulting progress.
func (s *Store) Update(key string, change func(p Progress) Progress) (Progress, error) {
	unlock := s.lock(key)
	defer unlock()
	p, err := s.Load(key)
	if err != nil {
		return Progress{}, err
	}
	next := change(p)
	if next.Current == p.Current && slices.Equal(next.Completed, p.Completed) &&
		next.CourseCompleted == p.CourseCompleted {
		return p, nil
	}
	if p.Version > formatVersion {
		// Writing would drop what the newer app stored.
		s.log.Error("progress in a newer format not written", "key", key, "version", p.Version)
		return Progress{}, apperr.New(apperr.ProgressWriteFailed)
	}
	if err := s.save(key, next); err != nil {
		return Progress{}, err
	}
	return next, nil
}

// Delete deletes the progress of the course with key, ending its enrollment.
// A course without progress is not an error.
func (s *Store) Delete(key string) error {
	unlock := s.lock(key)
	defer unlock()
	if err := os.Remove(s.path(key)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		s.log.Error("delete progress", "key", key, "err", err)
		return apperr.New(apperr.ProgressWriteFailed)
	}
	return nil
}

// Path returns the progress file of the course with key.
func (s *Store) Path(key string) string { return s.path(key) }

func (s *Store) save(key string, p Progress) error {
	p.Version = formatVersion
	if p.Completed == nil {
		p.Completed = []string{}
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(s.path(key), append(data, '\n'), 0o644); err != nil {
		s.log.Error("write progress", "key", key, "err", err)
		return apperr.New(apperr.ProgressWriteFailed)
	}
	return nil
}

// lock locks changes of the progress of key and returns the unlock function.
func (s *Store) lock(key string) func() {
	m, _ := s.locks.LoadOrStore(key, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (s *Store) path(key string) string { return filepath.Join(s.dir, key+".json") }

// Resolve returns p reconciled with order, the activity slugs of the course in
// course order. The current activity becomes the first activity of order that
// is not completed, or the last one when all are; this covers a removed
// current activity and activities added anywhere in the course.
// CourseCompleted becomes whether every activity of order is completed.
func Resolve(order []string, p Progress) Progress {
	if len(order) == 0 {
		return p
	}
	i := slices.IndexFunc(order, func(slug string) bool { return !slices.Contains(p.Completed, slug) })
	p.CourseCompleted = i < 0
	if i < 0 {
		i = len(order) - 1
	}
	p.Current = order[i]
	return p
}
