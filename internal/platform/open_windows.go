package platform

import "os/exec"

// OpenFolder opens path in the file manager and returns without waiting.
// Explorer exits with code 1 even on success, so its result says nothing.
func OpenFolder(path string) error {
	cmd := exec.Command("explorer.exe", path)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
