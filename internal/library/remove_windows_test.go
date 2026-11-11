package library

import (
	"os"
	"path/filepath"
	"testing"
)

// On Windows an open file keeps its folder from being deleted; the library
// item is removed anyway and the folder is reported.
func TestRemoveWithFolderHeldOpen(t *testing.T) {
	l, dataDir, key := removedWorkshop(t)
	f, err := os.Open(filepath.Join(l.WorkshopFolder(key), "mine"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	left, err := l.Remove(key, false, false)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if len(left) != 1 || left[0] != l.WorkshopFolder(key) {
		t.Errorf("left = %v, want the workshop folder", left)
	}
	wantNotInLibrary(t, l, key)
	if exists(l.CopyPath(key)) || exists(statePath(dataDir, key)) {
		t.Error("course copy or progress survived")
	}
}
