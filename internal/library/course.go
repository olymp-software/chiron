package library

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/source"
)

// Course copies are replaced as a whole: a new copy is built next to the
// current one and swapped in by renames, so an interruption leaves either the
// old or the new copy in full (see recover).
const (
	newSuffix = ".new-"
	oldSuffix = ".old-"
)

// AddLocal adds the local course folder at path. A folder already in the
// library returns its entry. The folder must have a valid course.yaml.
func (l *Library) AddLocal(path string) (Entry, error) {
	src, err := source.NormalizeLocal(path)
	if err != nil {
		return Entry{}, err
	}
	if e, ok, err := l.findSame(src); err != nil || ok {
		return e, err
	}
	key := source.Key(src)
	unlock := l.lockCourse(key)
	defer unlock()

	newDir, err := l.buildCopy(src, key)
	if err != nil {
		return Entry{}, err
	}
	c, err := manifest.ParseCourse(newDir)
	if err != nil {
		l.removeAll(newDir)
		return Entry{}, err
	}

	e := Entry{Origin: OriginLocal, Source: src, Key: key, Title: c.Title}
	err = l.update(func(f *libraryFile) error {
		for _, other := range f.Courses {
			if other.Key == key {
				return apperr.New(apperr.StorageKeyTaken, "key", key, "source", other.Source)
			}
		}
		// A copy left by a course added before is replaced.
		l.removeAll(l.CopyPath(key))
		if err := rename(newDir, l.CopyPath(key)); err != nil {
			l.log.Error("move new course copy", "key", key, "err", err)
			return apperr.New(apperr.CopyFailed, "path", src)
		}
		f.Courses = append(f.Courses, e)
		return nil
	})
	if err != nil {
		l.removeAll(newDir)
		return Entry{}, err
	}
	l.log.Info("course added", "source", src, "key", key)
	return e, nil
}

// findSame returns the entry whose source is the same course as src.
func (l *Library) findSame(src string) (Entry, bool, error) {
	courses, err := l.List()
	if err != nil {
		return Entry{}, false, err
	}
	for _, e := range courses {
		if source.SameSource(e.Source, src) {
			return e, true, nil
		}
	}
	return Entry{}, false, nil
}

// Sync replaces the course copy of a local-folder course with a fresh copy of
// its folder. The new copy is not validated; the stored title is updated only
// if its course.yaml is valid. If the folder cannot be copied, the copy is
// left unchanged.
func (l *Library) Sync(key string) (Entry, error) {
	e, err := l.entry(key)
	if err != nil {
		return Entry{}, err
	}
	unlock := l.lockCourse(e.Key)
	defer unlock()

	// A running script holds its working directory, which blocks the swap on
	// Windows.
	l.StopCheck(e.Key)
	newDir, err := l.buildCopy(e.Source, e.Key)
	if err != nil {
		return Entry{}, err
	}
	if err := l.swap(newDir, e.Key); err != nil {
		l.removeAll(newDir)
		return Entry{}, err
	}
	l.log.Info("course synced", "source", e.Source, "key", e.Key)

	c, err := manifest.ParseCourse(l.CopyPath(e.Key))
	if err != nil || c.Title == e.Title {
		return e, nil
	}
	err = l.update(func(f *libraryFile) error {
		for i := range f.Courses {
			if f.Courses[i].Key == e.Key {
				f.Courses[i].Title = c.Title
				e = f.Courses[i]
			}
		}
		return nil
	})
	return e, err
}

// buildCopy copies the folder src into a new temporary folder next to the
// course copy with key and returns it.
func (l *Library) buildCopy(src, key string) (string, error) {
	if _, err := source.NormalizeLocal(src); err != nil {
		return "", err
	}
	newDir := filepath.Join(l.coursesPath(), key+newSuffix+strconv.FormatInt(time.Now().UnixNano(), 10))
	skipped, err := source.CopyLocal(src, newDir)
	if err != nil {
		l.log.Error("copy course folder", "source", src, "err", err)
		l.removeAll(newDir)
		return "", apperr.New(apperr.CopyFailed, "path", src)
	}
	for _, s := range skipped {
		l.log.Info("symbolic link not copied", "source", src, "path", s)
	}
	return newDir, nil
}

