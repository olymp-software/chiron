package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/wailsapp/wails/v2/pkg/options"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/library"
	"github.com/olymp-software/chiron/internal/platform"
	"github.com/olymp-software/chiron/internal/runtime"
	"github.com/olymp-software/chiron/internal/settings"
)

// Events of every running script, told apart by run id; their payloads are
// CheckOutput and CheckExit.
const (
	checkOutputEvent = "check:output"
	checkExitEvent   = "check:exit"
)

// App is the Wails-bound facade. Its methods delegate to internal packages.
// Wails turns a returned Go error into a plain string, so methods return
// results carrying a coded error instead.
type App struct {
	ctx      context.Context
	lib      *library.Library
	settings *settings.Store
	log      *slog.Logger
	// runIDs numbers script runs, so the UI can tell events of runs apart.
	runIDs atomic.Int64
	// stops holds the stop function of each running script by run id.
	stops sync.Map
}

// NewApp creates the bound App.
func NewApp(lib *library.Library, st *settings.Store, log *slog.Logger) *App {
	return &App{lib: lib, settings: st, log: log}
}

// startup is called when the app starts; the context is kept for runtime calls.
// The library is repaired only here: a second instance quits before startup,
// so it never touches what the running one uses.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.lib.Recover(); err != nil {
		a.log.Error("repair library", "err", err)
	}
}

// secondInstance brings the window forward when the app is started again.
func (a *App) secondInstance(options.SecondInstanceData) {
	wailsruntime.WindowUnminimise(a.ctx)
	wailsruntime.WindowShow(a.ctx)
}

// ErrorResult is the error of an operation that returns nothing else.
type ErrorResult struct {
	Error *apperr.Error `json:"error,omitempty"`
}

// CoursesResult is the library list or the error that prevented reading it.
type CoursesResult struct {
	Courses []library.Course `json:"courses"`
	Error   *apperr.Error    `json:"error,omitempty"`
}

// CourseResult is one course or the error of the operation on it.
type CourseResult struct {
	Course library.Entry `json:"course"`
	Error  *apperr.Error `json:"error,omitempty"`
}

// OverviewResult is a course overview or the error that prevented reading it.
// A broken course is not an error here: its overview carries the course error.
type OverviewResult struct {
	Overview library.Overview `json:"overview"`
	Error    *apperr.Error    `json:"error,omitempty"`
}

// EnterResult is the current activity of an entered course or the error that
// prevented entering it.
type EnterResult struct {
	Current string        `json:"current"`
	Error   *apperr.Error `json:"error,omitempty"`
}

// TreeResult is the course tree of an enrolled course or the error that
// prevented reading it.
type TreeResult struct {
	Tree  library.Tree  `json:"tree"`
	Error *apperr.Error `json:"error,omitempty"`
}

// ActivityResult is an activity of an enrolled course or the error that
// prevented reading it. A broken activity is not an error here: the activity
// carries the activity error.
type ActivityResult struct {
	Activity library.Activity `json:"activity"`
	Error    *apperr.Error    `json:"error,omitempty"`
}

// CompleteResult is the activity to open after completing one ("" after the
// last activity) with the updated course tree, or the error that prevented
// completing it.
type CompleteResult struct {
	Next  string        `json:"next"`
	Tree  library.Tree  `json:"tree"`
	Error *apperr.Error `json:"error,omitempty"`
}

// StartCheckResult is the run id of a started environment check or the error
// that prevented starting it.
type StartCheckResult struct {
	RunID int64         `json:"runId"`
	Error *apperr.Error `json:"error,omitempty"`
}

// CheckOutput is a batch of decoded output of the check run RunID.
type CheckOutput struct {
	RunID int64  `json:"runId"`
	Text  string `json:"text"`
}

// CheckExit is the result of the check run RunID, sent after its output.
type CheckExit struct {
	RunID    int64 `json:"runId"`
	ExitCode int   `json:"exitCode"`
	Stopped  bool  `json:"stopped"`
}

// ListCourses returns the courses in the library.
func (a *App) ListCourses() CoursesResult {
	courses, err := a.lib.Courses()
	return CoursesResult{Courses: courses, Error: a.coded(err)}
}

