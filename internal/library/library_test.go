package library

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/source"
)

func courseYAML(title string) string {
	return "version: 1\ntitle: " + title + "\nsections:\n" +
		"  - {slug: basics, title: Basics, activities: [{slug: hello, title: Hello}]}\n"
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// courseFolder creates an author's course folder with a valid course.yaml.
func courseFolder(t *testing.T, title string) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(dir, "go-basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), courseYAML(title))
	return src
}

func openLib(t *testing.T, dataDir string) *Library {
	t.Helper()
	l, err := Open(dataDir, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return l
}

func newLib(t *testing.T) (*Library, string) {
	t.Helper()
	dataDir := t.TempDir()
	return openLib(t, dataDir), dataDir
}

func wantCode(t *testing.T, err error, code apperr.Code) {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("err = %v, want apperr %s", err, code)
	}
	if ae.Code != code {
		t.Fatalf("code = %s, want %s (%v)", ae.Code, code, ae)
	}
}

func addLocal(t *testing.T, l *Library, path string) Entry {
	t.Helper()
	e, err := l.AddLocal(path)
	if err != nil {
		t.Fatalf("AddLocal: %v", err)
	}
	return e
}

// courseEntries returns the names in the courses folder.
func courseEntries(t *testing.T, dataDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dataDir, coursesDir))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// 5.1 library.json

func TestListEmptyWithoutFile(t *testing.T) {
	l, _ := newLib(t)

	courses, err := l.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(courses) != 0 {
		t.Errorf("courses = %v, want none", courses)
	}
}

func TestLibraryRoundTrip(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)

	courses, err := openLib(t, dataDir).List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	want := []Entry{{Origin: OriginLocal, Source: src, Key: source.Key(src), Title: "Go Basics"}}
	if !reflect.DeepEqual(courses, want) || e != want[0] {
		t.Errorf("courses = %+v, entry = %+v, want %+v", courses, e, want)
	}
	file := readFile(t, filepath.Join(dataDir, fileName))
	if !strings.Contains(file, `"version": 1`) || !strings.Contains(file, `"origin": "local"`) {
		t.Errorf("library.json = %s, want version and origin", file)
	}
}

func TestNewerLibraryFormatIsNotWritten(t *testing.T) {
	dataDir := t.TempDir()
	newer := `{"version": 2, "courses": [], "labs": [{"source": "x"}]}`
	writeFile(t, filepath.Join(dataDir, fileName), newer)
	l := openLib(t, dataDir)

	_, err := l.List()
	wantCode(t, err, apperr.LibraryNewerFormat)
	_, err = l.AddLocal(courseFolder(t, "Go Basics"))
	wantCode(t, err, apperr.LibraryNewerFormat)

	if got := readFile(t, filepath.Join(dataDir, fileName)); got != newer {
		t.Errorf("library.json = %s, want unchanged", got)
	}
}

// 5.2 AddLocal

func TestAddLocalCopiesCourse(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, ".gitignore"), "bin/\n")
	writeFile(t, filepath.Join(src, "basics", "hello", "bin", "hello.exe"), "")
	writeFile(t, filepath.Join(src, "OVERVIEW.md"), "# Go")
	l, dataDir := newLib(t)

	e := addLocal(t, l, src)

	cp := l.CopyPath(e.Key)
	if got := readFile(t, filepath.Join(cp, "OVERVIEW.md")); got != "# Go" {
		t.Errorf("OVERVIEW.md in copy = %q", got)
	}
	if exists(filepath.Join(cp, "basics", "hello", "bin")) {
		t.Error("ignored bin/ was copied")
	}
	if names := courseEntries(t, dataDir); len(names) != 1 || names[0] != e.Key {
		t.Errorf("courses folder = %v, want only %s", names, e.Key)
	}
}

func TestAddLocalInvalidManifest(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		code apperr.Code
	}{
		{name: "no course.yaml", code: apperr.ManifestMissing},
		{name: "invalid course.yaml", yaml: "version: 1\ntitle: X\n", code: apperr.ManifestInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := t.TempDir()
			writeFile(t, filepath.Join(src, "OVERVIEW.md"), "")
			if tt.yaml != "" {
				writeFile(t, filepath.Join(src, manifest.CourseFile), tt.yaml)
			}
			l, dataDir := newLib(t)

			_, err := l.AddLocal(src)

			wantCode(t, err, tt.code)
			if courses, _ := l.List(); len(courses) != 0 {
				t.Errorf("courses = %v, want none", courses)
			}
			if names := courseEntries(t, dataDir); len(names) != 0 {
				t.Errorf("courses folder = %v, want empty", names)
			}
		})
	}
}

