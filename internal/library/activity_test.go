package library

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
)

func theoryActivity(content string) string {
	return "version: 1\ntype: theory\ntheory_blocks:\n  - kind: markdown\n    data: {content: '" + content + "'}\n"
}

// enrolledCourse adds and enrolls a course with activities basics/hello,
// more/loops and more/maps, each a valid theory activity. It returns the
// library, the data folder, the course folder and the storage key.
func enrolledCourse(t *testing.T) (*Library, string, string, string) {
	t.Helper()
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), threeActivities(""))
	for _, dir := range []string{"basics/hello", "more/loops", "more/maps"} {
		writeFile(t, filepath.Join(src, filepath.FromSlash(dir), manifest.ActivityFile), theoryActivity("# "+dir))
	}
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	return l, dataDir, src, e.Key
}

func syncCourse(t *testing.T, l *Library, key string) {
	t.Helper()
	if _, err := l.Sync(key); err != nil {
		t.Fatalf("Sync: %v", err)
	}
}

func TestActivityTheory(t *testing.T) {
	l, _, src, key := enrolledCourse(t)
	writeFile(t, filepath.Join(src, "basics", "hello", manifest.ActivityFile),
		"version: 1\ntype: theory\ntheory_blocks:\n"+
			"  - kind: markdown\n    data: {content: '# One'}\n"+
			"  - kind: markdown\n    data: {content: '# Two'}\n")
	syncCourse(t, l, key)

	a, err := l.Activity(key, "hello")
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}

	want := Activity{
		Slug:  "hello",
		Title: "Hello",
		Dir:   "basics/hello",
		Type:  manifest.TypeTheory,
		TheoryBlocks: []TheoryBlock{
			{Kind: manifest.KindMarkdown, Content: "# One"},
			{Kind: manifest.KindMarkdown, Content: "# Two"},
		},
	}
	if !reflect.DeepEqual(a, want) {
		t.Errorf("activity = %+v\nwant %+v", a, want)
	}
}

func TestActivityQuiz(t *testing.T) {
	l, _, src, key := enrolledCourse(t)
	writeFile(t, filepath.Join(src, "basics", "hello", manifest.ActivityFile),
		"version: 1\ntype: quiz\ntheory_blocks:\n"+
			"  - kind: markdown\n    data: {content: '# Quiz'}\n"+
			"data:\n  questions:\n"+
			"    - kind: choice\n      prompt: Pick one\n      explanation: Because\n"+
			"      data:\n        options:\n"+
			"          - {text: A, correct: true}\n"+
			"          - {text: B, feedback: Not B}\n"+
			"    - kind: choice\n"+
			"      data:\n        multiple: true\n        options:\n"+
			"          - {text: C, correct: true}\n"+
			"          - {text: D}\n")
	syncCourse(t, l, key)

	a, err := l.Activity(key, "hello")
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}

	want := Activity{
		Slug:         "hello",
		Title:        "Hello",
		Dir:          "basics/hello",
		Type:         manifest.TypeQuiz,
		TheoryBlocks: []TheoryBlock{{Kind: manifest.KindMarkdown, Content: "# Quiz"}},
		Quiz: &Quiz{Questions: []Question{
			{
				Kind:        "choice",
				Prompt:      "Pick one",
				Explanation: "Because",
				Options:     []Option{{Text: "A", Correct: true}, {Text: "B", Feedback: "Not B"}},
			},
			{
				Kind:     "choice",
				Multiple: true,
				Options:  []Option{{Text: "C", Correct: true}, {Text: "D"}},
			},
		}},
	}
	if !reflect.DeepEqual(a, want) {
		t.Errorf("activity = %+v\nwant %+v", a, want)
	}
}

