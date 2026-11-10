package library

import (
	"bytes"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/platform"
	"github.com/olymp-software/chiron/internal/runtime"
)

// helperEnv makes the test binary act as a server: it listens on a port,
// writes the address to the file in its first argument and blocks.
const helperEnv = "LIBRARY_TEST_HELPER"

func TestMain(m *testing.M) {
	if os.Getenv(helperEnv) != "" {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			os.Exit(3)
		}
		defer func() { _ = ln.Close() }()
		if err := platform.WriteFileAtomic(os.Args[1], []byte(ln.Addr().String()), 0o644); err != nil {
			os.Exit(3)
		}
		select {}
	}
	os.Exit(m.Run())
}

const workshopTheory = "theory_blocks:\n  - kind: markdown\n    data: {content: '# Hi'}\n"

// workshopYAML is a workshop.yaml with activities basics/first and
// basics/second, with extra top-level fields.
func workshopYAML(extra string) string {
	return "version: 1\ntitle: Rollouts\n" + extra + "sections:\n" +
		"  - {slug: basics, title: Basics, activities: [{slug: first, title: First}, {slug: second, title: Second}]}\n"
}

// workshopSource creates an author's workshop folder. first and second are
// extra fields of the activities; files are paths from the workshop root with
// their content.
func workshopSource(t *testing.T, extra, first, second string, files map[string]string) string {
	t.Helper()
	src := courseFolder(t, "Rollouts")
	if err := os.Remove(filepath.Join(src, manifest.CourseFile)); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(src, manifest.WorkshopFile), workshopYAML(extra))
	writeFile(t, filepath.Join(src, "basics", "first", manifest.ActivityFile), "version: 1\n"+workshopTheory+first)
	writeFile(t, filepath.Join(src, "basics", "second", manifest.ActivityFile), "version: 1\n"+workshopTheory+second)
	for path, content := range files {
		writeFile(t, filepath.Join(src, filepath.FromSlash(path)), content)
	}
	return src
}

// enrolledWorkshop adds and enrolls a workshop from workshopSource.
func enrolledWorkshop(t *testing.T, extra, first, second string, files map[string]string) (*Library, string, string) {
	t.Helper()
	l, dataDir := newLib(t)
	e := addLocal(t, l, workshopSource(t, extra, first, second, files))
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	return l, dataDir, e.Key
}

func openActivity(t *testing.T, l *Library, key, slug string) Activity {
	t.Helper()
	a, err := l.Activity(key, slug)
	if err != nil {
		t.Fatalf("Activity: %v", err)
	}
	if a.Error != nil {
		t.Fatalf("activity error: %v", a.Error)
	}
	return a
}

func completeActivity(t *testing.T, l *Library, key, slug string) {
	t.Helper()
	if _, _, err := l.CompleteActivity(key, slug); err != nil {
		t.Fatalf("CompleteActivity: %v", err)
	}
}

func startControl(t *testing.T, l *Library, key, slug string, index int) *started {
	t.Helper()
	s := &started{done: make(chan runtime.Result, 1)}
	var err error
	if s.stop, err = l.StartControl(key, slug, index, &s.out, func(r runtime.Result) { s.done <- r }); err != nil {
		t.Fatalf("StartControl: %v", err)
	}
	return s
}

func startActivityCheck(t *testing.T, l *Library, key, slug string) *started {
	t.Helper()
	s := &started{done: make(chan runtime.Result, 1)}
	var err error
	if s.stop, err = l.StartActivityCheck(key, slug, &s.out, func(r runtime.Result) { s.done <- r }); err != nil {
		t.Fatalf("StartActivityCheck: %v", err)
	}
	return s
}

func startWorkshopControl(t *testing.T, l *Library, key string, index int) *started {
	t.Helper()
	s := &started{done: make(chan runtime.Result, 1)}
	var err error
	if s.stop, err = l.StartWorkshopControl(key, index, &s.out, func(r runtime.Result) { s.done <- r }); err != nil {
		t.Fatalf("StartWorkshopControl: %v", err)
	}
	return s
}

