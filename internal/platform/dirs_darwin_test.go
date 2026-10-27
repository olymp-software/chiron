package platform

import (
	"path/filepath"
	"testing"
)

func TestDirsDarwin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	data, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if want := filepath.Join(home, "Library", "Application Support", "Chiron"); data != want {
		t.Errorf("DataDir = %q, want %q", data, want)
	}

	logDir, err := LogDir()
	if err != nil {
		t.Fatalf("LogDir: %v", err)
	}
	if want := filepath.Join(home, "Library", "Logs", "Chiron"); logDir != want {
		t.Errorf("LogDir = %q, want %q", logDir, want)
	}
}
