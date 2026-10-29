//go:build unix

package runtime

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestStopBuiltinCopy(t *testing.T) {
	var out syncBuffer
	r := Start(helperSpec(t, t.TempDir(), `echo started; cat /dev/zero > /dev/null`), &out)
	eventually(t, "output", func() bool { return out.String() == "started\n" })
	start := time.Now()
	r.Stop()
	if res := wait(t, r); !res.Stopped {
		t.Errorf("result = %+v, want stopped", res)
	}
	if d := time.Since(start); d >= abandonDelay {
		t.Errorf("stopped after %v, want the copy to end before the abandon delay", d)
	}
}

// A built-in blocked where the stop cannot reach it is left behind.
func TestStopAbandonsBlockedBuiltin(t *testing.T) {
	defer func(d time.Duration) { abandonDelay = d }(abandonDelay)
	abandonDelay = 100 * time.Millisecond
	dir := t.TempDir()
	fifo := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	var out syncBuffer
	r := Start(helperSpec(t, dir, `echo started; cat fifo`), &out)
	// Opening the FIFO for writing releases the abandoned cat; without a
	// reader the open fails instead of blocking.
	defer func() {
		if f, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
			_ = f.Close()
		}
	}()
	eventually(t, "output", func() bool { return out.String() == "started\n" })
	r.Stop()
	if res := wait(t, r); !res.Stopped {
		t.Errorf("result = %+v, want stopped", res)
	}
}

func TestReleaseForgetsExitedGroups(t *testing.T) {
	var out syncBuffer
	r := Start(helperSpec(t, t.TempDir(), `"$HELPER" exit 0; "$HELPER" exit 0; "$HELPER" exit 0`), &out)
	wait(t, r)
	r.tree.mu.Lock()
	defer r.tree.mu.Unlock()
	if len(r.tree.pgids) != 0 {
		t.Errorf("pgids = %v, want none", r.tree.pgids)
	}
}
