//go:build unix

package runtime

import (
	"errors"
	"os/exec"
	"slices"
	"sync"
	"syscall"
	"time"
)

// tree is the process tree of a run: every external command starts in its
// own process group, and kill sends SIGKILL to each group.
type tree struct {
	// err is a failure to set the tree up; it is always nil on Unix.
	err error

	mu     sync.Mutex
	killed bool
	pgids  []int
}

func newTree() *tree { return &tree{} }

// prepare makes cmd start in a new process group.
func (t *tree) prepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// add registers the started cmd; a command started after kill is killed and
// still registered, so close waits for it too.
func (t *tree) add(cmd *exec.Cmd) error {
	pgid := cmd.Process.Pid
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.killed {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
	t.pgids = append(t.pgids, pgid)
	return nil
}

// release forgets the group of cmd, which has been waited for, once no
// process is left in it: its id may then be reused by an unrelated group,
// which kill must not hit. A group whose processes outlive cmd stays
// registered; the id cannot be reused while any of them runs.
func (t *tree) release(cmd *exec.Cmd) {
	pgid := cmd.Process.Pid
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := syscall.Kill(-pgid, 0); !errors.Is(err, syscall.ESRCH) {
		return
	}
	t.pgids = slices.DeleteFunc(t.pgids, func(p int) bool { return p == pgid })
}

// kill kills every registered process group.
func (t *tree) kill() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.killed {
		return
	}
	t.killed = true
	for _, pgid := range t.pgids {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}
}

// close kills what is left when the script exits and waits until it is gone:
// SIGKILL only marks a process for death.
// The groups are read anew on every check, so a command added meanwhile is
// waited for as well.
func (t *tree) close() {
	t.kill()
	deadline := time.Now().Add(closeWait)
	for t.anyAlive() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
}

// anyAlive reports whether a registered group still has a process.
func (t *tree) anyAlive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return slices.ContainsFunc(t.pgids, func(pgid int) bool { return syscall.Kill(-pgid, 0) == nil })
}