func TestAddLocalMissingFolder(t *testing.T) {
	l, _ := newLib(t)

	_, err := l.AddLocal(filepath.Join(t.TempDir(), "missing"))

	wantCode(t, err, apperr.FolderNotFound)
}

func TestAddLocalSameFolderTwice(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	first := addLocal(t, l, src)
	// The author edits the folder; adding again must not refresh the copy.
	writeFile(t, filepath.Join(src, manifest.CourseFile), courseYAML("Renamed"))

	for _, spelling := range []string{
		src + string(filepath.Separator),
		filepath.Join(src, "basics") + string(filepath.Separator) + "..",
	} {
		again := addLocal(t, l, spelling)
		if again != first {
			t.Errorf("AddLocal(%q) = %+v, want the existing %+v", spelling, again, first)
		}
	}
	if courses, _ := l.List(); len(courses) != 1 {
		t.Errorf("courses = %v, want one", courses)
	}
}

func TestAddLocalStorageKeyTaken(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	dataDir := t.TempDir()
	writeFile(t, filepath.Join(dataDir, fileName), `{"version": 1, "courses": [`+
		`{"origin": "local", "source": "/elsewhere", "key": "`+source.Key(src)+`", "title": "Other"}]}`)
	l := openLib(t, dataDir)

	_, err := l.AddLocal(src)

	wantCode(t, err, apperr.StorageKeyTaken)
	if names := courseEntries(t, dataDir); len(names) != 0 {
		t.Errorf("courses folder = %v, want empty", names)
	}
}

// 5.3 Sync

func TestSyncShowsEdits(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, "OVERVIEW.md"), "old")
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	writeFile(t, filepath.Join(src, "OVERVIEW.md"), "new")
	writeFile(t, filepath.Join(src, manifest.CourseFile), courseYAML("Go Basics 2"))

	synced, err := l.Sync(e.Key)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if got := readFile(t, filepath.Join(l.CopyPath(e.Key), "OVERVIEW.md")); got != "new" {
		t.Errorf("OVERVIEW.md in copy = %q, want new", got)
	}
	if synced.Title != "Go Basics 2" {
		t.Errorf("synced title = %q, want the new title", synced.Title)
	}
	if courses, _ := l.List(); courses[0].Title != "Go Basics 2" {
		t.Errorf("stored title = %q, want the new title", courses[0].Title)
	}
	if names := courseEntries(t, dataDir); len(names) != 1 {
		t.Errorf("courses folder = %v, want only the copy", names)
	}
}

func TestSyncBrokenManifestKeepsTitle(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	broken := "version: 1\ntitle: Broken\n"
	writeFile(t, filepath.Join(src, manifest.CourseFile), broken)

	synced, err := l.Sync(e.Key)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if got := readFile(t, filepath.Join(l.CopyPath(e.Key), manifest.CourseFile)); got != broken {
		t.Errorf("course.yaml in copy = %q, want the broken one", got)
	}
	if synced.Title != "Go Basics" {
		t.Errorf("title = %q, want the last valid title", synced.Title)
	}
}

func TestSyncFolderGone(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	if err := os.RemoveAll(src); err != nil {
		t.Fatal(err)
	}

	_, err := l.Sync(e.Key)

	wantCode(t, err, apperr.FolderNotFound)
	if got := readFile(t, filepath.Join(l.CopyPath(e.Key), manifest.CourseFile)); got != courseYAML("Go Basics") {
		t.Errorf("course.yaml in copy = %q, want unchanged", got)
	}
}

func TestSyncDropsFilesWrittenIntoCopy(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	written := filepath.Join(l.CopyPath(e.Key), "basics", "hello", "state.txt")
	writeFile(t, written, "")

	if _, err := l.Sync(e.Key); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if exists(written) {
		t.Error("file written into the copy survived Sync")
	}
	if exists(filepath.Join(src, "basics", "hello", "state.txt")) {
		t.Error("author's folder was changed")
	}
}

func TestSyncUnknownCourse(t *testing.T) {
	l, _ := newLib(t)

	_, err := l.Sync("not-added-0000")

	wantCode(t, err, apperr.CourseNotFound)
}

// 5.4 recovery

