package library

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/runtime"
)

// Kinds of scripts in Runs.
const (
	runControl         = "control"
	runActivityCheck   = "activityCheck"
	runWorkshopControl = "workshopControl"
)

// slot is what a run occupies: two runs with the same slot never run at once.
// Controls of an activity share a slot by exclusion group; a course has one
// Check and one workshop control slot.
type slot struct {
	course string
	kind   string
	slug   string
	group  string
}

// scriptRun is a running script.
type scriptRun struct {
	slot slot
	run  *runtime.Run
	// finished is closed once the run has ended and its TEMP_DIR is gone.
	finished chan struct{}
}

// Runs holds the running scripts of open activities and workshop controls;
// the environment check is not one of them. Exclusion is decided here, so a
// double click cannot start a second run.
type Runs struct {
	mu     sync.Mutex
	active []*scriptRun
}

// start starts spec in slot s unless the slot is taken. After the run ends,
// cleanup runs and then done gets the result. The returned function stops
// this run only, so a late stop never hits a newer run in the same slot.
func (r *Runs) start(s slot, spec runtime.Spec, out io.Writer, cleanup func(), done func(runtime.Result)) (func(), error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if slices.ContainsFunc(r.active, func(a *scriptRun) bool { return a.slot == s }) {
		return nil, apperr.New(apperr.ScriptRunning)
	}
	sr := &scriptRun{slot: s, run: runtime.Start(spec, out), finished: make(chan struct{})}
	r.active = append(r.active, sr)
	go func() {
		res := sr.run.Wait()
		r.remove(sr)
		cleanup()
		close(sr.finished)
		done(res)
	}()
	return func() { r.stop(func(a *scriptRun) bool { return a == sr }) }, nil
}

func (r *Runs) remove(sr *scriptRun) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.active = slices.DeleteFunc(r.active, func(a *scriptRun) bool { return a == sr })
}

// stop stops the matching runs and waits until they have finished; a stopped
// run always finishes within the runtime's abandon delay.
func (r *Runs) stop(match func(*scriptRun) bool) {
	r.mu.Lock()
	var runs []*scriptRun
	for _, a := range r.active {
		if match(a) {
			runs = append(runs, a)
		}
	}
	r.mu.Unlock()
	for _, a := range runs {
		a.run.Stop()
	}
	for _, a := range runs {
		<-a.finished
	}
}

func (r *Runs) stopCourse(key string) {
	r.stop(func(a *scriptRun) bool { return a.slot.course == key })
}

// StartControl starts the control with index of the activity slug of the
// enrolled workshop with key and returns without waiting. The script's output
// goes to out; done is called with the result after TEMP_DIR is deleted. The
// returned function stops the run and waits until it has finished.
func (l *Library) StartControl(key, slug string, index int, out io.Writer, done func(runtime.Result)) (func(), error) {
	return l.withActivity(key, slug, func(e Entry, root, section string, a *manifest.WorkshopActivity) (func(), error) {
		if index < 0 || index >= len(a.Controls) {
			return nil, errors.New("no such control")
		}
		c := a.Controls[index]
		s := slot{course: e.Key, kind: runControl, slug: slug, group: c.Concurrency}
		return l.startActivityScript(s, root, section, slug, c.Script, out, done)
	})
}

// StopActivityRuns stops every script of the workshop with key and waits until
// they have finished: the learner is leaving the activity.
func (l *Library) StopActivityRuns(key string) {
	l.runs.stopCourse(key)
}

// StartActivityCheck starts the Check of the activity slug of the enrolled
// workshop with key beside its controls and returns without waiting. Nothing
// about the result is remembered. The returned function is as in StartControl.
func (l *Library) StartActivityCheck(key, slug string, out io.Writer, done func(runtime.Result)) (func(), error) {
	return l.withActivity(key, slug, func(e Entry, root, section string, a *manifest.WorkshopActivity) (func(), error) {
		if a.Check == "" {
			return nil, errors.New("activity declares no check")
		}
		s := slot{course: e.Key, kind: runActivityCheck}
		return l.startActivityScript(s, root, section, slug, a.Check, out, done)
	})
}