// refused returns the error of a start that must fail.
func refused(stop func(), err error) error {
	if stop != nil {
		stop()
	}
	return err
}

// waitFor waits until the file exists.
func waitFor(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for !exists(path) {
		if time.Now().After(deadline) {
			t.Fatalf("%s did not appear", path)
		}
		time.Sleep(time.Millisecond)
	}
}

func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// loopUntilGoFile runs until a file named go appears in the workshop folder.
const loopUntilGoFile = `while [ ! -f "$WORKSPACE_DIR/go" ]; do :; done` + "\n"

// startedMark makes a script write a file named after it to the workshop
// folder, so a test can wait until it runs.
func startedMark(name string) string {
	return `echo x > "$WORKSPACE_DIR/` + name + `"` + "\n"
}

// 3.1 Content type

func TestWorkshopInLibrary(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "", "", nil)

	courses, err := l.Courses()
	if err != nil {
		t.Fatal(err)
	}
	if got := courses[0].Type; got != manifest.ContentWorkshop {
		t.Errorf("library type = %q", got)
	}
	o, err := l.Overview(key)
	if err != nil {
		t.Fatal(err)
	}
	if o.Type != manifest.ContentWorkshop {
		t.Errorf("overview type = %q", o.Type)
	}
	tree, err := l.Tree(key)
	if err != nil {
		t.Fatal(err)
	}
	if tree.Type != manifest.ContentWorkshop || tree.WorkshopFolder != l.WorkshopFolder(key) {
		t.Errorf("tree = %+v", tree)
	}
}

func TestWorkshopBothManifestsNotAdded(t *testing.T) {
	src := workshopSource(t, "", "", "", map[string]string{manifest.CourseFile: courseYAML("Both")})
	l, _ := newLib(t)

	_, err := l.AddLocal(src)

	wantCode(t, err, apperr.ManifestConflict)
}

func TestSyncUpdatesContentType(t *testing.T) {
	src := courseFolder(t, "Go Basics")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	if err := os.Remove(filepath.Join(src, manifest.CourseFile)); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(src, manifest.WorkshopFile), workshopYAML(""))

	got, err := l.Sync(e.Key)
	if err != nil {
		t.Fatal(err)
	}

	if got.Type != manifest.ContentWorkshop {
		t.Errorf("type = %q", got.Type)
	}
}

func TestWorkshopActivityFields(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "",
		"solution: Undo it\ncheck: check.sh\ncontrols:\n  - {label: Up, script: up.sh, concurrency: server}\n  - {label: Down, script: down.sh}\n",
		"", map[string]string{"basics/first/check.sh": "", "basics/first/up.sh": "", "basics/first/down.sh": ""})

	a := openActivity(t, l, key, "first")

	want := &WorkshopActivity{
		Solution: "Undo it",
		HasCheck: true,
		Controls: []ActivityControl{{Label: "Up", Group: "server"}, {Label: "Down"}},
	}
	if a.Type != "" || !reflect.DeepEqual(a.Workshop, want) {
		t.Errorf("activity = %+v, workshop = %+v, want %+v", a, a.Workshop, want)
	}
	if len(a.TheoryBlocks) != 1 {
		t.Errorf("theory blocks = %+v", a.TheoryBlocks)
	}
}

func TestWorkshopBrokenActivity(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "type: theory\n", "", nil)

	a, err := l.Activity(key, "first")
	if err != nil {
		t.Fatal(err)
	}

	if a.Error == nil || a.Error.Code != apperr.ManifestInvalid || a.Error.Problems[0].Params["field"] != "type" {
		t.Errorf("activity error = %v", a.Error)
	}
	completeActivity(t, l, key, "first")
}

// 3.2 The workshop folder

func TestWorkshopFolderEmpty(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "", "", nil)

	openActivity(t, l, key, "first")

	if got := dirEntries(t, l.WorkshopFolder(key)); len(got) != 0 {
		t.Errorf("folder = %v, want it empty", got)
	}
}

