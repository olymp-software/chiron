package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func TestWriteFileAtomicCreates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "library.json")

	if err := WriteFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}

	if got := readFile(t, path); got != "new" {
		t.Errorf("content = %q, want %q", got, "new")
	}
	if names := dirNames(t, dir); len(names) != 1 {
		t.Errorf("folder = %v, want only library.json", names)
	}
}

func TestWriteFileAtomicReplaces(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "library.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := WriteFileAtomic(path, []byte("new"), 0o644); err != nil {
		t.Fatalf("WriteFileAtomic: %v", err)
	}

	if got := readFile(t, path); got != "new" {
		t.Errorf("content = %q, want %q", got, "new")
	}
	if names := dirNames(t, dir); len(names) != 1 {
		t.Errorf("folder = %v, want only library.json", names)
	}
}

func TestWriteFileAtomicFailureKeepsOld(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "library.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A folder in the way of the rename target makes the final rename fail.
	blocked := filepath.Join(dir, "blocked")
	if err := os.MkdirAll(filepath.Join(blocked, "child"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := WriteFileAtomic(blocked, []byte("new"), 0o644); err == nil {
		t.Fatal("WriteFileAtomic over a non-empty folder succeeded, want an error")
	}

	if got := readFile(t, path); got != "old" {
		t.Errorf("content = %q, want untouched", got)
	}
	if names := dirNames(t, dir); len(names) != 2 {
		t.Errorf("folder = %v, want no temp files left", names)
	}
}
