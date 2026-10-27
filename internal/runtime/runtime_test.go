package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, script string) Result {
	t.Helper()
	res, err := Run(context.Background(), script, t.TempDir())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res
}

func TestRunStdout(t *testing.T) {
	res := run(t, `
mkdir sub
touch sub/a.txt
echo hello > sub/b.txt
cat sub/a.txt sub/b.txt
cd sub
cat b.txt
`)
	if want := "hello\nhello\n"; res.Stdout != want {
		t.Errorf("stdout = %q, want %q", res.Stdout, want)
	}
	if res.Stderr != "" {
		t.Errorf("stderr = %q, want empty", res.Stderr)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", res.ExitCode)
	}
}

func TestRunStderr(t *testing.T) {
	res := run(t, `echo x >&2`)
	if res.Stderr != "x\n" {
		t.Errorf("stderr = %q, want %q", res.Stderr, "x\n")
	}
	if res.Stdout != "" {
		t.Errorf("stdout = %q, want empty", res.Stdout)
	}
}

func TestRunExitCode(t *testing.T) {
	res := run(t, `exit 3`)
	if res.ExitCode != 3 {
		t.Errorf("exit code = %d, want 3", res.ExitCode)
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

func TestRunParseError(t *testing.T) {
	if _, err := Run(context.Background(), `if then`, t.TempDir()); err == nil {
		t.Error("Run: want a parse error")
	}
}
