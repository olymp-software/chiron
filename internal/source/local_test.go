package source

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
)

// realTempDir returns a temp folder with symbolic links in its path resolved
// (macOS /var is a link to /private/var).
func realTempDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func wantCode(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("err = %v, want apperr %s", err, code)
	}
	if ae.Code != code {
		t.Fatalf("code = %s, want %s", ae.Code, code)
	}
}

func TestNormalizeLocal(t *testing.T) {
	root := realTempDir(t)
	course := filepath.Join(root, "go-basics")
	mkdir(t, filepath.Join(course, "basics"))

	tests := []struct {
		name string
		in   string
	}{
		{name: "clean", in: course},
		{name: "trailing separator", in: course + string(filepath.Separator)},
		{name: "dot dot", in: filepath.Join(course, "basics") + string(filepath.Separator) + ".."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeLocal(tt.in)
			if err != nil {
				t.Fatalf("NormalizeLocal: %v", err)
			}
			if got != course {
				t.Errorf("NormalizeLocal(%q) = %q, want %q", tt.in, got, course)
			}
		})
	}
}

func TestNormalizeLocalSymlink(t *testing.T) {
	root := realTempDir(t)
	course := filepath.Join(root, "go-basics")
	mkdir(t, course)
	link := filepath.Join(root, "link-to-go")
	if err := os.Symlink(course, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	got, err := NormalizeLocal(link)
	if err != nil {
		t.Fatalf("NormalizeLocal: %v", err)
	}
	if got != course {
		t.Errorf("NormalizeLocal(link) = %q, want %q", got, course)
	}
}

func TestNormalizeLocalErrors(t *testing.T) {
	root := realTempDir(t)
	file := filepath.Join(root, "course.yaml")
	writeFile(t, file, "")

	_, err := NormalizeLocal(file)
	wantCode(t, err, apperr.NotAFolder)

	_, err = NormalizeLocal(filepath.Join(root, "missing"))
	wantCode(t, err, apperr.FolderNotFound)
}
