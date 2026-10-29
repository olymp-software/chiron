//go:build unix

package runtime

import (
	"os/exec"
	"sync"
	"syscall"
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

// add registers the started cmd; a command started after kill is killed.
func (t *tree) add(cmd *exec.Cmd) error {
	pgid := cmd.Process.Pid
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.killed {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		return nil
	}
	t.pgids = append(t.pgids, pgid)
	return nil
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

// close kills what is left when the script exits.
func (t *tree) close() { t.kill() }
