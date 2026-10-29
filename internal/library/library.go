// Package library keeps the list of added courses in library.json and owns
// the course copies under <data>/courses.
package library

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/platform"
	"github.com/olymp-software/chiron/internal/progress"
)

const (
	fileName = "library.json"
	// formatVersion is the library.json format this app reads and writes.
	formatVersion = 1
	coursesDir    = "courses"
	stateDir      = "state"
)

// Origin tells where a course comes from.
type Origin string

// Origins of a course.
const (
	OriginLocal Origin = "local"
	OriginRepo  Origin = "repo"
)

// Entry is one added course. CheckHash is the SHA-256 (lowercase hex) of the
// check file the environment check last succeeded for, "" if none.
type Entry struct {
	Origin    Origin `json:"origin"`
	Source    string `json:"source"`
	Key       string `json:"key"`
	Title     string `json:"title"`
	CheckHash string `json:"checkHash,omitempty"`
}

// Course is a course in the library list. Enrolled is not stored in
// library.json: a course is enrolled exactly when it has progress.
type Course struct {
	Entry
	Enrolled bool `json:"enrolled"`
}

type libraryFile struct {
	Version int     `json:"version"`
	Courses []Entry `json:"courses"`
}

// Library is the learner's library in a data folder. Its methods are safe for
// concurrent use; errors are *apperr.Error.
type Library struct {
	dataDir  string
	log      *slog.Logger
	progress *progress.Store

	// mu serializes reads and writes of library.json.
	mu sync.Mutex
	// courseLocks serializes changes of one course copy, keyed by storage key.
	courseLocks sync.Map

	// checksMu guards checks, the running environment checks by storage key.
	checksMu sync.Mutex
	checks   map[string]*check
}

// Open opens the library in dataDir and repairs course copies left by an
// interrupted add or Sync.
func Open(dataDir string, log *slog.Logger) (*Library, error) {
	l := &Library{dataDir: dataDir, log: log, checks: map[string]*check{}}
	if err := os.MkdirAll(l.coursesPath(), 0o755); err != nil {
		log.Error("create courses folder", "err", err)
		return nil, apperr.New(apperr.LibraryWriteFailed)
	}
	p, err := progress.Open(filepath.Join(dataDir, stateDir), log)
	if err != nil {
		return nil, err
	}
	l.progress = p
	if err := l.recover(); err != nil {
		return nil, err
	}
	return l, nil
}

// Courses returns the added courses with their stored titles and enrollment.
func (l *Library) Courses() ([]Course, error) {
	entries, err := l.List()
	if err != nil {
		return nil, err
	}
	courses := make([]Course, len(entries))
	for i, e := range entries {
		courses[i] = Course{Entry: e, Enrolled: l.enrolled(e.Key)}
	}
	return courses, nil
}

// enrolled reports whether the course with key has progress. A progress file
// that cannot be checked is logged and counts as none.
func (l *Library) enrolled(key string) bool {
	ok, _ := l.progress.Exists(key)
	return ok
}

// List returns the added courses with their stored titles.
func (l *Library) List() ([]Entry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := l.load()
	if err != nil {
		return nil, err
	}
	if f.Courses == nil {
		return []Entry{}, nil
	}
	return f.Courses, nil
}

// entry returns the library entry with the storage key.
func (l *Library) entry(key string) (Entry, error) {
	courses, err := l.List()
	if err != nil {
		return Entry{}, err
	}
	for _, e := range courses {
		if e.Key == key {
			return e, nil
		}
	}
	return Entry{}, apperr.New(apperr.CourseNotFound, "key", key)
}

// update applies change to library.json under the lock.
func (l *Library) update(change func(f *libraryFile) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := l.load()
	if err != nil {
		return err
	}
	if err := change(&f); err != nil {
		return err
	}
	return l.save(f)
}

func (l *Library) load() (libraryFile, error) {
	data, err := os.ReadFile(l.filePath())
	if errors.Is(err, fs.ErrNotExist) {
		return libraryFile{Version: formatVersion}, nil
	}
	if err != nil {
		l.log.Error("read library", "err", err)
		return libraryFile{}, apperr.New(apperr.LibraryReadFailed)
	}
	var f libraryFile
	if err := json.Unmarshal(data, &f); err != nil {
		l.log.Error("parse library", "err", err)
		return libraryFile{}, apperr.New(apperr.LibraryReadFailed)
	}
	if f.Version > formatVersion {
		return libraryFile{}, apperr.New(apperr.LibraryNewerFormat)
	}
	return f, nil
}

func (l *Library) save(f libraryFile) error {
	f.Version = formatVersion
	if f.Courses == nil {
		f.Courses = []Entry{}
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(l.filePath(), append(data, '\n'), 0o644); err != nil {
		l.log.Error("write library", "err", err)
		return apperr.New(apperr.LibraryWriteFailed)
	}
	return nil
}

// lockCourse locks changes of the course copy with key and returns the unlock
// function.
func (l *Library) lockCourse(key string) func() {
	m, _ := l.courseLocks.LoadOrStore(key, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

func (l *Library) filePath() string    { return filepath.Join(l.dataDir, fileName) }
func (l *Library) coursesPath() string { return filepath.Join(l.dataDir, coursesDir) }

// CopyPath returns the folder of the course copy with key.
func (l *Library) CopyPath(key string) string { return filepath.Join(l.coursesPath(), key) }
