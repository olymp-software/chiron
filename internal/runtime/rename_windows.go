package runtime

import (
	"errors"

	"golang.org/x/sys/windows"
)

// crossDevice reports whether a rename failed because source and target are
// on different volumes.
func crossDevice(err error) bool {
	return errors.Is(err, windows.ERROR_NOT_SAME_DEVICE)
}
