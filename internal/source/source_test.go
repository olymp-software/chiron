package source

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

// serveFile serves file:/// URLs from dir in-process, so tests need no system git.
func serveFile(t *testing.T, dir string) {
	t.Helper()
	client.InstallProtocol("file", server.NewClient(server.NewFilesystemLoader(osfs.New(dir))))
	t.Cleanup(func() { client.InstallProtocol("file", nil) })
}

// fixtureRepo creates a repository with one commit tagged with tags.
func fixtureRepo(t *testing.T, tags ...string) string {
	t.Helper()
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "course.yaml"), []byte("title: Fixture\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wt.Add("course.yaml"); err != nil {
		t.Fatal(err)
	}
	sig := &object.Signature{Name: "Author", Email: "author@example.com", When: time.Now()}
	head, err := wt.Commit("Initial", &git.CommitOptions{Author: sig})
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range tags {
		if _, err := repo.CreateTag(tag, head, nil); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestCloneTags(t *testing.T) {
	src := fixtureRepo(t, "v1.1.0", "v1.0.0")
	// The in-process loader treats a folder with a config file as a bare repository.
	serveFile(t, filepath.Join(src, ".git"))
	dst := filepath.Join(t.TempDir(), "clone")

	tags, err := CloneTags("file:///", dst)
	if err != nil {
		t.Fatalf("CloneTags: %v", err)
	}
	if want := []string{"v1.0.0", "v1.1.0"}; !slices.Equal(tags, want) {
		t.Errorf("tags = %v, want %v", tags, want)
	}
	if _, err := os.Stat(filepath.Join(dst, "course.yaml")); !os.IsNotExist(err) {
		t.Errorf("course.yaml in clone: err = %v, want not exist (no checkout)", err)
	}
}
