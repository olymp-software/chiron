package platform

import (
	"os"
	"path/filepath"
)

// DataDir returns ~/Library/Application Support/Chiron.
func DataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "Chiron"), nil
}

// LogDir returns ~/Library/Logs/Chiron, where Console.app looks for app logs.
func LogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", "Chiron"), nil
}