// PickFolder asks the learner for a folder and returns its path, or "" if
// the dialog was cancelled. title is the dialog title in the UI language.
func (a *App) PickFolder(title string) string {
	path, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{Title: title})
	if err != nil {
		a.log.Error("folder dialog", "err", err)
		return ""
	}
	return path
}

// AddLocalCourse adds the course in a local folder; the learner has already
// confirmed the script warning.
func (a *App) AddLocalCourse(path string) CourseResult {
	e, err := a.lib.AddLocal(path)
	return CourseResult{Course: e, Error: a.coded(err)}
}

// SyncCourse refreshes the course copy of a local-folder course.
func (a *App) SyncCourse(key string) CourseResult {
	e, err := a.lib.Sync(key)
	return CourseResult{Course: e, Error: a.coded(err)}
}

// RemoveResult is the paths left on disk by a removal, or the error that
// prevented it; with an error nothing was removed.
type RemoveResult struct {
	Left  []string      `json:"left"`
	Error *apperr.Error `json:"error,omitempty"`
}

// RemoveCourse removes a library item, keeping its progress and its learner
// folders or workshop folder if asked. The library item is removed even if
// some of its files are left on disk. A removed course entered last is
// forgotten, so a re-added one is not entered on startup.
func (a *App) RemoveCourse(key string, keepProgress, keepFiles bool) RemoveResult {
	left, err := a.lib.Remove(key, keepProgress, keepFiles)
	if err != nil {
		return RemoveResult{Left: []string{}, Error: a.coded(err)}
	}
	if a.settings.Get().LastCourse == key {
		if err := a.settings.Update(func(st *settings.Settings) {
			if st.LastCourse == key {
				st.LastCourse = ""
			}
		}); err != nil {
			a.log.Error("forget course entered last", "key", key, "err", err)
		}
	}
	return RemoveResult{Left: left}
}

// CourseOverview returns the course overview of a course in the library.
func (a *App) CourseOverview(key string) OverviewResult {
	o, err := a.lib.Overview(key)
	return OverviewResult{Overview: o, Error: a.coded(err)}
}

// EnrollCourse enrolls a course in the library.
func (a *App) EnrollCourse(key string) ErrorResult {
	return ErrorResult{Error: a.coded(a.lib.Enroll(key))}
}

// EnterCourse enters an enrolled course and returns its current activity. The
// course is remembered as the course entered last.
func (a *App) EnterCourse(key string) EnterResult {
	cur, err := a.lib.Enter(key)
	if err != nil {
		return EnterResult{Error: a.coded(err)}
	}
	if err := a.settings.Update(func(st *settings.Settings) { st.LastCourse = key }); err != nil {
		a.log.Error("save course entered last", "key", key, "err", err)
	}
	return EnterResult{Current: cur}
}

// CourseTree returns the course tree of an enrolled course.
func (a *App) CourseTree(key string) TreeResult {
	t, err := a.lib.Tree(key)
	return TreeResult{Tree: t, Error: a.coded(err)}
}

// Activity returns an available activity of an enrolled course.
func (a *App) Activity(key, slug string) ActivityResult {
	act, err := a.lib.Activity(key, slug)
	return ActivityResult{Activity: act, Error: a.coded(err)}
}

// CompleteActivity completes an available activity of an enrolled course: the
// effect of Next and Finish course.
func (a *App) CompleteActivity(key, slug string) CompleteResult {
	next, t, err := a.lib.CompleteActivity(key, slug)
	return CompleteResult{Next: next, Tree: t, Error: a.coded(err)}
}

// StartCheck starts the environment check of a course. Its output and result
// arrive as check:output and check:exit events with the returned run id;
// output may arrive before StartCheck returns.
func (a *App) StartCheck(key string) StartCheckResult {
	return a.startRun(func(out io.Writer, done func(runtime.Result)) (func(), error) {
		return nil, a.lib.StartCheck(key, out, done)
	})
}

