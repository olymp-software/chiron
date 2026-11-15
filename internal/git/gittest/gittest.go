// Package gittest builds git repositories in temp folders for tests: a bare
// repository that the code under test clones by its file:// URL, and a work
// repository the test commits, tags and pushes from.
package gittest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Repo is a bare repository with a work repository pushing to it.
type Repo struct {
	t *testing.T
	// Bare is the folder of the bare repository; URL is its file:// URL.
	Bare string
	URL  string
	// Work is the work repository; its branch main is pushed to Bare.
	Work string
}

// New creates an empty bare repository and a work repository for it.
func New(t *testing.T) *Repo {
	t.Helper()
	dir := t.TempDir()
	r := &Repo{t: t, Bare: filepath.Join(dir, "remote.git"), Work: filepath.Join(dir, "work")}
	r.URL = FileURL(r.Bare)
	r.git("", "init", "-q", "--bare", "-b", "main", r.Bare)
	r.git("", "init", "-q", "-b", "main", r.Work)
	r.git(r.Work, "remote", "add", "origin", r.URL)
	return r
}

// FileURL returns the file:// URL of a local path.
func FileURL(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + p
}

// Commit writes files (slash-separated paths relative to the root, with their
// content; "" content deletes the file), commits everything on main, pushes
// main and returns the commit.
func (r *Repo) Commit(files map[string]string) string {
	r.t.Helper()
	r.write(files)
	r.git(r.Work, "add", "-A")
	r.git(r.Work, "commit", "-q", "--allow-empty", "-m", "commit")
	r.git(r.Work, "push", "-q", "origin", "main")
	return r.Head()
}

// CommitDetached commits files on top of main without moving main and returns
// the commit; it is reachable only from tags pushed for it.
func (r *Repo) CommitDetached(files map[string]string) string {
	r.t.Helper()
	r.git(r.Work, "checkout", "-q", "--detach")
	r.write(files)
	r.git(r.Work, "add", "-A")
	r.git(r.Work, "commit", "-q", "--allow-empty", "-m", "detached")
	commit := r.Head()
	r.git(r.Work, "checkout", "-q", "-f", "main")
	return commit
}

// Tag creates an annotated tag name at commit, replacing a tag of that name,
// and force-pushes it.
func (r *Repo) Tag(name, commit string) {
	r.t.Helper()
	r.git(r.Work, "tag", "-f", "-a", "-m", name, name, commit)
	r.git(r.Work, "push", "-q", "-f", "origin", "refs/tags/"+name)
}

// Head returns the commit at the work repository's HEAD.
func (r *Repo) Head() string {
	r.t.Helper()
	return strings.TrimSpace(r.git(r.Work, "rev-parse", "HEAD"))
}

// Git runs git in dir with the test identity and the file protocol allowed,
// and returns its output.
func (r *Repo) Git(dir string, args ...string) string {
	r.t.Helper()
	return r.git(dir, args...)
}

func (r *Repo) write(files map[string]string) {
	r.t.Helper()
	for path, content := range files {
		p := filepath.Join(r.Work, filepath.FromSlash(path))
		if content == "" {
			if err := os.Remove(p); err != nil {
				r.t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
}

func (r *Repo) git(dir string, args ...string) string {
	r.t.Helper()
	args = append([]string{
		"-c", "user.name=Author", "-c", "user.email=author@example.com",
		"-c", "protocol.file.allow=always", "-c", "core.autocrlf=false",
		"-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false",
	}, args...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
