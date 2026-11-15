package runtime

import (
	"context"
	"os/exec"
)

// RunCommand starts cmd in a process tree of its own and waits for it, the way
// a script's external commands run: on Windows without a console window.
// Cancelling ctx kills cmd with every process it started and makes
// RunCommand return ctx.Err(). Processes cmd leaves running are killed when
// it exits; when RunCommand returns they are gone. RunCommand sets
// cmd.SysProcAttr, and cmd.WaitDelay when it is zero.
func RunCommand(ctx context.Context, cmd *exec.Cmd) error {
	t := newTree()
	if t.err != nil {
		return t.err
	}
	defer t.close()
	t.prepare(cmd)
	if cmd.WaitDelay == 0 {
		cmd.WaitDelay = waitDelay
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := t.add(cmd); err != nil {
		_ = cmd.Wait()
		return err
	}
	stop := context.AfterFunc(ctx, t.kill)
	defer stop()
	err := cmd.Wait()
	t.release(cmd)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}
