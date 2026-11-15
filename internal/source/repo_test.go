package source

import (
	"errors"
	"regexp"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
)

func TestNormalizeRepo(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		// Rows of the design table.
		{addr: "https://Host/Path/", want: "https://host/Path"},
		{addr: "ssh://User@Host:22/Path", want: "ssh://User@host:22/Path"},
		{addr: "user@host:path", want: "ssh://user@host/path"},
		{addr: "github.com/o/r", want: "https://github.com/o/r"},
		// Spec scenarios.
		{addr: "HTTPS://GitHub.com/Owner/Repo.git/", want: "https://github.com/Owner/Repo.git"},
		{addr: "HTTPS://Example.com/Team/Course.git/", want: "https://example.com/Team/Course.git"},
		{addr: "github.com/owner/repo", want: "https://github.com/owner/repo"},
		{addr: "git@gitlab.example.com:team/course.git", want: "ssh://git@gitlab.example.com/team/course.git"},
		// More forms.
		{addr: "  https://github.com/owner/repo  ", want: "https://github.com/owner/repo"},
		{addr: "GitHub.com/Owner/Repo", want: "https://github.com/Owner/Repo"},
		{addr: "git@github.com:/owner/repo.git", want: "ssh://git@github.com/owner/repo.git"},
		{addr: "ssh://gitlab.example.com/team/course", want: "ssh://gitlab.example.com/team/course"},
		{addr: "https://user:token@Example.com/team/course", want: "https://user:token@example.com/team/course"},
		{addr: "https://example.com:8443/team/course//", want: "https://example.com:8443/team/course"},
	}
	for _, tt := range tests {
		got, err := NormalizeRepo(tt.addr)
		if err != nil || got != tt.want {
			t.Errorf("NormalizeRepo(%q) = %q, %v, want %q", tt.addr, got, err, tt.want)
		}
	}
}

func TestNormalizeRepoInvalid(t *testing.T) {
	for _, addr := range []string{
		"",
		"http://example.com/course.git",
		"git://example.com/course.git",
		"file:///home/learner/course",
		"ftp://example.com/course",
		"https://",
		"https://example.com",
		"https://example.com/",
		"https:///path",
		"ssh://@host/path",
		"https://example.com/course?ref=v1",
		"https://example.com/course#v1",
		"https://example.com/my course",
		"example.com/owner/repo",
		"host:path",
		`C:\Courses\course`,
		"C:/Courses/course",
		"/home/learner/course",
		"owner/repo",
		"@host:path",
		"user@host:",
	} {
		_, err := NormalizeRepo(addr)
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Code != apperr.RepoAddressInvalid || ae.Params["address"] != addr {
			t.Errorf("NormalizeRepo(%q) err = %v, want %s with the address", addr, err, apperr.RepoAddressInvalid)
		}
	}
}

func mustNormalize(t *testing.T, addr string) string {
	t.Helper()
	s, err := NormalizeRepo(addr)
	if err != nil {
		t.Fatalf("NormalizeRepo(%q): %v", addr, err)
	}
	return s
}

func TestSameRepoSource(t *testing.T) {
	same := [][2]string{
		{"https://GitHub.com/Owner/Repo.git/", "https://github.com/owner/repo"},
		{"github.com/owner/repo", "https://github.com/owner/repo"},
		{"https://example.com/team/course", "https://example.com/Team/Course.git"},
	}
	for _, p := range same {
		if !SameSource(mustNormalize(t, p[0]), mustNormalize(t, p[1])) {
			t.Errorf("%q and %q are different courses, want the same", p[0], p[1])
		}
	}
	different := [][2]string{
		{"git@github.com:owner/repo.git", "https://github.com/owner/repo"},
		{"github.com/other/repo", "github.com/owner/repo"},
	}
	for _, p := range different {
		if SameSource(mustNormalize(t, p[0]), mustNormalize(t, p[1])) {
			t.Errorf("%q and %q are the same course, want different", p[0], p[1])
		}
	}
}

// The repository rule does not depend on the OS, so these keys are the same
// in the Linux and the Windows test run.
func TestRepoKey(t *testing.T) {
	a := Key(mustNormalize(t, "https://GitHub.com/Owner/Repo.git/"))
	b := Key(mustNormalize(t, "github.com/owner/repo"))
	if a != b {
		t.Errorf("keys %q and %q differ, want the same", a, b)
	}
	// Literal on purpose: stored data is found by the key.
	if want := "repo-d7606d02"; a != want {
		t.Errorf("Key = %q, want %q", a, want)
	}
	if !regexp.MustCompile(`^repo-[0-9a-f]{8}$`).MatchString(Key("https://github.com/owner/repo.git")) {
		t.Errorf("Key = %q, want repo-<hash>", Key("https://github.com/owner/repo.git"))
	}
	if s := Key(mustNormalize(t, "git@github.com:owner/repo.git")); s == a {
		t.Errorf("ssh and https give the same key %q, want different", s)
	}
}

func TestRepoLabel(t *testing.T) {
	for source, want := range map[string]string{
		"https://github.com/owner/repo.git":   "repo",
		"https://github.com/owner/Repo.GIT":   "repo",
		"ssh://git@host/team/course.git":      "course",
		"https://github.com/owner/repo":       "repo",
		"https://github.com/owner/.git":       "git",
		"https://github.com/owner/my.git.git": "my-git",
	} {
		if got := label(source); got != want {
			t.Errorf("label(%q) = %q, want %q", source, got, want)
		}
	}
	// Local folders keep .git in the label.
	if got := label("/home/learner/course.git"); got != "course-git" {
		t.Errorf("label of a local folder = %q, want course-git", got)
	}
}
