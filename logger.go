package main

import (
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

// wailsLogger adapts Wails' logger.Logger to slog so webview and runtime
// messages land in the app log.
type wailsLogger struct {
	log *slog.Logger
}

var _ logger.Logger = wailsLogger{}

func (l wailsLogger) Print(message string)   { l.log.Info(message, "source", "wails") }
func (l wailsLogger) Trace(message string)   { l.log.Debug(message, "source", "wails") }
func (l wailsLogger) Debug(message string)   { l.log.Debug(message, "source", "wails") }
func (l wailsLogger) Info(message string)    { l.log.Info(message, "source", "wails") }
func (l wailsLogger) Warning(message string) { l.log.Warn(message, "source", "wails") }
func (l wailsLogger) Error(message string)   { l.log.Error(message, "source", "wails") }
func (l wailsLogger) Fatal(message string)   { l.log.Error(message, "source", "wails", "fatal", true) }
