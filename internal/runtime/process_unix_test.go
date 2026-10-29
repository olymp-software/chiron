//go:build unix

package runtime

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

// processAlive reports whether pid is a running process; a zombie waiting to
// be reaped counts as dead.
func processAlive(pid int) bool {
	if syscall.Kill(pid, 0) != nil {
		return false
	}
	stat, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return true
	}
	// The state follows the command name in parentheses.
	s := string(stat)
	i := strings.LastIndexByte(s, ')')
	return i < 0 || i+2 >= len(s) || s[i+2] != 'Z'
}
