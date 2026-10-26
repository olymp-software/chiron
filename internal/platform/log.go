package platform

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	logName = "chiron.log"
	// maxLogSize is the size above which the log is rolled over to chiron.log.1 on open.
	maxLogSize = 10 << 20
)

// OpenLog opens <dir>/chiron.log for appending and returns a text logger
// writing to it. A log larger than 10 MB is first renamed to chiron.log.1,
// replacing the previous one. The caller closes the returned file.
func OpenLog(dir string) (*slog.Logger, *os.File, error) {
	path := filepath.Join(dir, logName)
	if err := rollOver(path); err != nil {
		return nil, nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	return slog.New(slog.NewTextHandler(f, nil)), f, nil
}

func rollOver(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Size() <= maxLogSize {
		return nil
	}
	return os.Rename(path, path+".1")
}
