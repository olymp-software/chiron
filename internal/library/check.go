package library

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/runtime"
)

// Check statuses of the course overview; "" means the course has no check or
// its course.yaml is invalid.
const (
	CheckNotChecked = "notChecked"
	CheckPassed     = "passed"
	CheckChanged    = "changed"
)

// StartCheck starts the environment check of the course with key and returns
// without waiting. The script's output goes to out; done is called with the
// result after it has been remembered. The result is remembered by the hash
// of the check file as it was read for this run.
func (l *Library) StartCheck(key string, out io.Writer, done func(runtime.Result)) error {
	e, err := l.entry(key)
	if err != nil {
		return err
	}
	// The course lock keeps Sync from replacing the copy between reading
	// course.yaml and registering the run, so the check file is read from
	// the same copy as course.yaml and Sync always sees the run and stops it.
	unlock := l.lockCourse(e.Key)
	defer unlock()
	root := l.CopyPath(e.Key)
	c, err := manifest.ParseCourse(root)
	if err != nil {
		return err
	}
	if c.Check == "" {
		return errors.New("course declares no check")
	}
	script, err := readCheck(root, c.Check)
	if err != nil {
		l.log.Error("read check", "key", e.Key, "path", c.Check, "err", err)
		return apperr.New(apperr.CheckReadFailed, "path", c.Check)
	}
	hash := hashOf(script)

	l.checksMu.Lock()
	defer l.checksMu.Unlock()
	if _, ok := l.checks[e.Key]; ok {
		return apperr.New(apperr.CheckRunning)
	}
	run := runtime.Start(runtime.Spec{Script: script, Name: c.Check, Dir: root, Env: checkEnv()}, out)
	ch := &check{run: run, finished: make(chan struct{})}
	l.checks[e.Key] = ch
	l.log.Info("check started", "key", e.Key)
	go func() {
		res := run.Wait()
		l.log.Info("check exited", "key", e.Key, "code", res.ExitCode, "stopped", res.Stopped)
		l.rememberCheck(e.Key, hash, res)
		l.checksMu.Lock()
		delete(l.checks, e.Key)
		l.checksMu.Unlock()
		close(ch.finished)
		done(res)
	}()
	return nil
}

// check is a running environment check.
type check struct {
	run *runtime.Run
	// finished is closed once the result is remembered and the check is no
	// longer running, so a new check of the course can start.
	finished chan struct{}
}

// StopCheck stops the running check of the course with key, if any, and
// waits until it has finished; a stopped run always finishes within the
// runtime's abandon delay.
func (l *Library) StopCheck(key string) {
	l.checksMu.Lock()
	ch := l.checks[key]
	l.checksMu.Unlock()
	if ch != nil {
		ch.run.Stop()
		<-ch.finished
	}
}

// StopAll stops every running check and waits until they have finished.
func (l *Library) StopAll() {
	l.checksMu.Lock()
	checks := make([]*check, 0, len(l.checks))
	for _, ch := range l.checks {
		checks = append(checks, ch)
	}
	l.checksMu.Unlock()
	for _, ch := range checks {
		ch.run.Stop()
	}
	for _, ch := range checks {
		<-ch.finished
	}
}

// rememberCheck applies the result of a check run of the check file with
// hash: exit code 0 remembers hash, another exit code forgets it if it is
// the remembered one, a stopped run changes nothing.
func (l *Library) rememberCheck(key, hash string, res runtime.Result) {
	if res.Stopped {
		return
	}
	err := l.update(func(f *libraryFile) error {
		for i := range f.Courses {
			if f.Courses[i].Key != key {
				continue
			}
			switch {
			case res.ExitCode == 0:
				f.Courses[i].CheckHash = hash
			case f.Courses[i].CheckHash == hash:
				f.Courses[i].CheckHash = ""
			}
		}
		return nil
	})
	if err != nil {
		l.log.Error("remember check result", "key", key, "err", err)
	}
}

// checkStatus returns the check status of the course e with course.yaml c.
// An unreadable check file is logged and counts as not checked.
func (l *Library) checkStatus(e Entry, c *manifest.Course) string {
	if c.Check == "" {
		return ""
	}
	unlock := l.lockCourse(e.Key)
	script, err := readCheck(l.CopyPath(e.Key), c.Check)
	unlock()
	switch {
	case err != nil:
		l.log.Error("read check", "key", e.Key, "path", c.Check, "err", err)
		return CheckNotChecked
	case e.CheckHash == "":
		return CheckNotChecked
	case e.CheckHash == hashOf(script):
		return CheckPassed
	default:
		return CheckChanged
	}
}

// checkSatisfied reports whether the course e with course.yaml c may be
// enrolled in or entered: it has no check or the check passed.
func (l *Library) checkSatisfied(e Entry, c *manifest.Course) bool {
	s := l.checkStatus(e, c)
	return s == "" || s == CheckPassed
}

// readCheck reads the check file at the course path inside the course copy
// root; the read cannot leave the root.
func readCheck(root, path string) ([]byte, error) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return r.ReadFile(filepath.FromSlash(path))
}

func hashOf(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// checkEnv is the inherited environment without the activity variables,
// which the check does not get.
func checkEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.EqualFold(name, "ACTIVITY_DIR") || strings.EqualFold(name, "WORKSPACE_DIR") {
			continue
		}
		env = append(env, kv)
	}
	return env
}
