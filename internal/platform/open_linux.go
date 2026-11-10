package platform

import "os/exec"

// OpenFolder opens path in the file manager.
func OpenFolder(path string) error {
	return exec.Command("xdg-open", path).Run()
}