func TestActivityMissingFile(t *testing.T) {
	l, _, src, key := enrolledCourse(t)
	if err := os.Remove(filepath.Join(src, "basics", "hello", manifest.ActivityFile)); err != nil {
		t.Fatal(err)
	}
	syncCourse(t, l, key)

	a, err := l.Activity(key, "hello")
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}

	want := Activity{
		Slug:  "hello",
		Title: "Hello",
		Dir:   "basics/hello",
		Error: apperr.New(apperr.ActivityFileMissing, "file", "basics/hello/activity.yaml"),
	}
	if !reflect.DeepEqual(a, want) {
		t.Errorf("activity = %+v\nwant %+v", a, want)
	}
	if _, err := l.Tree(key); err != nil {
		t.Errorf("Tree: %v", err)
	}
}

func TestActivityUnavailable(t *testing.T) {
	l, _, _, key := enrolledCourse(t)

	for _, slug := range []string{"loops", "nope"} {
		_, err := l.Activity(key, slug)
		wantCode(t, err, apperr.ActivityUnavailable)
	}
}

func TestActivityFixedBySync(t *testing.T) {
	l, _, src, key := enrolledCourse(t)
	path := filepath.Join(src, "basics", "hello", manifest.ActivityFile)
	writeFile(t, path, "version: 1\ntype: theory\ntheory_blocks: []\n")
	syncCourse(t, l, key)
	if a, err := l.Activity(key, "hello"); err != nil || a.Error == nil {
		t.Fatalf("broken Activity = %+v, %v, want an activity error", a, err)
	}

	writeFile(t, path, theoryActivity("# Fixed"))
	syncCourse(t, l, key)
	a, err := l.Activity(key, "hello")

	if err != nil || a.Error != nil || len(a.TheoryBlocks) != 1 || a.TheoryBlocks[0].Content != "# Fixed" {
		t.Errorf("fixed Activity = %+v, %v, want its theory block", a, err)
	}
}

// complete completes slug and checks the next activity.
func complete(t *testing.T, l *Library, key, slug, wantNext string) Tree {
	t.Helper()
	next, tree, err := l.CompleteActivity(key, slug)
	if err != nil {
		t.Fatalf("CompleteActivity(%s): %v", slug, err)
	}
	if next != wantNext {
		t.Errorf("CompleteActivity(%s) next = %q, want %q", slug, next, wantNext)
	}
	return tree
}

func TestCompleteCurrentActivity(t *testing.T) {
	l, _, _, key := enrolledCourse(t)

	tree := complete(t, l, key, "hello", "loops")

	if tree.Current != "loops" || !tree.Sections[0].Activities[0].Completed || tree.Sections[1].Activities[0].Locked {
		t.Errorf("tree = %+v, want hello completed and loops current and available", tree)
	}
	if a, err := l.Activity(key, "loops"); err != nil {
		t.Errorf("Activity(loops) = %+v, %v, want available", a, err)
	}
}

func TestCompleteEarlierActivity(t *testing.T) {
	l, dataDir, _, key := enrolledCourse(t)
	writeFile(t, statePath(dataDir, key), `{"version": 1, "current": "maps", "completed": ["hello", "loops"]}`)

	tree := complete(t, l, key, "hello", "loops")

	if tree.Current != "maps" {
		t.Errorf("current = %q, want maps", tree.Current)
	}
}

func TestCompleteBrokenActivity(t *testing.T) {
	l, _, src, key := enrolledCourse(t)
	if err := os.RemoveAll(filepath.Join(src, "basics", "hello")); err != nil {
		t.Fatal(err)
	}
	syncCourse(t, l, key)

	tree := complete(t, l, key, "hello", "loops")

	if tree.Current != "loops" || !tree.Sections[0].Activities[0].Completed {
		t.Errorf("tree = %+v, want hello completed and loops current", tree)
	}
}

