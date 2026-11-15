// Package git runs the git installed on the learner's machine for repository
// courses: non-interactively, with a fixed set of overridden settings and
// failures mapped to apperr codes. Authentication and transport stay with the
// learner's git configuration.
package git

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/runtime"
)

// Git runs git commands. Its methods are safe for concurrent use.
type Git struct {
	// hooksDir is an empty folder used as core.hooksPath, so no hook of the
	// learner's configuration runs.
	hooksDir  string
	log       *slog.Logger
	allowFile bool
}

// Option configures a Git.
type Option func(*Git)

// AllowFileProtocol also allows file:// repositories. Only tests set it: they
// work with repositories in a temp folder; the learner-facing path never
// allows file.
func AllowFileProtocol() Option {
	return func(g *Git) { g.allowFile = true }
}

// New returns a Git that uses hooksDir as its empty hooks folder; the folder
// is created when needed.
func New(hooksDir string, log *slog.Logger, opts ...Option) *Git {
	g := &Git{hooksDir: hooksDir, log: log}
	for _, o := range opts {
		o(g)
	}
	return g
}

// Run runs git with args in dir ("" for the app's working folder) and returns
// its standard output. Cancelling ctx kills git with every process it started
// and returns ctx.Err(). A missing git gives apperr.GitNotFound; a failure is
// mapped to an *apperr.Error by its error output, which is logged in full.
func (g *Git) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd, err := g.command(ctx, dir, args)
	if err != nil {
		return nil, err
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = runtime.RunCommand(ctx, cmd)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		g.log.Error("git failed", "args", args, "dir", dir, "err", err, "stderr", stderr.String())
		return nil, classify(stderr.String())
	}
	return stdout.Bytes(), nil
}

// command builds the git command line and environment for args.
func (g *Git) command(ctx context.Context, dir string, args []string) (*exec.Cmd, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		g.log.Error("find git", "err", err)
		return nil, apperr.New(apperr.GitNotFound)
	}
	if err := os.MkdirAll(g.hooksDir, 0o755); err != nil {
		g.log.Error("create git hooks folder", "err", err)
		return nil, apperr.New(apperr.Internal)
	}
	ssh := sshCommand(os.Getenv("GIT_SSH_COMMAND"), g.configSSHCommand(ctx, path))
	cmd := exec.Command(path, append(g.configArgs(), args...)...)
	cmd.Dir = dir
	cmd.Env = environ(os.Environ(), ssh)
	return cmd, nil
}

// configArgs returns the -c options that override the learner's git settings
// for every command.
func (g *Git) configArgs() []string {
	cfg := []string{
		"core.autocrlf=false",
		"core.hooksPath=" + g.hooksDir,
		"filter.lfs.smudge=",
		"filter.lfs.process=",
		"filter.lfs.required=false",
		"protocol.allow=never",
		"protocol.https.allow=always",
		"protocol.ssh.allow=always",
		"advice.detachedHead=false",
	}
	if g.allowFile {
		cfg = append(cfg, "protocol.file.allow=always")
	}
	args := make([]string, 0, 2*len(cfg))
	for _, c := range cfg {
		args = append(args, "-c", c)
	}
	return args
}

// environ returns base with the variables that keep git non-interactive and
// its messages in English; later values win over inherited ones.
func environ(base []string, ssh string) []string {
	return append(base[:len(base):len(base)],
		"GIT_TERMINAL_PROMPT=0",
		"GIT_LFS_SKIP_SMUDGE=1",
		"LC_ALL=C",
		"LANGUAGE=C",
		"GIT_SSH_COMMAND="+ssh,
	)
}

// sshCommand returns the ssh command git runs: the learner's own, from the
// GIT_SSH_COMMAND variable or else core.sshCommand, or plain ssh, with batch
// mode appended so ssh never asks for a passphrase, password or host key.
func sshCommand(env, config string) string {
	cmd := strings.TrimSpace(env)
	if cmd == "" {
		cmd = strings.TrimSpace(config)
	}
	if cmd == "" {
		cmd = "ssh"
	}
	return cmd + " -o BatchMode=yes"
}

// configSSHCommand returns core.sshCommand from the learner's git
// configuration, "" when it is unset or cannot be read.
func (g *Git) configSSHCommand(ctx context.Context, path string) string {
	cmd := exec.Command(path, "config", "--get", "core.sshCommand")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := runtime.RunCommand(ctx, cmd)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return ""
	}
	if err != nil {
		g.log.Warn("read core.sshCommand", "err", err, "stderr", stderr.String())
		return ""
	}
	return strings.TrimSpace(stdout.String())
}
