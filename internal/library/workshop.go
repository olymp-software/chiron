package library

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
)

// partSuffix marks a file being copied to the workshop folder. It is derived
// from the final path, so the next open of the activity finds what an
// interrupted copy left.
const partSuffix = ".chiron-part"

// WorkshopActivity is what an activity of a workshop shows beside its theory
// blocks.
type WorkshopActivity struct {
	Solution string            `json:"solution,omitempty"`
	HasCheck bool              `json:"hasCheck"`
	Controls []ActivityControl `json:"controls"`
}

// ActivityControl is a control of an activity; Group is its exclusion group,
// "" for the default one.
type ActivityControl struct {
	Label string `json:"label"`
	Group string `json:"group"`
}

// WorkshopFolder returns the workshop folder of the workshop with key.
func (l *Library) WorkshopFolder(key string) string {
	return filepath.Join(l.learnerRoot, key)
}

// EnsureWorkshopFolder returns the workshop folder of the workshop with key,
// created if the learner deleted it.
func (l *Library) EnsureWorkshopFolder(key string) (string, error) {
	e, err := l.entry(key)
	if err != nil {
		return "", err
	}
	folder := l.WorkshopFolder(e.Key)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		l.log.Error("create workshop folder", "key", e.Key, "err", err)
		return "", apperr.New(apperr.WorkshopFolderFailed)
	}
	return folder, nil
}

// fillWorkshopActivity reads the activity into a, after preparing the workshop
// folder and copying its files there.
func (l *Library) fillWorkshopActivity(key, root, section, slug string, a *Activity) error {
	folder, err := l.EnsureWorkshopFolder(key)
	if err != nil {
		return err
	}
	parsed, err := manifest.ParseWorkshopActivity(root, section, slug)
	if err != nil {
		return err
	}
	if err := l.copyFiles(filepath.Join(root, section, slug), folder, parsed.Files); err != nil {
		return err
	}
	a.TheoryBlocks = theoryBlocks(parsed.TheoryBlocks)
	w := &WorkshopActivity{
		Solution: parsed.Solution,
		HasCheck: parsed.Check != "",
		Controls: make([]ActivityControl, len(parsed.Controls)),
	}
	for i, c := range parsed.Controls {
		w.Controls[i] = ActivityControl{Label: c.Label, Group: c.Concurrency}
	}
	a.Workshop = w
	return nil
}

// copyFiles copies the files of items from the activity folder dir to folder,
// keeping their relative paths. Existing files are left as they are.
func (l *Library) copyFiles(dir, folder string, items []manifest.WorkspaceItem) error {
	for _, item := range items {
		src := filepath.Join(dir, filepath.FromSlash(item.Path))
		err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			rel, relErr := filepath.Rel(dir, path)
			if err == nil {
				err = relErr
			}
			if err == nil {
				err = copyEntry(path, filepath.Join(folder, rel), d)
			}
			if err != nil {
				l.log.Error("copy workshop file", "path", path, "err", err)
				return apperr.New(apperr.FileCopyFailed, "path", filepath.ToSlash(rel))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// copyEntry creates the folder or copies the file d at src to dst. Course
// copies hold no symbolic links, so nothing else is copied.
func copyEntry(src, dst string, d fs.DirEntry) error {
	switch {
	case d.IsDir():
		return os.MkdirAll(dst, 0o755)
	case d.Type().IsRegular():
		return copyMissing(src, dst)
	default:
		return nil
	}
}

// copyMissing copies the file src to dst unless something is there. The file
// is written next to dst and renamed, so dst only ever holds all of it.
func copyMissing(src, dst string) error {
	part := dst + partSuffix
	// What cannot be removed here fails the write below if it matters.
	_ = os.Remove(part)
	info, err := os.Lstat(dst)
	if err == nil {
		if info.IsDir() {
			return &fs.PathError{Op: "copy", Path: dst, Err: errors.New("a folder is in the way")}
		}
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := writePart(src, part); err != nil {
		_ = os.Remove(part)
		return err
	}
	// rename rides out the moment a scanner holds the new file on Windows.
	if err := rename(part, dst); err != nil {
		_ = os.Remove(part)
		return err
	}
	return nil
}

// writePart copies src to part byte for byte. The learner must be able to
// edit the copy, so of the source permissions only the executable bits stay.
func writePart(src, part string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(part, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644|info.Mode().Perm()&0o111)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
