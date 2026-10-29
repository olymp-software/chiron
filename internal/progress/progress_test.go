package progress

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
)

func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "state")
	s, err := Open(dir, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return s, dir
}

func wantCode(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("err = %v, want apperr %s", err, code)
	}
	if ae.Code != code {
		t.Fatalf("code = %s, want %s (%v)", ae.Code, code, ae)
	}
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

func TestEnroll(t *testing.T) {
	s, dir := newStore(t)

	if err := s.Enroll("go-1", "hello"); err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	if ok, err := s.Exists("go-1"); err != nil || !ok {
		t.Errorf("Exists = %v, %v, want true", ok, err)
	}
	p, err := s.Load("go-1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := Progress{Version: 1, Current: "hello", Completed: []string{}}
	if !reflect.DeepEqual(p, want) {
		t.Errorf("progress = %+v, want %+v", p, want)
	}
	want1 := "{\n  \"version\": 1,\n  \"current\": \"hello\",\n  \"completed\": []\n}\n"
	if got := readFile(t, filepath.Join(dir, "go-1.json")); got != want1 {
		t.Errorf("file = %q, want %q", got, want1)
	}
}

func TestRepeatedEnrollKeepsProgress(t *testing.T) {
	s, dir := newStore(t)
	path := filepath.Join(dir, "go-1.json")
	stored := `{"version": 1, "current": "loops", "completed": ["hello"]}`
	writeFile(t, path, stored)

	if err := s.Enroll("go-1", "hello"); err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	if got := readFile(t, path); got != stored {
		t.Errorf("file = %s, want unchanged", got)
	}
}

func TestMissingProgress(t *testing.T) {
	s, _ := newStore(t)

	ok, err := s.Exists("go-1")
	if err != nil || ok {
		t.Errorf("Exists = %v, %v, want false", ok, err)
	}
	_, err = s.Load("go-1")
	wantCode(t, err, apperr.NotEnrolled)
	_, err = s.Update("go-1", func(p Progress) Progress { return p })
	wantCode(t, err, apperr.NotEnrolled)
}

func TestCorruptProgress(t *testing.T) {
	s, dir := newStore(t)
	writeFile(t, filepath.Join(dir, "go-1.json"), "{")

	_, err := s.Load("go-1")

	wantCode(t, err, apperr.ProgressReadFailed)
}

func TestUpdateSavesOnlyChanges(t *testing.T) {
	s, dir := newStore(t)
	path := filepath.Join(dir, "go-1.json")
	stored := `{"version": 1, "current": "hello", "completed": []}`
	writeFile(t, path, stored)

	p, err := s.Update("go-1", func(p Progress) Progress { return p })
	if err != nil || p.Current != "hello" {
		t.Fatalf("Update = %+v, %v", p, err)
	}
	if got := readFile(t, path); got != stored {
		t.Errorf("file = %s, want unchanged", got)
	}

	p, err = s.Update("go-1", func(p Progress) Progress { p.Current = "loops"; return p })
	if err != nil || p.Current != "loops" {
		t.Fatalf("Update = %+v, %v", p, err)
	}
	if got, _ := s.Load("go-1"); got.Current != "loops" {
		t.Errorf("stored current = %q, want loops", got.Current)
	}
}

func TestNewerProgressFormatIsNotWritten(t *testing.T) {
	s, dir := newStore(t)
	path := filepath.Join(dir, "go-1.json")
	newer := `{"version": 2, "current": "hello", "completed": [], "answers": {}}`
	writeFile(t, path, newer)

	p, err := s.Load("go-1")
	if err != nil || p.Current != "hello" {
		t.Errorf("Load = %+v, %v, want the known fields", p, err)
	}
	_, err = s.Update("go-1", func(p Progress) Progress { p.Current = "loops"; return p })
	wantCode(t, err, apperr.ProgressWriteFailed)
	if err := s.Enroll("go-1", "loops"); err != nil {
		t.Errorf("Enroll: %v", err)
	}

	if got := readFile(t, path); got != newer {
		t.Errorf("file = %s, want unchanged", got)
	}
}

func TestResolve(t *testing.T) {
	order := []string{"a", "b", "c", "d"}
	tests := []struct {
		name      string
		current   string
		completed []string
		want      string
	}{
		{name: "current kept", current: "c", completed: []string{"a", "b"}, want: "c"},
		{name: "after the last completed", current: "gone", completed: []string{"a", "b"}, want: "c"},
		{name: "furthest completed in new order", current: "gone", completed: []string{"c", "a"}, want: "d"},
		{name: "removed completed ignored", current: "gone", completed: []string{"a", "x"}, want: "b"},
		{name: "nothing completed left", current: "gone", completed: []string{"x"}, want: "a"},
		{name: "nothing completed", current: "gone", completed: []string{}, want: "a"},
		{name: "last completed is last", current: "gone", completed: []string{"a", "d"}, want: "d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Progress{Version: 1, Current: tt.current, Completed: tt.completed}

			got := Resolve(order, p)

			if got.Current != tt.want {
				t.Errorf("current = %q, want %q", got.Current, tt.want)
			}
			if !reflect.DeepEqual(got.Completed, tt.completed) {
				t.Errorf("completed = %v, want unchanged %v", got.Completed, tt.completed)
			}
		})
	}
}
