package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// An open file keeps git clean from deleting it on Windows; the switch has
// moved HEAD anyway, so the stored title follows the new version.
func TestSwitchVersionWithFileHeldOpen(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	held := filepath.Join(l.CopyPath(e.Key), "leftover", "held.txt")
	writeFile(t, held, "x")
	f, err := os.Open(held)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	got, err := l.SwitchVersion(context.Background(), e.Key, "v1.1.0")

	if err == nil {
		t.Error("SwitchVersion succeeded, want the clean error")
	}
	if got.Title != "Go v2" {
		t.Errorf("title = %q, want Go v2", got.Title)
	}
	if c := courseOf(t, l, e.Key); c.Title != "Go v2" {
		t.Errorf("stored title = %q, want Go v2", c.Title)
	}
}
