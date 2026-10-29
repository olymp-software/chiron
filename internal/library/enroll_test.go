package library

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
)

// threeActivities is a course.yaml with activities hello, loops, maps; drop
// leaves one of them out.
func threeActivities(drop string) string {
	acts := ""
	for _, a := range []string{"loops", "maps"} {
		if a != drop {
			acts += ", {slug: " + a + ", title: " + a + "}"
		}
	}
	return "version: 1\ntitle: Go Basics\nsections:\n" +
		"  - {slug: basics, title: Basics, activities: [{slug: hello, title: Hello}]}\n" +
		"  - {slug: more, title: More, activities: [" + acts[2:] + "]}\n"
}

func statePath(dataDir, key string) string {
	return filepath.Join(dataDir, stateDir, key+".json")
}

func TestEnrolledFlagFollowsStateFile(t *testing.T) {
	l, dataDir := newLib(t)
	e := addLocal(t, l, courseFolder(t, "Go Basics"))
	libraryJSON := readFile(t, filepath.Join(dataDir, fileName))

	enrolled := func() (bool, bool) {
		t.Helper()
		courses, err := l.Courses()
		if err != nil {
			t.Fatalf("Courses: %v", err)
		}
		o, err := l.Overview(e.Key)
		if err != nil {
			t.Fatalf("Overview: %v", err)
		}
		return courses[0].Enrolled, o.Enrolled
	}

	if list, overview := enrolled(); list || overview {
		t.Errorf("enrolled before Enroll: list %v, overview %v", list, overview)
	}
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	if list, overview := enrolled(); !list || !overview {
		t.Errorf("not enrolled after Enroll: list %v, overview %v", list, overview)
	}
	if err := os.Remove(statePath(dataDir, e.Key)); err != nil {
		t.Fatal(err)
	}
	if list, overview := enrolled(); list || overview {
		t.Errorf("enrolled without a state file: list %v, overview %v", list, overview)
	}

	if got := readFile(t, filepath.Join(dataDir, fileName)); got != libraryJSON {
		t.Errorf("library.json = %s, want unchanged %s", got, libraryJSON)
	}
}

func TestEnrollStartsAtFirstActivity(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), threeActivities(""))
	l, _ := newLib(t)
	e := addLocal(t, l, src)

	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	cur, err := l.Enter(e.Key)
	if err != nil || cur != "hello" {
		t.Errorf("Enter = %q, %v, want hello", cur, err)
	}
}

func TestRepeatedEnrollKeepsProgress(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), threeActivities(""))
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	stored := `{"version": 1, "current": "loops", "completed": ["hello"]}`
	writeFile(t, statePath(dataDir, e.Key), stored)

	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	if got := readFile(t, statePath(dataDir, e.Key)); got != stored {
		t.Errorf("state file = %s, want unchanged", got)
	}
}

func TestEnrollInvalidManifest(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	writeFile(t, filepath.Join(src, manifest.CourseFile), "version: 1\ntitle: Broken\n")
	if _, err := l.Sync(e.Key); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	err := l.Enroll(e.Key)

	wantCode(t, err, apperr.ManifestInvalid)
	if exists(statePath(dataDir, e.Key)) {
		t.Error("course enrolled with an invalid course.yaml")
	}
}

func TestEnterRefused(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)

	_, err := l.Enter("not-added-0000")
	wantCode(t, err, apperr.CourseNotFound)

	_, err = l.Enter(e.Key)
	wantCode(t, err, apperr.NotEnrolled)

	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	writeFile(t, filepath.Join(src, manifest.CourseFile), "version: 1\ntitle: Broken\n")
	if _, err := l.Sync(e.Key); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	_, err = l.Enter(e.Key)
	wantCode(t, err, apperr.ManifestInvalid)
	_, err = l.Tree(e.Key)
	wantCode(t, err, apperr.ManifestInvalid)
}

func TestEnterMovesRemovedCurrentActivity(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), threeActivities(""))
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	writeFile(t, statePath(dataDir, e.Key), `{"version": 1, "current": "loops", "completed": ["hello"]}`)
	writeFile(t, filepath.Join(src, manifest.CourseFile), threeActivities("loops"))
	if _, err := l.Sync(e.Key); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	cur, err := l.Enter(e.Key)

	if err != nil || cur != "maps" {
		t.Errorf("Enter = %q, %v, want maps", cur, err)
	}
	if tree, err := l.Tree(e.Key); err != nil || tree.Current != "maps" {
		t.Errorf("stored current = %q, %v, want maps", tree.Current, err)
	}
}

func TestTree(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), threeActivities(""))
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}

	tree, err := l.Tree(e.Key)
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}

	want := Tree{Title: "Go Basics", Current: "hello", Sections: []TreeSection{
		{Slug: "basics", Title: "Basics", Activities: []TreeActivity{{Slug: "hello", Title: "Hello"}}},
		{Slug: "more", Title: "More", Activities: []TreeActivity{
			{Slug: "loops", Title: "loops"}, {Slug: "maps", Title: "maps"},
		}},
	}}
	if !reflect.DeepEqual(tree, want) {
		t.Errorf("tree = %+v, want %+v", tree, want)
	}
}
