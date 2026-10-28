package source

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// tree returns the slash-separated paths of all files under root, sorted.
func tree(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(files)
	return files
}

func copyTree(t *testing.T, src string) []string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "copy")
	if _, err := CopyLocal(src, dst); err != nil {
		t.Fatalf("CopyLocal: %v", err)
	}
	return tree(t, dst)
}

func TestCopyLocalPlainFolder(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "course.yaml"), "title: X\n")
	writeFile(t, filepath.Join(src, "basics", "hello", "activity.yaml"), "version: 1\n")
	mkdir(t, filepath.Join(src, "empty"))

	got := copyTree(t, src)

	want := []string{"basics/hello/activity.yaml", "course.yaml"}
	if !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
}

func TestCopyLocalContent(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "course.yaml"), "title: Go Basics\n")
	dst := filepath.Join(t.TempDir(), "copy")

	if _, err := CopyLocal(src, dst); err != nil {
		t.Fatalf("CopyLocal: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dst, "course.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "title: Go Basics\n" {
		t.Errorf("course.yaml = %q, want the source content", b)
	}
}

func TestCopyLocalIgnoredFolder(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, ".gitignore"), "bin/\n")
	writeFile(t, filepath.Join(src, "basics", "hello", "bin", "hello.exe"), "")
	writeFile(t, filepath.Join(src, "basics", "hello", "main.go"), "")

	got := copyTree(t, src)

	want := []string{".gitignore", "basics/hello/main.go"}
	if !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
}

func TestCopyLocalNestedNegation(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "basics", ".gitignore"), "*.log\n!keep.log\n")
	writeFile(t, filepath.Join(src, "basics", "keep.log"), "")
	writeFile(t, filepath.Join(src, "basics", "run.log"), "")
	writeFile(t, filepath.Join(src, "basics", "hello", "build.log"), "")
	writeFile(t, filepath.Join(src, "root.log"), "")

	got := copyTree(t, src)

	// The nested .gitignore applies only inside basics/.
	want := []string{"basics/.gitignore", "basics/keep.log", "root.log"}
	if !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
}

func TestCopyLocalCannotReincludeInIgnoredFolder(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, ".gitignore"), "out/\n!out/keep.txt\n")
	writeFile(t, filepath.Join(src, "out", "keep.txt"), "")

	got := copyTree(t, src)

	// As in git, a file cannot be re-included when its folder is excluded.
	want := []string{".gitignore"}
	if !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
}

func TestCopyLocalSkipsGitAndInfoExclude(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, ".git", "HEAD"), "ref: refs/heads/main\n")
	writeFile(t, filepath.Join(src, ".git", "info", "exclude"), "*.md\n")
	writeFile(t, filepath.Join(src, "basics", ".git"), "gitdir: ../.git/modules/basics\n")
	writeFile(t, filepath.Join(src, "OVERVIEW.md"), "")

	got := copyTree(t, src)

	want := []string{"OVERVIEW.md"}
	if !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
}

func TestCopyLocalIgnoresGlobalPatterns(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "global-ignore"), "*.md\n")
	writeFile(t, filepath.Join(home, ".gitconfig"),
		"[core]\n\texcludesfile = "+filepath.ToSlash(filepath.Join(home, "global-ignore"))+"\n")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "OVERVIEW.md"), "")

	got := copyTree(t, src)

	if want := []string{"OVERVIEW.md"}; !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
}

func TestCopyLocalSkipsSymlinks(t *testing.T) {
	outside := t.TempDir()
	writeFile(t, filepath.Join(outside, "secret.txt"), "")
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "course.yaml"), "")
	if err := os.Symlink(outside, filepath.Join(src, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "copy")

	skipped, err := CopyLocal(src, dst)
	if err != nil {
		t.Fatalf("CopyLocal: %v", err)
	}

	if got, want := tree(t, dst), []string{"course.yaml"}; !slices.Equal(got, want) {
		t.Errorf("copy = %v, want %v", got, want)
	}
	if want := []string{"linked"}; !slices.Equal(skipped, want) {
		t.Errorf("skipped = %v, want %v", skipped, want)
	}
}

func TestCopyLocalMissingSource(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "copy")

	if _, err := CopyLocal(filepath.Join(t.TempDir(), "missing"), dst); err == nil {
		t.Fatal("CopyLocal of a missing folder succeeded, want an error")
	}
}
