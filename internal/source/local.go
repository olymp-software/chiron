package source

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/olymp-software/chiron/internal/apperr"
)

// NormalizeLocal returns the course source of a local folder: its absolute
// path with symbolic links and . and .. resolved and no trailing separator.
// The folder must exist.
func NormalizeLocal(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", apperr.New(apperr.FolderNotFound, "path", path)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return "", apperr.New(apperr.FolderNotFound, "path", path)
	}
	if err != nil {
		return "", apperr.New(apperr.CourseReadFailed, "path", path)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", apperr.New(apperr.CourseReadFailed, "path", path)
	}
	if !info.IsDir() {
		return "", apperr.New(apperr.NotAFolder, "path", path)
	}
	return filepath.Clean(resolved), nil
}

// SameSource reports whether two course sources are the same course.
func SameSource(a, b string) bool {
	return ComparisonForm(a) == ComparisonForm(b)
}
