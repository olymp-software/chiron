package platform

import (
	"errors"
	"os"
	"path/filepath"
)

// DataDir returns %LOCALAPPDATA%\Chiron.
func DataDir() (string, error) {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return "", errors.New("LOCALAPPDATA is not set")
	}
	return filepath.Join(local, "Chiron"), nil
}

// LogDir returns %LOCALAPPDATA%\Chiron\logs.
func LogDir() (string, error) {
	data, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(data, "logs"), nil
}