func TestWorkshopFolderDeleted(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "files:\n  - path: main.go\n", "",
		map[string]string{"basics/first/main.go": "package main\n"})
	openActivity(t, l, key, "first")
	if err := os.RemoveAll(l.WorkshopFolder(key)); err != nil {
		t.Fatal(err)
	}

	openActivity(t, l, key, "first")

	if got := readFile(t, filepath.Join(l.WorkshopFolder(key), "main.go")); got != "package main\n" {
		t.Errorf("main.go = %q", got)
	}
}

func TestWorkshopFilesKeepEdits(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "files:\n  - path: starter\n", "",
		map[string]string{"basics/first/starter/main.go": "original", "basics/first/starter/go.mod": "module"})
	openActivity(t, l, key, "first")
	folder := l.WorkshopFolder(key)
	writeFile(t, filepath.Join(folder, "starter", "main.go"), "edited")
	if err := os.Remove(filepath.Join(folder, "starter", "go.mod")); err != nil {
		t.Fatal(err)
	}

	openActivity(t, l, key, "first")

	if got := readFile(t, filepath.Join(folder, "starter", "main.go")); got != "edited" {
		t.Errorf("edited file = %q", got)
	}
	if got := readFile(t, filepath.Join(folder, "starter", "go.mod")); got != "module" {
		t.Errorf("missing file = %q, want it copied again", got)
	}
}

func TestWorkshopFilesMerge(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "files:\n  - path: manifests\n", "files:\n  - path: manifests\n",
		map[string]string{"basics/first/manifests/a.yaml": "a", "basics/second/manifests/b.yaml": "b"})

	openActivity(t, l, key, "first")
	writeFile(t, filepath.Join(l.WorkshopFolder(key), "notes.txt"), "mine")
	completeActivity(t, l, key, "first")
	openActivity(t, l, key, "second")

	got := dirEntries(t, filepath.Join(l.WorkshopFolder(key), "manifests"))
	if want := []string{"a.yaml", "b.yaml"}; !reflect.DeepEqual(got, want) {
		t.Errorf("manifests = %v, want %v", got, want)
	}
	if !exists(filepath.Join(l.WorkshopFolder(key), "notes.txt")) {
		t.Error("the learner's file is gone in the next activity")
	}
}

func TestWorkshopFilesExact(t *testing.T) {
	content := "line\r\nline\n\x00\xff\xfe"
	l, _, key := enrolledWorkshop(t, "", "files:\n  - path: data.bin\n  - path: run.sh\n", "",
		map[string]string{"basics/first/data.bin": content, "basics/first/run.sh": "exit 0\n"})
	run := filepath.Join(l.CopyPath(key), "basics", "first", "run.sh")
	if err := os.Chmod(run, 0o755); err != nil {
		t.Fatal(err)
	}

	openActivity(t, l, key, "first")

	folder := l.WorkshopFolder(key)
	if got, err := os.ReadFile(filepath.Join(folder, "data.bin")); err != nil || !bytes.Equal(got, []byte(content)) {
		t.Errorf("data.bin = %q, %v", got, err)
	}
	if goruntime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(filepath.Join(folder, "run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("run.sh mode = %v, want it executable", info.Mode())
	}
}

// The "Interrupted copy" scenario: what the copy left is replaced by the file.
func TestWorkshopInterruptedCopy(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "files:\n  - path: big.bin\n", "",
		map[string]string{"basics/first/big.bin": "all of it"})
	folder := l.WorkshopFolder(key)
	writeFile(t, filepath.Join(folder, "big.bin"+partSuffix), "al")

	openActivity(t, l, key, "first")

	if got := readFile(t, filepath.Join(folder, "big.bin")); got != "all of it" {
		t.Errorf("big.bin = %q", got)
	}
	if got := dirEntries(t, folder); !reflect.DeepEqual(got, []string{"big.bin"}) {
		t.Errorf("folder = %v, want only big.bin", got)
	}
}

