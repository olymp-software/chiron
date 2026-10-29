package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runOK runs script in dir and fails the test unless it exits with code 0.
func runOK(t *testing.T, dir, script string) {
	t.Helper()
	if res := runIn(t, dir, script); res.ExitCode != 0 {
		t.Fatalf("exit code = %d, output = %q", res.ExitCode, res.Output)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(data) != want {
		t.Errorf("%s = %q, want %q", path, data, want)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Errorf("%s exists (err = %v), want it missing", path, err)
	}
}

func TestMkdirParents(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `mkdir -p a/b/c && mkdir -p a/b/c`)
	if info, err := os.Stat(filepath.Join(dir, "a", "b", "c")); err != nil || !info.IsDir() {
		t.Errorf("a/b/c: %v, want a directory", err)
	}
}

func TestMkdirWithoutParents(t *testing.T) {
	res := runIn(t, t.TempDir(), `mkdir a/b`)
	if res.ExitCode != 1 || !strings.HasPrefix(res.Output, "mkdir: a/b: no such file or directory") {
		t.Errorf("exit code = %d, stderr = %q", res.ExitCode, res.Output)
	}
}

// The scenario from the script-runtime spec.
func TestSameResultEveryOS(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `mkdir -p a/b && touch a/b/f && mkdir c && cp -r a c`)
	assertFile(t, filepath.Join(dir, "c", "a", "b", "f"), "")
}

func TestCpRecursiveIntoNewDir(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `mkdir -p a/b && echo x > a/b/f && cp -r a c`)
	assertFile(t, filepath.Join(dir, "c", "b", "f"), "x\n")
}

func TestCpFile(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `echo x > a && cp a b && echo y > c && cp a c`)
	assertFile(t, filepath.Join(dir, "b"), "x\n")
	assertFile(t, filepath.Join(dir, "c"), "x\n")
}

func TestCpDirWithoutRecursive(t *testing.T) {
	dir := t.TempDir()
	res := runIn(t, dir, `mkdir a && cp a b`)
	if res.ExitCode != 1 || res.Output == "" {
		t.Errorf("exit code = %d, stderr = %q", res.ExitCode, res.Output)
	}
	assertMissing(t, filepath.Join(dir, "b"))
}

func TestCpIntoItself(t *testing.T) {
	res := runIn(t, t.TempDir(), `mkdir a && cp -r a a`)
	if res.ExitCode != 1 || res.Output == "" {
		t.Errorf("exit code = %d, stderr = %q", res.ExitCode, res.Output)
	}
}

func TestRmRecursiveForceReadOnly(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `mkdir -p a/b && echo x > a/b/f`)
	if err := os.Chmod(filepath.Join(dir, "a", "b", "f"), 0o444); err != nil {
		t.Fatal(err)
	}
	runOK(t, dir, `rm -rf a`)
	assertMissing(t, filepath.Join(dir, "a"))
}

func TestRmForceMissing(t *testing.T) {
	runOK(t, t.TempDir(), `rm -f missing`)
}

func TestRmMissing(t *testing.T) {
	res := runIn(t, t.TempDir(), `rm missing`)
	if res.ExitCode != 1 || res.Output != "rm: missing: no such file or directory\n" {
		t.Errorf("exit code = %d, stderr = %q", res.ExitCode, res.Output)
	}
}

func TestRmDirWithoutRecursive(t *testing.T) {
	dir := t.TempDir()
	res := runIn(t, dir, `mkdir a && rm a`)
	if res.ExitCode != 1 || res.Output != "rm: a: is a directory\n" {
		t.Errorf("exit code = %d, stderr = %q", res.ExitCode, res.Output)
	}
}

func TestRmDot(t *testing.T) {
	dir := t.TempDir()
	res := runIn(t, dir, `touch f && rm -rf .`)
	if res.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", res.ExitCode)
	}
	assertFile(t, filepath.Join(dir, "f"), "")
}

func TestTouchCreates(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `touch a b`)
	assertFile(t, filepath.Join(dir, "a"), "")
	assertFile(t, filepath.Join(dir, "b"), "")
}

func TestTouchUpdatesTime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a")
	if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	runOK(t, dir, `touch a`)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().After(old.Add(time.Minute)) {
		t.Errorf("mod time = %v, want updated from %v", info.ModTime(), old)
	}
	assertFile(t, path, "x")
}

func TestMvRenames(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `echo x > a && mv a b && mkdir d && mv b d`)
	assertMissing(t, filepath.Join(dir, "a"))
	assertMissing(t, filepath.Join(dir, "b"))
	assertFile(t, filepath.Join(dir, "d", "b"), "x\n")
}

func TestUnsupportedFlag(t *testing.T) {
	for _, script := range []string{`cp -a a b`, `cp --recursive a b`, `mkdir -pv a`, `rm -i a`, `cat -n a`, `touch -c a`, `mv -f a b`} {
		t.Run(script, func(t *testing.T) {
			dir := t.TempDir()
			runOK(t, dir, `echo x > a`)
			res := runIn(t, dir, script)
			if res.ExitCode != 1 || !strings.Contains(res.Output, "unsupported option") {
				t.Errorf("exit code = %d, stderr = %q", res.ExitCode, res.Output)
			}
			assertFile(t, filepath.Join(dir, "a"), "x\n")
			assertMissing(t, filepath.Join(dir, "b"))
		})
	}
}

func TestDoubleDash(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `touch -- -x && cp -- -x y && cat -- -x && rm -- -x`)
	assertMissing(t, filepath.Join(dir, "-x"))
	assertFile(t, filepath.Join(dir, "y"), "")
}

func TestCatStdin(t *testing.T) {
	res := runIn(t, t.TempDir(), `echo x | cat`)
	if res.Output != "x\n" {
		t.Errorf("output = %q, want %q", res.Output, "x\n")
	}
}

func TestPathsRelativeToScriptDir(t *testing.T) {
	dir := t.TempDir()
	runOK(t, dir, `mkdir sub && cd sub && touch f && mkdir -p g && cp f g && mv f h`)
	assertFile(t, filepath.Join(dir, "sub", "g", "f"), "")
	assertFile(t, filepath.Join(dir, "sub", "h"), "")
}