func TestFinishCourse(t *testing.T) {
	l, dataDir, _, key := enrolledCourse(t)
	writeFile(t, statePath(dataDir, key), `{"version": 1, "current": "maps", "completed": ["hello", "loops"]}`)

	tree := complete(t, l, key, "maps", "")

	if tree.Current != "maps" || !tree.Completed {
		t.Errorf("tree = %+v, want maps current and the course completed", tree)
	}
	a, err := l.Activity(key, "maps")
	if err != nil || !a.Last || !a.CourseCompleted {
		t.Errorf("Activity(maps) = %+v, %v, want the last activity of a completed course", a, err)
	}
}

// listedCompleted returns whether the only course of l is listed as
// completed.
func listedCompleted(t *testing.T, l *Library) bool {
	t.Helper()
	courses, err := l.Courses()
	if err != nil || len(courses) != 1 {
		t.Fatalf("Courses = %+v, %v, want one course", courses, err)
	}
	return courses[0].Completed
}

func TestCompletedCourseInLibrary(t *testing.T) {
	l, dataDir, src, key := enrolledCourse(t)
	writeFile(t, statePath(dataDir, key), `{"version": 1, "current": "maps", "completed": ["hello", "loops"]}`)
	if listedCompleted(t, l) {
		t.Fatal("course listed as completed before Finish course")
	}

	complete(t, l, key, "maps", "")

	if !listedCompleted(t, l) {
		t.Error("course not listed as completed after Finish course")
	}

	// A Sync adding an activity after the former last one clears it.
	writeFile(t, filepath.Join(src, manifest.CourseFile), "version: 1\ntitle: Go Basics\nsections:\n"+
		"  - {slug: basics, title: Basics, activities: [{slug: hello, title: Hello}]}\n"+
		"  - {slug: more, title: More, activities: [{slug: loops, title: loops}, {slug: maps, title: maps}, {slug: extra, title: Extra}]}\n")
	syncCourse(t, l, key)

	if listedCompleted(t, l) {
		t.Error("course listed as completed after an activity was added after the last")
	}
	o, err := l.Overview(key)
	if err != nil || o.Completed {
		t.Errorf("Overview = %+v, %v, want not completed", o, err)
	}
	tree, err := l.Tree(key)
	if err != nil || tree.Current != "extra" || tree.Completed {
		t.Errorf("Tree = %+v, %v, want extra current and the course not completed", tree, err)
	}
	if a, err := l.Activity(key, "maps"); err != nil || a.Last || a.CourseCompleted {
		t.Errorf("Activity(maps) = %+v, %v, want Next again", a, err)
	}
}

func TestSyncInsertsActivityIntoFinishedCourse(t *testing.T) {
	l, dataDir, src, key := enrolledCourse(t)
	writeFile(t, statePath(dataDir, key), `{"version": 1, "current": "maps", "completed": ["hello", "loops"]}`)
	complete(t, l, key, "maps", "")
	writeFile(t, filepath.Join(src, manifest.CourseFile), "version: 1\ntitle: Go Basics\nsections:\n"+
		"  - {slug: basics, title: Basics, activities: [{slug: hello, title: Hello}]}\n"+
		"  - {slug: more, title: More, activities: [{slug: middle, title: Middle}, {slug: loops, title: loops}, {slug: maps, title: maps}]}\n")
	writeFile(t, filepath.Join(src, "more", "middle", manifest.ActivityFile), theoryActivity("# Middle"))
	syncCourse(t, l, key)

	if o, err := l.Overview(key); err != nil || o.Completed {
		t.Errorf("Overview = %+v, %v, want not completed", o, err)
	}
	tree, err := l.Tree(key)
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}
	if tree.Current != "middle" || tree.Completed {
		t.Errorf("tree = %+v, want middle current and the course not completed", tree)
	}
	for _, a := range tree.Sections[1].Activities {
		if a.Locked {
			t.Errorf("activity %s locked, want completed activities after the current one available", a.Slug)
		}
	}

	tree = complete(t, l, key, "middle", "loops")

	if tree.Current != "maps" || !tree.Completed {
		t.Errorf("tree = %+v, want the course completed again", tree)
	}
}

