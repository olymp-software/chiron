package runtime

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// The test binary doubles as the external commands the scripts start: with
// helperEnv set it runs the helper mode named by its first argument.
const helperEnv = "RUNTIME_TEST_HELPER"

func TestMain(m *testing.M) {
	if os.Getenv(helperEnv) != "" {
		helper(os.Args[1], os.Args[2:])
		return
	}
	os.Exit(m.Run())
}

func helper(mode string, args []string) {
	switch mode {
	case "child":
		writePid(filepath.Join(args[0], "child.pid"))
		gc := exec.Command(os.Args[0], "grandchild", args[0])
		if err := gc.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		time.Sleep(10 * time.Minute)
	case "grandchild":
		writePid(filepath.Join(args[0], "grandchild.pid"))
		time.Sleep(10 * time.Minute)
	case "waitfile":
		for {
			if _, err := os.Stat(args[0]); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	case "interleave":
		_, _ = os.Stdout.WriteString("1\n")
		_, _ = os.Stderr.WriteString("2\n")
		_, _ = os.Stdout.WriteString("3\n")
	case "env":
		fmt.Println(os.Getenv("HOST_OS"), os.Getenv("HOST_ARCH"), os.Getenv("FOO"))
	case "stdin":
		n, _ := io.Copy(io.Discard, os.Stdin)
		fmt.Println(n)
	case "exit":
		code, _ := strconv.Atoi(args[0])
		os.Exit(code)
	}
}

// writePid writes the process id to path atomically, so a reader never sees
// a partial file.
func writePid(path string) {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		panic(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		panic(err)
	}
}

// syncBuffer is an output writer that the test reads while the run writes.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func helperSpec(t *testing.T, dir, script string, env ...string) Spec {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	env = append(append(os.Environ(), helperEnv+"=1", "HELPER="+exe), env...)
	return Spec{Script: []byte(script), Name: "script.sh", Dir: dir, Env: env}
}

// ran is the outcome of a script run to the end.
type ran struct {
	Output   string
	ExitCode int
}

// runIn runs script in dir to the end.
func runIn(t *testing.T, dir, script string, env ...string) ran {
	t.Helper()
	var out syncBuffer
	res := wait(t, Start(helperSpec(t, dir, script, env...), &out))
	if res.Stopped {
		t.Fatal("run reported as stopped")
	}
	return ran{Output: out.String(), ExitCode: res.ExitCode}
}

// wait waits for r to exit, failing the test if it takes too long.
func wait(t *testing.T, r *Run) Result {
	t.Helper()
	ch := make(chan Result, 1)
	go func() { ch <- r.Wait() }()
	select {
	case res := <-ch:
		return res
	case <-time.After(30 * time.Second):
		r.Stop()
		t.Fatal("run did not exit")
		return Result{}
	}
}

// eventually fails the test unless cond becomes true within 10 seconds.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func readPid(t *testing.T, path string) int {
	t.Helper()
	var pid int
	eventually(t, path, func() bool {
		data, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		pid, err = strconv.Atoi(string(data))
		return err == nil
	})
	return pid
}

func TestRunOutput(t *testing.T) {
	res := runIn(t, t.TempDir(), `
mkdir sub
touch sub/a.txt
echo hello > sub/b.txt
cat sub/a.txt sub/b.txt
cd sub
cat b.txt
`)
	if want := "hello\nhello\n"; res.Output != want {
		t.Errorf("output = %q, want %q", res.Output, want)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", res.ExitCode)
	}
}

func TestRunInterleavedOutput(t *testing.T) {
	res := runIn(t, t.TempDir(), `echo a; echo b >&2; echo c; "$HELPER" interleave`)
	if want := "a\nb\nc\n1\n2\n3\n"; res.Output != want {
		t.Errorf("output = %q, want %q", res.Output, want)
	}
}

func TestRunEnv(t *testing.T) {
	res := runIn(t, t.TempDir(), `echo "$HOST_OS $HOST_ARCH $FOO"; "$HELPER" env`, "FOO=bar")
	line := goruntime.GOOS + " " + goruntime.GOARCH + " bar\n"
	if want := line + line; res.Output != want {
		t.Errorf("output = %q, want %q", res.Output, want)
	}
}