// swap makes newDir the course copy with key: the current copy is moved
// aside, newDir is moved in and the old copy is deleted.
func (l *Library) swap(newDir, key string) error {
	cur := l.CopyPath(key)
	old := filepath.Join(l.coursesPath(), key+oldSuffix+strconv.FormatInt(time.Now().UnixNano(), 10))
	if err := rename(cur, old); err != nil && !errors.Is(err, os.ErrNotExist) {
		l.log.Error("move course copy aside", "key", key, "err", err)
		return apperr.New(apperr.CopyFailed, "path", cur)
	}
	if err := rename(newDir, cur); err != nil {
		l.log.Error("move new course copy", "key", key, "err", err)
		if err := rename(old, cur); err != nil && !errors.Is(err, os.ErrNotExist) {
			l.log.Error("restore course copy", "key", key, "err", err)
		}
		return apperr.New(apperr.CopyFailed, "path", cur)
	}
	l.removeAll(old)
	return nil
}

// recover repairs course copies after an interrupted swap: a course whose
// copy is missing gets back its newest moved-aside copy, and every leftover
// new or old copy is deleted. It also deletes the skeleton's smoke folder.
// With an unreadable library.json nothing is repaired; List reports the error.
func (l *Library) recover() error {
	l.removeAll(filepath.Join(l.dataDir, "smoke"))
	courses, err := l.List()
	if err != nil {
		l.log.Error("course copies not repaired", "err", err)
		return nil
	}
	entries, err := os.ReadDir(l.coursesPath())
	if err != nil {
		l.log.Error("read courses folder", "err", err)
		return apperr.New(apperr.LibraryReadFailed)
	}
	for _, c := range courses {
		if _, err := os.Stat(l.CopyPath(c.Key)); err == nil {
			continue
		}
		if old := newestOld(entries, c.Key); old != "" {
			if err := rename(filepath.Join(l.coursesPath(), old), l.CopyPath(c.Key)); err != nil {
				l.log.Error("restore course copy", "key", c.Key, "err", err)
			} else {
				l.log.Info("course copy restored", "key", c.Key, "from", old)
			}
		}
	}
	entries, err = os.ReadDir(l.coursesPath())
	if err != nil {
		l.log.Error("read courses folder", "err", err)
		return apperr.New(apperr.LibraryReadFailed)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), newSuffix) || strings.Contains(e.Name(), oldSuffix) {
			l.removeAll(filepath.Join(l.coursesPath(), e.Name()))
		}
	}
	return nil
}

// newestOld returns the name of the newest moved-aside copy of key, or "".
// Names end in a nanosecond timestamp.
func newestOld(entries []os.DirEntry, key string) string {
	var best string
	var bestTime int64
	for _, e := range entries {
		stamp, ok := strings.CutPrefix(e.Name(), key+oldSuffix)
		if !ok {
			continue
		}
		t, err := strconv.ParseInt(stamp, 10, 64)
		if err == nil && t > bestTime {
			best, bestTime = e.Name(), t
		}
	}
	return best
}

func (l *Library) removeAll(path string) {
	if err := os.RemoveAll(path); err != nil {
		l.log.Error("remove folder", "path", path, "err", err)
	}
}

// rename renames a folder, retrying while the OS reports it busy.
func rename(from, to string) error {
	var err error
	for i := range renameAttempts {
		if err = os.Rename(from, to); err == nil || errors.Is(err, os.ErrNotExist) {
			return err
		}
		if i < renameAttempts-1 {
			time.Sleep(renameDelay)
		}
	}
	return err
}
