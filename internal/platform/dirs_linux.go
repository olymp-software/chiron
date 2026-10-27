package platform

import (
	"os"
	"path/filepath"
)

// DataDir returns $XDG_DATA_HOME/chiron, defaulting to ~/.local/share/chiron.
func DataDir() (string, error) {
	return xdgDir("XDG_DATA_HOME", filepath.Join(".local", "share"))
}

// LogDir returns $XDG_STATE_HOME/chiron, defaulting to ~/.local/state/chiron.
// The XDG Base Directory spec names logs as state data.
func LogDir() (string, error) {
	return xdgDir("XDG_STATE_HOME", filepath.Join(".local", "state"))
}

// xdgDir resolves an XDG base directory by hand: an unset or relative value
// falls back to defaultUnderHome, as the spec requires.
func xdgDir(env, defaultUnderHome string) (string, error) {
	if base := os.Getenv(env); filepath.IsAbs(base) {
		return filepath.Join(base, "chiron"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultUnderHome, "chiron"), nil
}
