package runtime

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func helperCommand(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), helperEnv+"=1")
	return cmd
}

// runCommand runs RunCommand in the background and returns its result
// channel.
func runCommand(ctx context.Context, cmd *exec.Cmd) <-chan error {
	ch := make(chan error, 1)
	go func() { ch <- RunCommand(ctx, cmd) }()
	return ch
}

func waitCommand(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(30 * time.Second):
		t.Fatal("command did not exit")
		return nil
	}
}

func TestRunCommandExitCode(t *testing.T) {
	err := waitCommand(t, runCommand(context.Background(), helperCommand(t, "exit", "3")))
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
		t.Errorf("err = %v, want exit code 3", err)
	}
	if err := waitCommand(t, runCommand(context.Background(), helperCommand(t, "exit", "0"))); err != nil {
		t.Errorf("err = %v, want nil", err)
	}
}

func TestRunCommandCancelKillsProcessTree(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := runCommand(ctx, helperCommand(t, "child", dir))
	child := readPid(t, filepath.Join(dir, "child.pid"))
	grandchild := readPid(t, filepath.Join(dir, "grandchild.pid"))
	cancel()
	if err := waitCommand(t, ch); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if processAlive(child) || processAlive(grandchild) {
		t.Errorf("child alive = %v, grandchild alive = %v after return, want both gone",
			processAlive(child), processAlive(grandchild))
	}
}

func TestRunCommandExitKillsLeftProcesses(t *testing.T) {
	dir := t.TempDir()
	ch := runCommand(context.Background(), helperCommand(t, "spawn", dir))
	if err := waitCommand(t, ch); err != nil {
		t.Fatalf("err = %v", err)
	}
	grandchild := readPid(t, filepath.Join(dir, "grandchild.pid"))
	if processAlive(grandchild) {
		t.Error("grandchild alive after return, want gone")
	}
}
