package main

import (
	"context"
)

// App is the Wails-bound facade. Its methods delegate to internal packages.
type App struct {
	ctx context.Context
}

// NewApp creates the bound App.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts; the context is kept for runtime calls.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}