func TestWorkshopFolderWhereFileExpected(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "files:\n  - path: conf/app.yaml\n", "",
		map[string]string{"basics/first/conf/app.yaml": "a: 1\n"})
	mine := filepath.Join(l.WorkshopFolder(key), "conf", "app.yaml", "mine.txt")
	writeFile(t, mine, "mine")

	a, err := l.Activity(key, "first")
	if err != nil {
		t.Fatal(err)
	}

	want := apperr.New(apperr.FileCopyFailed, "path", "conf/app.yaml")
	if !reflect.DeepEqual(a.Error, want) {
		t.Errorf("activity error = %v, want %v", a.Error, want)
	}
	if got := readFile(t, mine); got != "mine" {
		t.Errorf("file in the folder = %q", got)
	}
}

func TestWorkshopFileWhereFolderExpected(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "files:\n  - path: conf\n", "",
		map[string]string{"basics/first/conf/app.yaml": "a: 1\n"})
	writeFile(t, filepath.Join(l.WorkshopFolder(key), "conf"), "mine")

	a, err := l.Activity(key, "first")
	if err != nil {
		t.Fatal(err)
	}

	if a.Error == nil || a.Error.Code != apperr.FileCopyFailed || a.Error.Params["path"] != "conf" {
		t.Errorf("activity error = %v", a.Error)
	}
	if got := readFile(t, filepath.Join(l.WorkshopFolder(key), "conf")); got != "mine" {
		t.Errorf("file in the folder = %q", got)
	}
}

// 3.3 The script environment

func TestTempDirEmptyOnEveryRun(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "check: check.sh\n", "", map[string]string{
		"basics/first/check.sh": `[ -d "$TEMP_DIR" ] || echo no folder
[ -e "$TEMP_DIR/left" ] && echo left
echo x > "$TEMP_DIR/left"
echo "$TEMP_DIR"
`,
	})

	for range 2 {
		s := startActivityCheck(t, l, key, "first")
		s.wait(t)
		out := strings.TrimSpace(s.out.String())
		if !filepath.IsAbs(out) {
			t.Fatalf("output = %q, want only the path of an empty TEMP_DIR", out)
		}
		if strings.HasPrefix(out, l.WorkshopFolder(key)) {
			t.Errorf("TEMP_DIR = %q is in the workshop folder", out)
		}
	}
}

func TestTempDirGoneAfterStop(t *testing.T) {
	l, dataDir, key := enrolledWorkshop(t, "", "check: check.sh\n", "", map[string]string{
		"basics/first/check.sh": `echo x > "$TEMP_DIR/left"` + "\n" + startedMark("ran") + loopUntilGoFile,
	})
	s := startActivityCheck(t, l, key, "first")
	waitFor(t, filepath.Join(l.WorkshopFolder(key), "ran"))

	s.stop()

	if r := s.wait(t); !r.Stopped {
		t.Errorf("result = %+v, want stopped", r)
	}
	if got := dirEntries(t, filepath.Join(dataDir, tempDir, key)); len(got) != 0 {
		t.Errorf("temp folder = %v, want it empty", got)
	}
}

func TestTempDirCleanedOnStart(t *testing.T) {
	dataDir := t.TempDir()
	left := filepath.Join(dataDir, tempDir, "go-basics-0000", "run-1", "left")
	writeFile(t, left, "")

	l, err := Open(dataDir, filepath.Join(dataDir, "learner"), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	// Opening alone, as a second app instance does before it quits, keeps
	// what the running one uses.
	if !exists(left) {
		t.Fatal("Open deleted TEMP_DIRs")
	}
	if err := l.Recover(); err != nil {
		t.Fatal(err)
	}

	if exists(filepath.Join(dataDir, tempDir)) {
		t.Error("leftovers not deleted on start")
	}
}

func TestContentDirHelper(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "check: check.sh\n", "", map[string]string{
		"lib/server.sh":         "echo helper loaded\n",
		"basics/first/check.sh": `. "$CONTENT_DIR/lib/server.sh"` + "\n",
	})

	s := startActivityCheck(t, l, key, "first")

	if r := s.wait(t); r.ExitCode != 0 {
		t.Errorf("result = %+v, output = %q", r, s.out.String())
	}
	if got := s.out.String(); got != "helper loaded\n" {
		t.Errorf("output = %q", got)
	}
}

