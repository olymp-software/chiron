package settings

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	dataDir := t.TempDir()
	return Open(dataDir, slog.New(slog.DiscardHandler)), filepath.Join(dataDir, fileName)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setLast(key string) func(st *Settings) {
	return func(st *Settings) { st.LastCourse = key }
}

func TestMissingSettings(t *testing.T) {
	s, path := newStore(t)

	if got := s.Get(); got.LastCourse != "" {
		t.Errorf("settings = %+v, want empty", got)
	}

	if err := s.Update(setLast("go-1")); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := s.Get(); got.LastCourse != "go-1" {
		t.Errorf("lastCourse = %q, want go-1", got.LastCourse)
	}
	want := "{\n  \"version\": 1,\n  \"lastCourse\": \"go-1\"\n}\n"
	if got := readFile(t, path); got != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

func TestCorruptSettings(t *testing.T) {
	s, path := newStore(t)
	writeFile(t, path, "{")

	if got := s.Get(); got.LastCourse != "" {
		t.Errorf("settings = %+v, want empty", got)
	}
	if err := s.Update(setLast("go-1")); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got := s.Get(); got.LastCourse != "go-1" {
		t.Errorf("lastCourse = %q, want go-1", got.LastCourse)
	}
}

func TestNewerSettingsFormatIsNotWritten(t *testing.T) {
	s, path := newStore(t)
	newer := `{"version": 2, "lastCourse": "go-1", "editor": "code"}`
	writeFile(t, path, newer)

	if got := s.Get(); got.LastCourse != "go-1" {
		t.Errorf("lastCourse = %q, want the stored one", got.LastCourse)
	}
	if err := s.Update(setLast("go-2")); err == nil {
		t.Error("Update of a newer format succeeded, want an error")
	}
	if got := readFile(t, path); got != newer {
		t.Errorf("file = %s, want unchanged", got)
	}
}
