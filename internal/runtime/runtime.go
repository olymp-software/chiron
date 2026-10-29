// Package runtime runs course scripts with the embedded mvdan.cc/sh
// interpreter, the app's guaranteed utilities and process-tree termination.
package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	goruntime "runtime"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// waitDelay bounds how long a finished command's output is read when a
// process it started still holds the output pipe.
const waitDelay = time.Second

// abandonDelay is how long Stop gives the interpreter to return. A built-in
// blocked where the stop cannot reach it, such as opening a FIFO, is left
// behind and the run counts as stopped.
var abandonDelay = 3 * time.Second

// Spec is a script to run. Name is the script's file name used in parse
// errors; Env is the inherited environment as key=value pairs.
type Spec struct {
	Script []byte
	Name   string
	Dir    string
	Env    []string
}

// Result is the outcome of a finished run. Stopped tells that the run was
// ended by Stop; its exit code is then meaningless.
type Result struct {
	ExitCode int
	Stopped  bool
}

// Run is a running script with every process it started.
type Run struct {
	cancel   context.CancelFunc
	tree     *tree
	stopOnce sync.Once
	done     chan struct{}

	mu sync.Mutex
	// exited is set once the script has exited on its own or after Stop;
	// a later Stop no longer changes the result.
	exited  bool
	stopped bool
	// finished and res are set before done is closed.
	finished bool
	res      Result
}

// Start runs the script in the background. stdout and stderr go to out in the
// order they are written, so out must be safe for concurrent use; stdin is
// empty. A script that cannot be parsed writes the parse error to out and
// exits with code 2.
func Start(spec Spec, out io.Writer) *Run {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Run{cancel: cancel, done: make(chan struct{})}
	r.tree = newTree()
	go func() {
		code := r.run(ctx, spec, out)
		r.mu.Lock()
		r.exited = true
		stopped := r.stopped
		r.mu.Unlock()
		r.tree.close()
		cancel()
		r.finish(Result{ExitCode: code, Stopped: stopped})
	}()
	return r
}

// finish records the result and ends Wait; only the first call counts.
func (r *Run) finish(res Result) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.finished {
		return
	}
	r.finished = true
	r.res = res
	close(r.done)
}

// Wait waits for the script to exit and returns its result. Processes the
// script left running are killed when it exits.
func (r *Run) Wait() Result {
	<-r.done
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.res
}

// Stop kills the script with all its processes at once and returns without
// waiting. Stopping a finished or stopped run does nothing. If the
// interpreter does not return within abandonDelay, Wait returns a stopped
// result anyway.
func (r *Run) Stop() {
	r.stopOnce.Do(func() {
		r.mu.Lock()
		if r.exited {
			r.mu.Unlock()
			return
		}
		r.stopped = true
		r.mu.Unlock()
		r.cancel()
		r.tree.kill()
		go func() {
			t := time.NewTimer(abandonDelay)
			defer t.Stop()
			select {
			case <-r.done:
			case <-t.C:
				r.finish(Result{Stopped: true})
			}
		}()
	})
}

func (r *Run) run(ctx context.Context, spec Spec, out io.Writer) int {
	file, err := syntax.NewParser().Parse(bytes.NewReader(spec.Script), spec.Name)
	if err != nil {
		_, _ = fmt.Fprintln(out, err)
		return 2
	}
	if err := r.tree.err; err != nil {
		_, _ = fmt.Fprintf(out, "%s: %v\n", spec.Name, err)
		return 1
	}
	env := append(spec.Env[:len(spec.Env):len(spec.Env)],
		"HOST_OS="+goruntime.GOOS,
		"HOST_ARCH="+goruntime.GOARCH,
	)
	runner, err := interp.New(
		interp.Dir(spec.Dir),
		interp.Env(expand.ListEnviron(env...)),
		interp.StdIO(strings.NewReader(""), out, out),
		interp.ExecHandlers(utilitiesHandler, r.execHandler),
	)
	if err != nil {
		_, _ = fmt.Fprintf(out, "%s: %v\n", spec.Name, err)
		return 1
	}
	err = runner.Run(ctx, file)
	var status interp.ExitStatus
	switch {
	case err == nil:
		return 0
	case errors.As(err, &status):
		return int(status)
	case ctx.Err() != nil:
		return 1
	default:
		_, _ = fmt.Fprintf(out, "%s: %v\n", spec.Name, err)
		return 1
	}
}

// execHandler starts external commands as part of the run's process tree, so
// Stop kills them together with everything they start.
func (r *Run) execHandler(interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		hc := interp.HandlerCtx(ctx)
		path, err := interp.LookPathDir(hc.Dir, hc.Env, args[0])
		if err != nil {
			_, _ = fmt.Fprintln(hc.Stderr, err)
			return interp.ExitStatus(127)
		}
		cmd := exec.Command(path)
		cmd.Args = args
		cmd.Env = execEnv(hc.Env)
		cmd.Dir = hc.Dir
		cmd.Stdin = hc.Stdin
		cmd.Stdout = hc.Stdout
		cmd.Stderr = hc.Stderr
		cmd.WaitDelay = waitDelay
		r.tree.prepare(cmd)
		if err := cmd.Start(); err != nil {
			_, _ = fmt.Fprintln(hc.Stderr, err)
			return interp.ExitStatus(127)
		}
		if err := r.tree.add(cmd); err != nil {
			_ = cmd.Wait()
			_, _ = fmt.Fprintf(hc.Stderr, "%s: %v\n", args[0], err)
			return interp.ExitStatus(126)
		}
		err = cmd.Wait()
		r.tree.release(cmd)
		var exitErr *exec.ExitError
		switch {
		case err == nil, errors.Is(err, exec.ErrWaitDelay):
			return nil
		case errors.As(err, &exitErr):
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				return interp.ExitStatus(128 + int(ws.Signal()))
			}
			return interp.ExitStatus(exitErr.ExitCode())
		default:
			return err
		}
	}
}

// execEnv returns the exported variables of env as key=value pairs. It
// follows the unexported execEnv of mvdan.cc/sh/v3/interp, which the default
// exec handler uses; recheck it when upgrading that module. Unlike it, names
// compare case-insensitively on Windows, where the environment does.
func execEnv(env expand.Environ) []string {
	var list []string
	for name, vr := range env.Each {
		if !vr.IsSet() {
			// Unset in the script but inherited: drop the inherited value.
			list = slices.DeleteFunc(list, func(kv string) bool {
				n, _, _ := strings.Cut(kv, "=")
				return sameEnvName(n, name)
			})
		}
		if vr.Exported && vr.Kind == expand.String {
			list = append(list, name+"="+vr.String())
		}
	}
	return list
}

func sameEnvName(a, b string) bool {
	if goruntime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}