func TestEnvironmentCheckHasNoScriptVariables(t *testing.T) {
	t.Setenv("CONTENT_DIR", "content")
	t.Setenv("TEMP_DIR", "temp")
	src := workshopSource(t, "check: check.sh\n", "", "",
		map[string]string{"check.sh": `echo "[$CONTENT_DIR][$TEMP_DIR]"`})
	l, _ := newLib(t)
	e := addLocal(t, l, src)

	s := startCheck(t, l, e.Key)
	s.wait(t)

	if got := s.out.String(); got != "[][]\n" {
		t.Errorf("output = %q", got)
	}
}

func TestActivityScriptEnvironment(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "controls:\n  - {label: Env, script: env.sh}\n", "", map[string]string{
		"basics/first/env.sh": `echo "$ACTIVITY_DIR"; echo "$WORKSPACE_DIR"; echo "$CONTENT_DIR"; pwd`,
	})

	s := startControl(t, l, key, "first", 0)
	s.wait(t)

	dir := filepath.Join(l.CopyPath(key), "basics", "first")
	want := strings.Join([]string{dir, l.WorkshopFolder(key), l.CopyPath(key), dir, ""}, "\n")
	if got := s.out.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

// A late stop of a finished run must not stop the next run of the control.
func TestStopHitsOnlyItsRun(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "controls:\n  - {label: Once, script: once.sh}\n", "", map[string]string{
		// Each run waits for its own go file and consumes it.
		"basics/first/once.sh": loopUntilGoFile + `rm "$WORKSPACE_DIR/go"` + "\n",
	})
	goFile := filepath.Join(l.WorkshopFolder(key), "go")
	old := startControl(t, l, key, "first", 0)
	writeFile(t, goFile, "")
	old.wait(t)
	again := startControl(t, l, key, "first", 0)

	old.stop()

	writeFile(t, goFile, "")
	if r := again.wait(t); r.Stopped || r.ExitCode != 0 {
		t.Errorf("new run result = %+v, want it to finish on its own", r)
	}
}

// 3.4 Runs

const twoControls = "controls:\n  - {label: One, script: one.sh}\n  - {label: Two, script: two.sh}\n" +
	"  - {label: Server, script: server.sh, concurrency: server}\n"

func twoControlFiles() map[string]string {
	return map[string]string{
		"basics/first/one.sh":    startedMark("one") + loopUntilGoFile,
		"basics/first/two.sh":    "exit 0\n",
		"basics/first/server.sh": startedMark("server") + loopUntilGoFile,
	}
}

func TestDefaultGroupExcludes(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", twoControls, "", twoControlFiles())
	s := startControl(t, l, key, "first", 0)

	wantCode(t, refused(l.StartControl(key, "first", 1, &syncBuffer{}, func(runtime.Result) {})), apperr.ScriptRunning)
	// The same control counts too.
	wantCode(t, refused(l.StartControl(key, "first", 0, &syncBuffer{}, func(runtime.Result) {})), apperr.ScriptRunning)

	writeFile(t, filepath.Join(l.WorkshopFolder(key), "go"), "")
	if r := s.wait(t); r.ExitCode != 0 || r.Stopped {
		t.Errorf("result = %+v, want the running control untouched", r)
	}
	if r := startControl(t, l, key, "first", 1).wait(t); r.ExitCode != 0 {
		t.Errorf("result after the first finished = %+v", r)
	}
}

func TestSeparateGroupRunsBeside(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", twoControls, "", twoControlFiles())
	one := startControl(t, l, key, "first", 0)
	server := startControl(t, l, key, "first", 2)

	server.stop()

	if r := server.wait(t); !r.Stopped {
		t.Errorf("server result = %+v, want stopped", r)
	}
	writeFile(t, filepath.Join(l.WorkshopFolder(key), "go"), "")
	if r := one.wait(t); r.Stopped {
		t.Errorf("first result = %+v, want it to keep running", r)
	}
}

func TestStopActivityRuns(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", twoControls, "", twoControlFiles())
	one := startControl(t, l, key, "first", 0)
	server := startControl(t, l, key, "first", 2)

	l.StopActivityRuns(key)

	for _, s := range []*started{one, server} {
		if r := s.wait(t); !r.Stopped {
			t.Errorf("result = %+v, want stopped", r)
		}
	}
}

