package platform

import "os/exec"

// OpenFolder opens path in Finder.
func OpenFolder(path string) error {
	return exec.Command("open", path).Run()
}