func TestSyncInsertsSeveralActivities(t *testing.T) {
	l, dataDir, src, key := enrolledCourse(t)
	writeFile(t, statePath(dataDir, key), `{"version": 1, "current": "maps", "completed": ["hello", "loops"]}`)
	complete(t, l, key, "maps", "")
	writeFile(t, filepath.Join(src, manifest.CourseFile), "version: 1\ntitle: Go Basics\nsections:\n"+
		"  - {slug: basics, title: Basics, activities: [{slug: hello, title: Hello}, {slug: x, title: X}]}\n"+
		"  - {slug: more, title: More, activities: [{slug: loops, title: loops}, {slug: y, title: Y}, {slug: maps, title: maps}]}\n")
	syncCourse(t, l, key)

	tree, err := l.Tree(key)
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}
	if tree.Current != "x" || tree.Completed {
		t.Errorf("tree = %+v, want x current and the course not completed", tree)
	}
	if !tree.Sections[1].Activities[1].Locked {
		t.Errorf("tree = %+v, want y locked", tree)
	}

	tree = complete(t, l, key, "x", "loops")

	if tree.Current != "y" || tree.Sections[1].Activities[1].Locked {
		t.Errorf("tree = %+v, want y current", tree)
	}

	tree = complete(t, l, key, "y", "maps")

	if tree.Current != "maps" || !tree.Completed {
		t.Errorf("tree = %+v, want the course completed again", tree)
	}
}

func TestFinishCourseWithEarlierActivityLeft(t *testing.T) {
	l, dataDir, _, key := enrolledCourse(t)
	writeFile(t, statePath(dataDir, key), `{"version": 1, "current": "loops", "completed": ["hello", "maps"]}`)

	tree := complete(t, l, key, "maps", "")

	if tree.Current != "loops" || tree.Completed {
		t.Errorf("tree = %+v, want loops current and the course not completed", tree)
	}
}

func TestActivityNonFiniteNumber(t *testing.T) {
	l, _, src, key := enrolledCourse(t)
	writeFile(t, filepath.Join(src, "basics", "hello", manifest.ActivityFile), "version: .inf\n")
	syncCourse(t, l, key)

	a, err := l.Activity(key, "hello")

	if err != nil || a.Error == nil || a.Error.Code != apperr.ManifestInvalid {
		t.Errorf("Activity = %+v, %v, want a manifestInvalid activity error", a, err)
	}
}

func TestEnterNewerProgressFormat(t *testing.T) {
	l, dataDir, _, key := enrolledCourse(t)
	stored := `{"version": 2, "current": "maps", "completed": ["hello"], "answers": {}}`
	writeFile(t, statePath(dataDir, key), stored)

	cur, err := l.Enter(key)

	if err != nil || cur != "loops" {
		t.Errorf("Enter = %q, %v, want loops", cur, err)
	}
	if got := readFile(t, statePath(dataDir, key)); got != stored {
		t.Errorf("state file = %s, want unchanged %s", got, stored)
	}
}

func TestOverviewCompleted(t *testing.T) {
	l, dataDir, _, key := enrolledCourse(t)
	writeFile(t, statePath(dataDir, key), `{"version": 1, "current": "maps", "completed": ["hello", "loops"]}`)
	if o, err := l.Overview(key); err != nil || o.Completed {
		t.Fatalf("Overview = %+v, %v, want not completed", o, err)
	}

	complete(t, l, key, "maps", "")

	if o, err := l.Overview(key); err != nil || !o.Completed {
		t.Errorf("Overview = %+v, %v, want completed", o, err)
	}
}

func TestCompleteLockedActivity(t *testing.T) {
	l, dataDir, _, key := enrolledCourse(t)
	before := readFile(t, statePath(dataDir, key))

	_, _, err := l.CompleteActivity(key, "loops")

	wantCode(t, err, apperr.ActivityUnavailable)
	if got := readFile(t, statePath(dataDir, key)); got != before {
		t.Errorf("state file = %s, want unchanged %s", got, before)
	}
}