func TestSyncStopsControls(t *testing.T) {
	src := workshopSource(t, "", twoControls, "", twoControlFiles())
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatal(err)
	}
	s := startControl(t, l, e.Key, "first", 0)

	syncCourse(t, l, e.Key)

	if r := s.wait(t); !r.Stopped {
		t.Errorf("result = %+v, want stopped", r)
	}
}

func TestStopAllStopsControls(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", twoControls, "", twoControlFiles())
	s := startControl(t, l, key, "first", 0)

	l.StopAll()

	select {
	case r := <-s.done:
		if !r.Stopped {
			t.Errorf("result = %+v, want stopped", r)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("control did not stop")
	}
}

func TestControlOfLockedActivityRefused(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "", "controls:\n  - {label: Up, script: up.sh}\n",
		map[string]string{"basics/second/up.sh": ""})

	err := refused(l.StartControl(key, "second", 0, &syncBuffer{}, func(runtime.Result) {}))

	wantCode(t, err, apperr.ActivityUnavailable)
}

// 3.5 Check

func TestCheckRunsBesideControl(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", twoControls+"check: check.sh\n", "", func() map[string]string {
		files := twoControlFiles()
		files["basics/first/check.sh"] = "exit 1\n"
		return files
	}())
	server := startControl(t, l, key, "first", 2)

	if r := startActivityCheck(t, l, key, "first").wait(t); r.ExitCode != 1 || r.Stopped {
		t.Errorf("check result = %+v", r)
	}

	server.stop()
	if r := server.wait(t); !r.Stopped {
		t.Errorf("server result = %+v, want it running until stopped", r)
	}
}

func TestCheckRemembersNothing(t *testing.T) {
	l, dataDir, key := enrolledWorkshop(t, "", "check: check.sh\n", "",
		map[string]string{"basics/first/check.sh": "exit 0\n"})
	before := readFile(t, filepath.Join(dataDir, fileName))

	startActivityCheck(t, l, key, "first").wait(t)

	if got := readFile(t, filepath.Join(dataDir, fileName)); got != before {
		t.Errorf("library.json = %s\nwant %s", got, before)
	}
}

// useHelper lets scripts run this test binary as a server through $HELPER.
func useHelper(t *testing.T) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HELPER", exe)
	t.Setenv(helperEnv, "1")
}

