package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLine(t *testing.T, dir, msg string) {
	t.Helper()
	logger, f, err := OpenLog(dir)
	if err != nil {
		t.Fatalf("OpenLog: %v", err)
	}
	logger.Info(msg)
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(b)
}

func TestOpenLogCreates(t *testing.T) {
	dir := t.TempDir()
	writeLine(t, dir, "first")

	got := readFile(t, filepath.Join(dir, logName))
	if !strings.Contains(got, "level=INFO msg=first") {
		t.Errorf("log = %q, want a line with msg=first", got)
	}
}

func TestOpenLogAppends(t *testing.T) {
	dir := t.TempDir()
	writeLine(t, dir, "first")
	writeLine(t, dir, "second")

	got := readFile(t, filepath.Join(dir, logName))
	if !strings.Contains(got, "msg=first") || !strings.Contains(got, "msg=second") {
		t.Errorf("log = %q, want both lines", got)
	}
}

func TestOpenLogRollsOver(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		wantRoll bool
	}{
		{name: "at limit", size: maxLogSize, wantRoll: false},
		{name: "over limit", size: maxLogSize + 1, wantRoll: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, logName)
			if err := os.WriteFile(path+".1", []byte("stale"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Truncate(path, tt.size); err != nil {
				t.Fatal(err)
			}

			writeLine(t, dir, "after")

			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			old, err := os.Stat(path + ".1")
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantRoll {
				if info.Size() >= maxLogSize {
					t.Errorf("chiron.log size = %d, want a fresh file", info.Size())
				}
				if old.Size() != tt.size {
					t.Errorf("chiron.log.1 size = %d, want %d", old.Size(), tt.size)
				}
			} else {
				if info.Size() <= tt.size {
					t.Errorf("chiron.log size = %d, want appended to %d", info.Size(), tt.size)
				}
				if got := readFile(t, path+".1"); got != "stale" {
					t.Errorf("chiron.log.1 = %q, want untouched", got)
				}
			}
		})
	}
}
