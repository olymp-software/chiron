package main

import (
	"context"
	"errors"
	"log/slog"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/library"
	"github.com/olymp-software/chiron/internal/settings"
)

// App is the Wails-bound facade. Its methods delegate to internal packages.
// Wails turns a returned Go error into a plain string, so methods return
// results carrying a coded error instead.
type App struct {
	ctx      context.Context
	lib      *library.Library
	settings *settings.Store
	log      *slog.Logger
}

// NewApp creates the bound App.
func NewApp(lib *library.Library, st *settings.Store, log *slog.Logger) *App {
	return &App{lib: lib, settings: st, log: log}
}

// startup is called when the app starts; the context is kept for runtime calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
