package main

import (
	"context"
	"errors"
	"log/slog"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/library"
)

// App is the Wails-bound facade. Its methods delegate to internal packages.
// Wails turns a returned Go error into a plain string, so methods return
// results carrying a coded error instead.
type App struct {
	ctx context.Context
	lib *library.Library
	log *slog.Logger
}

// NewApp creates the bound App.
func NewApp(lib *library.Library, log *slog.Logger) *App {
	return &App{lib: lib, log: log}
}

// startup is called when the app starts; the context is kept for runtime calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// CoursesResult is the library list or the error that prevented reading it.
type CoursesResult struct {
	Courses []library.Entry `json:"courses"`
	Error   *apperr.Error   `json:"error,omitempty"`
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

// ListCourses returns the courses in the library.
func (a *App) ListCourses() CoursesResult {
	courses, err := a.lib.List()
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
func (a *App) SyncCourse(source string) CourseResult {
	e, err := a.lib.Sync(source)
	return CourseResult{Course: e, Error: a.coded(err)}
}

// CourseOverview returns the course overview of a course in the library.
func (a *App) CourseOverview(source string) OverviewResult {
	o, err := a.lib.Overview(source)
	return OverviewResult{Overview: o, Error: a.coded(err)}
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
