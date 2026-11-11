package library

import (
	"path/filepath"
	"slices"

	"github.com/olymp-software/chiron/internal/apperr"
)

// Remove removes the library item with key: it stops its scripts, deletes its
// library entry with the remembered check result, then deletes its course
// copy, its TEMP_DIRs, its progress unless keepProgress and its learner
// folders or workshop folder unless keepFiles. Once the entry is gone the item
// is removed; the paths that could not be deleted are returned, not an error.
func (l *Library) Remove(key string, keepProgress, keepFiles bool) ([]string, error) {
	e, err := l.entry(key)
	if err != nil {
		return nil, err
	}
	// The course lock keeps a script start, a Sync or a re-add of the same
	// source from interleaving with the removal.
	unlock := l.lockCourse(e.Key)
	defer unlock()
	// Stopped scripts release their working directories and TEMP_DIRs.
	l.StopCheck(e.Key)
	l.runs.stopCourse(e.Key)
	err = l.update(func(f *libraryFile) error {
		i := slices.IndexFunc(f.Courses, func(c Entry) bool { return c.Key == e.Key })
		if i < 0 {
			return apperr.New(apperr.CourseNotFound, "key", e.Key)
		}
		f.Courses = slices.Delete(f.Courses, i, i+1)
		return nil
	})
	if err != nil {
		return nil, err
	}
	l.log.Info("course removed", "source", e.Source, "key", e.Key,
		"keepProgress", keepProgress, "keepFiles", keepFiles)

	left := []string{}
	remove := func(path string) {
		if err := removeRetrying(path); err != nil {
			l.log.Error("remove after course removal", "path", path, "err", err)
			left = append(left, path)
		}
	}
	remove(l.CopyPath(e.Key))
	remove(filepath.Join(l.dataDir, tempDir, e.Key))
	if !keepProgress {
		if err := l.progress.Delete(e.Key); err != nil {
			left = append(left, l.progress.Path(e.Key))
		}
	}
	if !keepFiles {
		remove(l.WorkshopFolder(e.Key))
	}
	return left, nil
}