func TestRunEmptyStdin(t *testing.T) {
	res := runIn(t, t.TempDir(), `cat; "$HELPER" stdin; read x; echo "read $?"`)
	if want := "0\nread 1\n"; res.Output != want {
		t.Errorf("output = %q, want %q", res.Output, want)
	}
}

func TestRunParseError(t *testing.T) {
	res := runIn(t, t.TempDir(), `if then`)
	if res.ExitCode != 2 {
		t.Errorf("exit code = %d, want 2", res.ExitCode)
	}
	if !strings.HasPrefix(res.Output, "script.sh:1:") {
		t.Errorf("output = %q, want a parse error at script.sh:1", res.Output)
	}
}

func TestRunExitCodes(t *testing.T) {
	for script, want := range map[string]int{
		`exit 0`:                    0,
		`exit 1`:                    1,
		`exit 3`:                    3,
		`no-such-command-chiron`:    127,
		`"$HELPER" exit 0`:          0,
		`"$HELPER" exit 5`:          5,
		`"$HELPER" exit 1; exit $?`: 1,
	} {
		t.Run(script, func(t *testing.T) {
			if res := runIn(t, t.TempDir(), script); res.ExitCode != want {
				t.Errorf("exit code = %d, want %d (output %q)", res.ExitCode, want, res.Output)
			}
		})
	}
}

// The baseline's moreinterp/coreutils is replaced by the app's own utilities.
func TestNoMoreinterp(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "mvdan.cc/sh/moreinterp") {
		t.Error("go.mod requires mvdan.cc/sh/moreinterp")
	}
}

func TestStopBuiltinLoop(t *testing.T) {
	var out syncBuffer
	r := Start(helperSpec(t, t.TempDir(), `echo started; while true; do :; done`), &out)
	eventually(t, "output", func() bool { return out.String() == "started\n" })
	r.Stop()
	r.Stop()
	if res := wait(t, r); !res.Stopped {
		t.Errorf("result = %+v, want stopped", res)
	}
	r.Stop()
}

func TestStopAfterExit(t *testing.T) {
	var out syncBuffer
	r := Start(helperSpec(t, t.TempDir(), `exit 4`), &out)
	res := wait(t, r)
	r.Stop()
	if res.ExitCode != 4 || res.Stopped {
		t.Errorf("result = %+v, want exit code 4, not stopped", res)
	}
	if again := r.Wait(); again != res {
		t.Errorf("Wait after Stop = %+v, want %+v", again, res)
	}
}

// The "Stop a server" scenario: a background child with a grandchild of its
// own, both killed by Stop.
func TestStopKillsProcessTree(t *testing.T) {
	dir := t.TempDir()
	var out syncBuffer
	r := Start(helperSpec(t, dir, `"$HELPER" child "$PWD" & wait`), &out)
	child := readPid(t, filepath.Join(dir, "child.pid"))
	grandchild := readPid(t, filepath.Join(dir, "grandchild.pid"))
	r.Stop()
	if res := wait(t, r); !res.Stopped {
		t.Errorf("result = %+v, want stopped", res)
	}
	eventually(t, "child to die", func() bool { return !processAlive(child) })
	eventually(t, "grandchild to die", func() bool { return !processAlive(grandchild) })
}

func TestExitKillsLeftProcesses(t *testing.T) {
	dir := t.TempDir()
	res := runIn(t, dir, `"$HELPER" child "$PWD" &
"$HELPER" waitfile "$PWD/grandchild.pid"`)
	if res.ExitCode != 0 {
		t.Fatalf("exit code = %d, output = %q", res.ExitCode, res.Output)
	}
	child := readPid(t, filepath.Join(dir, "child.pid"))
	grandchild := readPid(t, filepath.Join(dir, "grandchild.pid"))
	eventually(t, "child to die", func() bool { return !processAlive(child) })
	eventually(t, "grandchild to die", func() bool { return !processAlive(grandchild) })
}