// wantServerGone fails if the server that wrote the address file still
// accepts connections. A new file can stay locked for a moment on Windows.
func wantServerGone(t *testing.T, addrFile string) {
	t.Helper()
	var addr []byte
	var err error
	for range 100 {
		if addr, err = os.ReadFile(addrFile); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	if conn, err := net.DialTimeout("tcp", string(addr), time.Second); err == nil {
		_ = conn.Close()
		t.Errorf("%s still accepts connections", addr)
	}
}

// waitAddr is a script fragment that waits for the server's address file.
const waitAddr = `while [ ! -f "$WORKSPACE_DIR/addr" ]; do :; done` + "\n"

// The "Check leaves a server running" scenario.
func TestBackgroundProcessGoneBeforeExitCode(t *testing.T) {
	useHelper(t)
	l, _, key := enrolledWorkshop(t, "", "check: check.sh\n", "", map[string]string{
		"basics/first/check.sh": `"$HELPER" "$WORKSPACE_DIR/addr" &` + "\n" + waitAddr,
	})

	s := startActivityCheck(t, l, key, "first")
	r := s.wait(t)

	if r.ExitCode != 0 {
		t.Fatalf("result = %+v, output = %q", r, s.out.String())
	}
	wantServerGone(t, filepath.Join(l.WorkshopFolder(key), "addr"))
}

// The "Binary in TEMP_DIR" scenario.
func TestBinaryInTempDirGone(t *testing.T) {
	useHelper(t)
	l, dataDir, key := enrolledWorkshop(t, "", "check: check.sh\n", "", map[string]string{
		"basics/first/check.sh": `cp "$HELPER" "$TEMP_DIR/server.exe"` + "\n" +
			`"$TEMP_DIR/server.exe" "$WORKSPACE_DIR/addr" &` + "\n" + waitAddr,
	})

	s := startActivityCheck(t, l, key, "first")
	r := s.wait(t)

	if r.ExitCode != 0 {
		t.Fatalf("result = %+v, output = %q", r, s.out.String())
	}
	wantServerGone(t, filepath.Join(l.WorkshopFolder(key), "addr"))
	if got := dirEntries(t, filepath.Join(dataDir, tempDir, key)); len(got) != 0 {
		t.Errorf("temp folder = %v, want it empty", got)
	}
}

// 3.5a Workshop controls

const leftoversControl = "controls:\n  - {label: Leftovers, script: scripts/leftovers.sh}\n"

func TestWorkshopControlEnvironment(t *testing.T) {
	t.Setenv("ACTIVITY_DIR", "inherited")
	l, dataDir, key := enrolledWorkshop(t, leftoversControl, "", "", map[string]string{
		"scripts/leftovers.sh": `echo "[$ACTIVITY_DIR]"; echo "$WORKSPACE_DIR"; echo "$CONTENT_DIR"; pwd
[ -d "$TEMP_DIR" ] || echo no temp folder
`,
	})
	before := readFile(t, filepath.Join(dataDir, fileName))

	s := startWorkshopControl(t, l, key, 0)
	s.wait(t)

	want := strings.Join([]string{"[]", l.WorkshopFolder(key), l.CopyPath(key), l.CopyPath(key), ""}, "\n")
	if got := s.out.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
	if got := readFile(t, filepath.Join(dataDir, fileName)); got != before {
		t.Errorf("library.json changed: %s", got)
	}
}

func TestWorkshopControlWithoutSatisfiedCheck(t *testing.T) {
	src := workshopSource(t, "check: check.sh\n"+leftoversControl, "", "", map[string]string{
		"check.sh":             "exit 0\n",
		"scripts/leftovers.sh": "exit 0\n",
	})
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	runCheck(t, l, e.Key)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatal(err)
	}
	syncCheck(t, l, e.Key, src, "echo changed\n")
	wantStatus(t, l, e.Key, CheckChanged)

	if r := startWorkshopControl(t, l, e.Key, 0).wait(t); r.ExitCode != 0 {
		t.Errorf("result = %+v", r)
	}
}

func TestWorkshopControlsNeedEnrollment(t *testing.T) {
	src := workshopSource(t, leftoversControl, "", "", map[string]string{"scripts/leftovers.sh": "exit 0\n"})
	l, _ := newLib(t)
	e := addLocal(t, l, src)

	o, err := l.Overview(e.Key)
	if err != nil {
		t.Fatal(err)
	}
	if len(o.Controls) != 0 {
		t.Errorf("controls = %v, want none", o.Controls)
	}
	wantCode(t, refused(l.StartWorkshopControl(e.Key, 0, &syncBuffer{}, func(runtime.Result) {})), apperr.NotEnrolled)

	if err := l.Enroll(e.Key); err != nil {
		t.Fatal(err)
	}
	if o, err = l.Overview(e.Key); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(o.Controls, []string{"Leftovers"}) {
		t.Errorf("controls = %v", o.Controls)
	}
}

// A symbolic link leading outside the course copy cannot be read as a script.
func TestScriptOutsideCopyRefused(t *testing.T) {
	l, _, key := enrolledWorkshop(t, "", "check: check.sh\n", "",
		map[string]string{"basics/first/check.sh": "exit 0\n"})
	check := filepath.Join(l.CopyPath(key), "basics", "first", "check.sh")
	outside := filepath.Join(t.TempDir(), "check.sh")
	writeFile(t, outside, "exit 0\n")
	if err := os.Remove(check); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, check); err != nil {
		t.Skipf("symbolic links are not available: %v", err)
	}

	err := refused(l.StartActivityCheck(key, "first", &syncBuffer{}, func(runtime.Result) {}))

	wantCode(t, err, apperr.ScriptStartFailed)
}