// startRun starts a script through start and sends its output and result as
// check:output and check:exit events with the returned run id. A run whose
// start returns a stop function can be stopped by StopRun with its run id.
func (a *App) startRun(start func(out io.Writer, done func(runtime.Result)) (func(), error)) StartCheckResult {
	id := a.runIDs.Add(1)
	b := runtime.NewBatcher(func(text string) {
		wailsruntime.EventsEmit(a.ctx, checkOutputEvent, CheckOutput{RunID: id, Text: text})
	})
	var finished atomic.Bool
	stop, err := start(b, func(res runtime.Result) {
		finished.Store(true)
		a.stops.Delete(id)
		b.Close()
		wailsruntime.EventsEmit(a.ctx, checkExitEvent, CheckExit{RunID: id, ExitCode: res.ExitCode, Stopped: res.Stopped})
	})
	if err != nil {
		b.Close()
		return StartCheckResult{Error: a.coded(err)}
	}
	if stop != nil {
		a.stops.Store(id, stop)
		// The run may have finished before its stop function was stored.
		if finished.Load() {
			a.stops.Delete(id)
		}
	}
	return StartCheckResult{RunID: id}
}

// StopRun stops the script run with the run id, if it is still running, and
// returns once it has exited. It never stops another run, even of the same
// control.
func (a *App) StopRun(runID int64) ErrorResult {
	if stop, ok := a.stops.LoadAndDelete(runID); ok {
		stop.(func())()
	}
	return ErrorResult{}
}

// CancelCheck stops the running environment check of a course and returns
// once it has exited; the remembered result is unchanged.
func (a *App) CancelCheck(key string) ErrorResult {
	a.lib.StopCheck(key)
	return ErrorResult{}
}

// StartControl starts a control of an activity of a workshop. Its events are
// those of StartCheck; StopRun stops it.
func (a *App) StartControl(key, slug string, index int) StartCheckResult {
	return a.startRun(func(out io.Writer, done func(runtime.Result)) (func(), error) {
		return a.lib.StartControl(key, slug, index, out, done)
	})
}

// StopActivityRuns stops every script of a workshop and returns once they have
// exited: the learner has confirmed leaving the activity.
func (a *App) StopActivityRuns(key string) ErrorResult {
	a.lib.StopActivityRuns(key)
	return ErrorResult{}
}

// StartActivityCheck starts the Check of an activity of a workshop. Its events
// are those of StartCheck; StopRun stops it.
func (a *App) StartActivityCheck(key, slug string) StartCheckResult {
	return a.startRun(func(out io.Writer, done func(runtime.Result)) (func(), error) {
		return a.lib.StartActivityCheck(key, slug, out, done)
	})
}

// StartWorkshopControl starts a control of workshop.yaml. Its events are those
// of StartCheck; StopRun stops it.
func (a *App) StartWorkshopControl(key string, index int) StartCheckResult {
	return a.startRun(func(out io.Writer, done func(runtime.Result)) (func(), error) {
		return a.lib.StartWorkshopControl(key, index, out, done)
	})
}

// OpenFolder opens the workshop folder in the file manager. There is no editor
// setting yet, and the file manager is what the button opens without one.
func (a *App) OpenFolder(key string) ErrorResult {
	folder, err := a.lib.EnsureWorkshopFolder(key)
	if err != nil {
		return ErrorResult{Error: a.coded(err)}
	}
	if err := platform.OpenFolder(folder); err != nil {
		a.log.Error("open workshop folder", "path", folder, "err", err)
		return ErrorResult{Error: apperr.New(apperr.WorkshopFolderFailed)}
	}
	return ErrorResult{}
}

// StartCourse returns the storage key of the course to enter on startup: the
// course entered last if it is still in the library and enrolled, else "".
func (a *App) StartCourse() string {
	key := a.settings.Get().LastCourse
	if key == "" {
		return ""
	}
	courses, err := a.lib.Courses()
	if err != nil {
		a.log.Error("list courses on startup", "err", err)
		return ""
	}
	for _, c := range courses {
		if c.Key == key && c.Enrolled {
			return key
		}
	}
	return ""
}

// coded returns err as a coded error. Errors without a code are bugs; they
// are logged and shown as internal.
func (a *App) coded(err error) *apperr.Error {
	if err == nil {
		return nil
	}
	var ae *apperr.Error
	if errors.As(err, &ae) {
		a.log.Info("error shown", "err", ae)
		return ae
	}
	a.log.Error("uncoded error", "err", err)
	return apperr.New(apperr.Internal)
}
