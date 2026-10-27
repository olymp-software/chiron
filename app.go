package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/olymp-software/chiron/internal/runtime"
	"github.com/olymp-software/chiron/internal/source"
)

// App is the Wails-bound facade. Its methods delegate to internal packages.
type App struct {
	ctx     context.Context
	dataDir string
	log     *slog.Logger
}

// NewApp creates the bound App.
func NewApp(dataDir string, log *slog.Logger) *App {
	return &App{dataDir: dataDir, log: log}
}

// startup is called when the app starts; the context is kept for runtime calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SmokeRunScript runs script in a fresh <data>/smoke/script folder.
func (a *App) SmokeRunScript(script string) (runtime.Result, error) {
	dir, err := a.freshSmokeDir("script")
	if err != nil {
		return runtime.Result{}, err
	}
	res, err := runtime.Run(a.ctx, script, dir)
	if err != nil {
		a.log.Error("smoke script failed", "err", err)
	}
	return res, err
}

// SmokeListTags clones url into a fresh <data>/smoke/git folder and returns
// its tags.
func (a *App) SmokeListTags(url string) ([]string, error) {
	dir, err := a.freshSmokeDir("git")
	if err != nil {
		return nil, err
	}
	tags, err := source.CloneTags(url, dir)
	if err != nil {
		a.log.Error("smoke clone failed", "url", url, "err", err)
	}
	return tags, err
}

// freshSmokeDir removes <data>/smoke/<name> and creates it empty.
func (a *App) freshSmokeDir(name string) (string, error) {
	dir := filepath.Join(a.dataDir, "smoke", name)
	if err := os.RemoveAll(dir); err != nil {
		a.log.Error("clear smoke folder", "dir", dir, "err", err)
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		a.log.Error("create smoke folder", "dir", dir, "err", err)
		return "", err
	}
	return dir, nil
}
