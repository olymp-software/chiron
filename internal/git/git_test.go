package git

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
)

// isolate makes the test independent of the machine's git configuration.
func isolate(t *testing.T) {
	t.Helper()
	global := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", global)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_SSH_COMMAND", "")
}

func newGit(t *testing.T, opts ...Option) *Git {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "hooks"), slog.New(slog.DiscardHandler), opts...)
}

func code(err error) apperr.Code {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	return ""
}

func envValue(env []string, name string) (string, bool) {
	for _, kv := range slices.Backward(env) {
		if v, ok := strings.CutPrefix(kv, name+"="); ok {
			return v, true
		}
	}
	return "", false
}

func TestCommandLine(t *testing.T) {
	isolate(t)
	t.Setenv("LC_ALL", "ru_RU.UTF-8")
	t.Setenv("GIT_TERMINAL_PROMPT", "1")
	g := newGit(t)
	dir := t.TempDir()
	cmd, err := g.command(context.Background(), dir, []string{"ls-remote", "--tags"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"-c", "core.autocrlf=false",
		"-c", "core.hooksPath=" + g.hooksDir,
		"-c", "filter.lfs.smudge=",
		"-c", "filter.lfs.process=",
		"-c", "filter.lfs.required=false",
		"-c", "protocol.allow=never",
		"-c", "protocol.https.allow=always",
		"-c", "protocol.ssh.allow=always",
		"-c", "advice.detachedHead=false",
		"ls-remote", "--tags",
	}
	if got := cmd.Args[1:]; !slices.Equal(got, want) {
		t.Errorf("args = %q\nwant %q", got, want)
	}
	if cmd.Dir != dir {
		t.Errorf("dir = %q, want %q", cmd.Dir, dir)
	}
	for name, want := range map[string]string{
		"GIT_TERMINAL_PROMPT": "0",
		"GIT_LFS_SKIP_SMUDGE": "1",
		"LC_ALL":              "C",
		"LANGUAGE":            "C",
		"GIT_SSH_COMMAND":     "ssh -o BatchMode=yes",
	} {
		if got, _ := envValue(cmd.Env, name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if _, ok := envValue(cmd.Env, "GIT_CONFIG_NOSYSTEM"); !ok {
		t.Error("inherited environment is not passed on")
	}
	if info, err := os.Stat(g.hooksDir); err != nil || !info.IsDir() {
		t.Errorf("hooks folder not created: %v", err)
	}
}

func TestCommandLineAllowFile(t *testing.T) {
	isolate(t)
	cmd, err := newGit(t, AllowFileProtocol()).command(context.Background(), "", []string{"status"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(cmd.Args, "protocol.file.allow=always") {
		t.Errorf("args = %q, want protocol.file.allow=always", cmd.Args)
	}
}

func TestSSHCommand(t *testing.T) {
	tests := []struct {
		name, env, config, want string
	}{
		{name: "default", want: "ssh -o BatchMode=yes"},
		{name: "config", config: "ssh -i ~/.ssh/course", want: "ssh -i ~/.ssh/course -o BatchMode=yes"},
		{name: "env", env: "ssh -p 2222", want: "ssh -p 2222 -o BatchMode=yes"},
		{name: "env wins over config", env: "ssh -p 2222", config: "ssh -i key", want: "ssh -p 2222 -o BatchMode=yes"},
		{name: "blank env", env: "  ", config: "ssh -i key", want: "ssh -i key -o BatchMode=yes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sshCommand(tt.env, tt.config); got != tt.want {
				t.Errorf("sshCommand(%q, %q) = %q, want %q", tt.env, tt.config, got, tt.want)
			}
		})
	}
}

func TestSSHCommandSources(t *testing.T) {
	tests := []struct {
		name, env, config, want string
	}{
		{name: "none", want: "ssh -o BatchMode=yes"},
		{name: "core.sshCommand", config: "ssh -i course_key", want: "ssh -i course_key -o BatchMode=yes"},
		{name: "GIT_SSH_COMMAND", env: "ssh -p 2222", config: "ssh -i course_key", want: "ssh -p 2222 -o BatchMode=yes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolate(t)
			if tt.config != "" {
				cfg := "[core]\n\tsshCommand = " + tt.config + "\n"
				if err := os.WriteFile(os.Getenv("GIT_CONFIG_GLOBAL"), []byte(cfg), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("GIT_SSH_COMMAND", tt.env)
			cmd, err := newGit(t).command(context.Background(), "", []string{"status"})
			if err != nil {
				t.Fatal(err)
			}
			if got, _ := envValue(cmd.Env, "GIT_SSH_COMMAND"); got != tt.want {
				t.Errorf("GIT_SSH_COMMAND = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGitNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := newGit(t).Run(context.Background(), "", "version")
	if code(err) != apperr.GitNotFound {
		t.Errorf("err = %v, want %s", err, apperr.GitNotFound)
	}
}

// fileURL returns the file:// URL of a local path.
func fileURL(path string) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return "file://" + p
}

func TestFileProtocol(t *testing.T) {
	isolate(t)
	ctx := context.Background()
	allowed := newGit(t, AllowFileProtocol())
	src := filepath.Join(t.TempDir(), "src.git")
	if _, err := allowed.Run(ctx, "", "init", "-q", "--bare", src); err != nil {
		t.Fatal(err)
	}

	_, err := newGit(t).Run(ctx, "", "clone", "-q", fileURL(src), filepath.Join(t.TempDir(), "refused"))
	if code(err) != apperr.GitFailed {
		t.Errorf("clone without the option: err = %v, want %s", err, apperr.GitFailed)
	}
	if _, err := allowed.Run(ctx, "", "clone", "-q", fileURL(src), filepath.Join(t.TempDir(), "allowed")); err != nil {
		t.Errorf("clone with the option: %v", err)
	}
}

func TestRunCancelled(t *testing.T) {
	isolate(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := newGit(t).Run(ctx, "", "version"); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
