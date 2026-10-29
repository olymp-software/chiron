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

	// stopped and res are written before done is closed.
	mu      sync.Mutex
	stopped bool
	res     Result
}

// Start runs the script in the background. stdout and stderr go to out in the
// order they are written; stdin is empty. A script that cannot be parsed
// writes the parse error to out and exits with code 2.
func Start(spec Spec, out io.Writer) *Run {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Run{cancel: cancel, done: make(chan struct{})}
	r.tree = newTree()
	go func() {
		defer close(r.done)
		code := r.run(ctx, spec, &lockedWriter{w: out})
		r.tree.close()
		cancel()
		r.mu.Lock()
		r.res = Result{ExitCode: code, Stopped: r.stopped}
		r.mu.Unlock()
	}()
	return r
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
// waiting. Stopping a finished or stopped run does nothing.
func (r *Run) Stop() {
	r.stopOnce.Do(func() {
		r.mu.Lock()
		r.stopped = true
		r.mu.Unlock()
		r.cancel()
		r.tree.kill()
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

// execEnv returns the exported variables of env as key=value pairs, like the
// interpreter's default exec handler does.
func execEnv(env expand.Environ) []string {
	var list []string
	for name, vr := range env.Each {
		if !vr.IsSet() {
			// Unset in the script but inherited: drop the inherited value.
			for i, kv := range list {
				if strings.HasPrefix(kv, name+"=") {
					list[i] = ""
				}
			}
		}
		if vr.Exported && vr.Kind == expand.String {
			list = append(list, name+"="+vr.String())
		}
	}
	return list
}

// lockedWriter serializes writes from the interpreter and the copying
// goroutines of external commands.
type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}