func TestRecoverInterruptedSwap(t *testing.T) {
	tests := []struct {
		name string
		// arrange moves the copy at cur into the state left by an interruption.
		arrange func(t *testing.T, cur, coursesPath, key string)
		want    string
	}{
		{
			name: "interrupted while copying",
			arrange: func(t *testing.T, cur, coursesPath, key string) {
				writeFile(t, filepath.Join(coursesPath, key+newSuffix+"2", manifest.CourseFile), "partial")
			},
			want: "old",
		},
		{
			name: "interrupted after moving the copy aside",
			arrange: func(t *testing.T, cur, coursesPath, key string) {
				if err := os.Rename(cur, filepath.Join(coursesPath, key+oldSuffix+"1")); err != nil {
					t.Fatal(err)
				}
				writeFile(t, filepath.Join(coursesPath, key+newSuffix+"2", manifest.CourseFile), "new")
			},
			want: "old",
		},
		{
			name: "interrupted before deleting the old copy",
			arrange: func(t *testing.T, cur, coursesPath, key string) {
				if err := os.Rename(cur, filepath.Join(coursesPath, key+oldSuffix+"1")); err != nil {
					t.Fatal(err)
				}
				writeFile(t, filepath.Join(cur, manifest.CourseFile), "new")
			},
			want: "new",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, dataDir := newLib(t)
			e := addLocal(t, l, courseFolder(t, "Go Basics"))
			cur := l.CopyPath(e.Key)
			writeFile(t, filepath.Join(cur, manifest.CourseFile), "old")
			tt.arrange(t, cur, filepath.Join(dataDir, coursesDir), e.Key)

			openLib(t, dataDir)

			if got := readFile(t, filepath.Join(cur, manifest.CourseFile)); got != tt.want {
				t.Errorf("course.yaml = %q, want %q", got, tt.want)
			}
			if names := courseEntries(t, dataDir); len(names) != 1 {
				t.Errorf("courses folder = %v, want only the copy", names)
			}
		})
	}
}

func TestRecoverPicksNewestOldCopy(t *testing.T) {
	l, dataDir := newLib(t)
	e := addLocal(t, l, courseFolder(t, "Go Basics"))
	coursesPath := filepath.Join(dataDir, coursesDir)
	if err := os.RemoveAll(l.CopyPath(e.Key)); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(coursesPath, e.Key+oldSuffix+"5", manifest.CourseFile), "older")
	writeFile(t, filepath.Join(coursesPath, e.Key+oldSuffix+"10", manifest.CourseFile), "newest")

	openLib(t, dataDir)

	if got := readFile(t, filepath.Join(l.CopyPath(e.Key), manifest.CourseFile)); got != "newest" {
		t.Errorf("course.yaml = %q, want the newest old copy", got)
	}
}

func TestRecoverDeletesSmokeFolder(t *testing.T) {
	dataDir := t.TempDir()
	writeFile(t, filepath.Join(dataDir, "smoke", "git", "x"), "")

	openLib(t, dataDir)

	if exists(filepath.Join(dataDir, "smoke")) {
		t.Error("smoke folder survived")
	}
}

// 5.5 List and Overview

func TestListDoesNotReadCourses(t *testing.T) {
	l, _ := newLib(t)
	e := addLocal(t, l, courseFolder(t, "Go Basics"))
	if err := os.RemoveAll(l.CopyPath(e.Key)); err != nil {
		t.Fatal(err)
	}

	courses, err := l.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != "Go Basics" {
		t.Errorf("courses = %+v, want the stored title", courses)
	}
}

func TestOverview(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, "overview.md"), "# Go")
	l, _ := newLib(t)
	e := addLocal(t, l, src)

	o, err := l.Overview(e.Key)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}

	want := Overview{Entry: e, Text: manifest.OverviewText{Markdown: "# Go"}}
	if !reflect.DeepEqual(o, want) {
		t.Errorf("overview = %+v, want %+v", o, want)
	}
}

func TestOverviewBrokenManifest(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	writeFile(t, filepath.Join(src, manifest.CourseFile), "version: 1\ntitle: Broken\n")
	if _, err := l.Sync(e.Key); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	o, err := l.Overview(e.Key)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}

	if o.Title != "Go Basics" || o.Error == nil || o.Error.Code != apperr.ManifestInvalid {
		t.Errorf("overview = %+v, want the stored title and %s", o, apperr.ManifestInvalid)
	}
	if o.Text != (manifest.OverviewText{}) {
		t.Errorf("text = %+v, want none", o.Text)
	}
}
