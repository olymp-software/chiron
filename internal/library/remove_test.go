package library

import (
	"path/filepath"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
)

// removedWorkshop is an enrolled workshop with a file of the learner in its
// workshop folder.
func removedWorkshop(t *testing.T) (*Library, string, string) {
	t.Helper()
	l, dataDir, key := enrolledWorkshop(t, "", "", "", nil)
	folder, err := l.EnsureWorkshopFolder(key)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(folder, "mine"), "mine")
	return l, dataDir, key
}

func remove(t *testing.T, l *Library, key string, keepProgress, keepFiles bool) {
	t.Helper()
	left, err := l.Remove(key, keepProgress, keepFiles)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(left) != 0 {
		t.Errorf("left = %v, want nothing", left)
	}
}

func wantNotInLibrary(t *testing.T, l *Library, key string) {
	t.Helper()
	courses, err := l.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range courses {
		if c.Key == key {
			t.Errorf("%s is still in the library", key)
		}
	}
}

func TestRemoveEverything(t *testing.T) {
	l, dataDir, key := removedWorkshop(t)
	other := addLocal(t, l, courseFolder(t, "Other"))

	remove(t, l, key, false, false)

	wantNotInLibrary(t, l, key)
	for _, path := range []string{
		l.CopyPath(key), statePath(dataDir, key), l.WorkshopFolder(key), filepath.Join(dataDir, tempDir, key),
	} {
		if exists(path) {
			t.Errorf("%s survived", path)
		}
	}
	if courses, _ := l.List(); len(courses) != 1 || courses[0] != other {
		t.Errorf("courses = %+v, want only %+v", courses, other)
	}
	if !exists(l.CopyPath(other.Key)) {
		t.Error("copy of another course deleted")
	}
}

func TestRemoveKeepProgress(t *testing.T) {
	l, dataDir, key := removedWorkshop(t)
	state := readFile(t, statePath(dataDir, key))

	remove(t, l, key, true, false)

	wantNotInLibrary(t, l, key)
	if got := readFile(t, statePath(dataDir, key)); got != state {
		t.Errorf("progress = %s, want unchanged %s", got, state)
	}
	if exists(l.WorkshopFolder(key)) || exists(l.CopyPath(key)) {
		t.Error("workshop folder or course copy survived")
	}
}

func TestRemoveKeepFiles(t *testing.T) {
	l, dataDir, key := removedWorkshop(t)

	remove(t, l, key, false, true)

	wantNotInLibrary(t, l, key)
	if got := readFile(t, filepath.Join(l.WorkshopFolder(key), "mine")); got != "mine" {
		t.Errorf("file in the workshop folder = %q, want unchanged", got)
	}
	if exists(statePath(dataDir, key)) || exists(l.CopyPath(key)) {
		t.Error("progress or course copy survived")
	}
}

func TestRemoveUnknownKey(t *testing.T) {
	l, dataDir, key := removedWorkshop(t)
	before := readFile(t, filepath.Join(dataDir, fileName))

	_, err := l.Remove("not-added-0000", false, false)

	wantCode(t, err, apperr.CourseNotFound)
	if got := readFile(t, filepath.Join(dataDir, fileName)); got != before {
		t.Errorf("library.json = %s, want unchanged", got)
	}
	if !exists(l.CopyPath(key)) || !exists(statePath(dataDir, key)) || !exists(l.WorkshopFolder(key)) {
		t.Error("files of another library item deleted")
	}
}

func TestRemoveForgetsCheck(t *testing.T) {
	src := checkCourse(t, "exit 0\n")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	runCheck(t, l, e.Key)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	remove(t, l, e.Key, true, true)
	again := addLocal(t, l, src)

	if again.CheckHash != "" {
		t.Errorf("check hash = %q, want none", again.CheckHash)
	}
	wantStatus(t, l, again.Key, CheckNotChecked)
	_, err := l.Enter(again.Key)
	wantCode(t, err, apperr.CheckRequired)
}

func TestRemoveStopsWorkshopControl(t *testing.T) {
	l, dataDir, key := enrolledWorkshop(t, leftoversControl, "", "", map[string]string{
		"scripts/leftovers.sh": `echo x > "$TEMP_DIR/left"` + "\n" + startedMark("ran") + loopUntilGoFile,
	})
	s := startWorkshopControl(t, l, key, 0)
	waitFor(t, filepath.Join(l.WorkshopFolder(key), "ran"))

	remove(t, l, key, false, false)

	if r := s.wait(t); !r.Stopped {
		t.Errorf("result = %+v, want stopped", r)
	}
	if exists(filepath.Join(dataDir, tempDir, key)) {
		t.Error("TEMP_DIR survived")
	}
	if exists(l.WorkshopFolder(key)) {
		t.Error("workshop folder survived")
	}
}

func TestRemoveStopsCheck(t *testing.T) {
	l, _ := newLib(t)
	e := addLocal(t, l, checkCourse(t, loopUntilGo))
	s := startCheck(t, l, e.Key)

	remove(t, l, e.Key, false, false)

	if r := s.wait(t); !r.Stopped {
		t.Errorf("check result = %+v, want stopped", r)
	}
	if exists(l.CopyPath(e.Key)) {
		t.Error("course copy survived")
	}
}

func TestRemoveThenAddKeepsProgress(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), threeActivities(""))
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	writeFile(t, statePath(dataDir, e.Key), `{"version": 1, "current": "loops", "completed": ["hello"]}`)

	remove(t, l, e.Key, true, false)
	// The kept progress survives a restart.
	l = openLib(t, dataDir)
	again := addLocal(t, l, src)

	o, err := l.Overview(again.Key)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if !o.Enrolled {
		t.Error("not enrolled after re-adding with kept progress")
	}
	if cur, err := l.Enter(again.Key); err != nil || cur != "loops" {
		t.Errorf("Enter = %q, %v, want loops", cur, err)
	}
}

func TestRemoveThenAddNotEnrolled(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	remove(t, l, e.Key, false, false)
	again := addLocal(t, l, src)

	o, err := l.Overview(again.Key)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if o.Enrolled {
		t.Error("enrolled after re-adding without kept progress")
	}
}
