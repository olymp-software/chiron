package platform

import (
	"os"
	"path/filepath"
)

// LearnerRoot returns ~/Chiron, the root of the folders the learner works in.
func LearnerRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Chiron"), nil
}
