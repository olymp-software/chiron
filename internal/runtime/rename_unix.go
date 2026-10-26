//go:build !windows

package runtime

import (
	"errors"
	"syscall"
)

// crossDevice reports whether a rename failed because source and target are
// on different file systems.
func crossDevice(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}
