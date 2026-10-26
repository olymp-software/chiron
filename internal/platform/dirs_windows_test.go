package platform

import (
	"path/filepath"
	"testing"
)

func TestDirsWindows(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)

	data, err := DataDir()
	if err != nil {
		t.Fatalf("DataDir: %v", err)
	}
	if want := filepath.Join(local, "Chiron"); data != want {
		t.Errorf("DataDir = %q, want %q", data, want)
	}

	logDir, err := LogDir()
	if err != nil {
		t.Fatalf("LogDir: %v", err)
	}
	if want := filepath.Join(local, "Chiron", "logs"); logDir != want {
		t.Errorf("LogDir = %q, want %q", logDir, want)
	}
}

func TestDirsWindowsUnset(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")

	if _, err := DataDir(); err == nil {
		t.Error("DataDir: want error when LOCALAPPDATA is unset")
	}
	if _, err := LogDir(); err == nil {
		t.Error("LogDir: want error when LOCALAPPDATA is unset")
	}
}