// StartWorkshopControl starts the control with index of workshop.yaml of the
// enrolled workshop with key and returns without waiting. It does not depend
// on the environment check and nothing about the result is remembered. The
// returned function is as in StartControl.
func (l *Library) StartWorkshopControl(key string, index int, out io.Writer, done func(runtime.Result)) (func(), error) {
	e, err := l.entry(key)
	if err != nil {
		return nil, err
	}
	if !l.enrolled(e.Key) {
		return nil, apperr.New(apperr.NotEnrolled)
	}
	unlock := l.lockCourse(e.Key)
	defer unlock()
	root := l.CopyPath(e.Key)
	c, err := manifest.Parse(root)
	if err != nil {
		return nil, err
	}
	if index < 0 || index >= len(c.Controls) {
		return nil, errors.New("no such workshop control")
	}
	s := slot{course: e.Key, kind: runWorkshopControl}
	return l.startScript(s, root, c.Controls[index].Script, root, nil, out, done)
}

// withActivity calls f with the parsed available activity slug of the enrolled
// workshop with key. The course lock keeps Sync from replacing the copy until
// a run started by f is registered, so Sync always sees it and stops it.
func (l *Library) withActivity(key, slug string,
	f func(e Entry, root, section string, a *manifest.WorkshopActivity) (func(), error)) (func(), error) {
	e, err := l.entry(key)
	if err != nil {
		return nil, err
	}
	unlock := l.lockCourse(e.Key)
	defer unlock()
	root := l.CopyPath(e.Key)
	c, err := manifest.Parse(root)
	if err != nil {
		return nil, err
	}
	if c.Type != manifest.ContentWorkshop {
		return nil, errors.New("scripts of course activities are not supported")
	}
	section, _, _, _, err := l.availableActivity(e.Key, c, slug)
	if err != nil {
		return nil, err
	}
	a, err := manifest.ParseWorkshopActivity(root, section, slug)
	if err != nil {
		return nil, err
	}
	return f(e, root, section, a)
}

// startActivityScript starts script, a path from the activity folder.
func (l *Library) startActivityScript(s slot, root, section, slug, script string,
	out io.Writer, done func(runtime.Result)) (func(), error) {
	dir := filepath.Join(root, section, slug)
	return l.startScript(s, root, path.Join(section, slug, script), dir,
		[]string{"ACTIVITY_DIR=" + dir}, out, done)
}

// startScript starts the script at the course path script in the course copy
// root with working directory dir and a new TEMP_DIR, which is deleted after
// the run.
func (l *Library) startScript(s slot, root, script, dir string, env []string,
	out io.Writer, done func(runtime.Result)) (func(), error) {
	body, err := readCheck(root, script)
	if err != nil {
		l.log.Error("read script", "key", s.course, "path", script, "err", err)
		return nil, apperr.New(apperr.ScriptStartFailed, "path", script)
	}
	folder, err := l.EnsureWorkshopFolder(s.course)
	if err != nil {
		return nil, err
	}
	temp, err := l.newTempDir(s.course)
	if err != nil {
		l.log.Error("create temp folder", "key", s.course, "err", err)
		return nil, apperr.New(apperr.ScriptStartFailed, "path", script)
	}
	spec := runtime.Spec{
		Script: body,
		Name:   path.Base(script),
		Dir:    dir,
		Env: append(append(checkEnv(), env...),
			"WORKSPACE_DIR="+folder, "CONTENT_DIR="+root, "TEMP_DIR="+temp),
	}
	cleanup := func() { l.removeTemp(temp) }
	stop, err := l.runs.start(s, spec, out, cleanup, done)
	if err != nil {
		cleanup()
		return nil, err
	}
	l.log.Info("script started", "key", s.course, "path", script)
	return stop, nil
}

// newTempDir creates the empty TEMP_DIR of a run of the course with key.
func (l *Library) newTempDir(key string) (string, error) {
	base := filepath.Join(l.dataDir, tempDir, key)
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(base, "run-")
}

// removeTemp deletes the TEMP_DIR of a finished run. A program that has just
// exited from it stays locked for a moment on Windows, so the delete is
// retried as a rename is; what is left then goes on the next start.
func (l *Library) removeTemp(temp string) {
	var err error
	for i := range renameAttempts {
		if err = os.RemoveAll(temp); err == nil {
			return
		}
		if i < renameAttempts-1 {
			time.Sleep(renameDelay)
		}
	}
	l.log.Error("remove temp folder", "path", temp, "err", err)
}

// checkEnv is the inherited environment without the variables the app sets
// for activity scripts, which the environment check does not get.
func checkEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if slices.ContainsFunc(scriptVars, func(v string) bool { return strings.EqualFold(name, v) }) {
			continue
		}
		env = append(env, kv)
	}
	return env
}

var scriptVars = []string{"ACTIVITY_DIR", "WORKSPACE_DIR", "CONTENT_DIR", "TEMP_DIR"}
