package main

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/olymp-software/chiron/internal/library"
	"github.com/olymp-software/chiron/internal/platform"
	"github.com/olymp-software/chiron/internal/settings"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dataDir, log, closeLog, err := setupLog()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	defer closeLog()
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic", "value", r, "stack", string(debug.Stack()))
			panic(r)
		}
	}()

	log.Info("starting")

	lib, err := library.Open(dataDir, log)
	if err != nil {
		log.Error("open library", "err", err)
		os.Exit(1)
	}
	app := NewApp(lib, settings.Open(dataDir, log), log)

	err = wails.Run(&options.App{
		Title:  "Chiron",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
			// Requests not found in the embedded assets, i.e. course files.
			Handler: lib.FilesHandler(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		Logger:           wailsLogger{log: log},
		OnStartup:        app.startup,
		Bind: []any{
			app,
		},
	})
	if err != nil {
		log.Error("wails run failed", "err", err)
	}
	log.Info("exiting")
}

// setupLog creates the data and log folders and opens the app log. It returns
// the data folder.
func setupLog() (string, *slog.Logger, func(), error) {
	dataDir, err := platform.DataDir()
	if err != nil {
		return "", nil, nil, err
	}
	logDir, err := platform.LogDir()
	if err != nil {
		return "", nil, nil, err
	}
	for _, dir := range []string{dataDir, logDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", nil, nil, err
		}
	}
	log, f, err := platform.OpenLog(logDir)
	if err != nil {
		return "", nil, nil, err
	}
	return dataDir, log, func() { _ = f.Close() }, nil
}
