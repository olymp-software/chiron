package library

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/runtime"
)

// loopUntilGo is a check that runs until a file named go appears in the
// course copy, then exits with code 0.
const loopUntilGo = "while [ ! -f go ]; do :; done\n"

// checkCourse creates an author's course folder whose course.yaml declares
// check.sh with the given script.
func checkCourse(t *testing.T, script string) string {
	t.Helper()
	src := courseFolder(t, "Go Basics")
	writeFile(t, filepath.Join(src, manifest.CourseFile), courseYAML("Go Basics")+"check: check.sh\n")
	writeFile(t, filepath.Join(src, "check.sh"), script)
	return src
}

// syncCheck writes script as the check of the course folder src and syncs it.
func syncCheck(t *testing.T, l *Library, key, src, script string) {
	t.Helper()
	writeFile(t, filepath.Join(src, "check.sh"), script)
	if _, err := l.Sync(key); err != nil {
		t.Fatalf("Sync: %v", err)
	}
}

// syncBuffer is check output the test reads while the run writes.
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

// started is a check started with StartCheck.
type started struct {
	out  syncBuffer
	done chan runtime.Result
}

func startCheck(t *testing.T, l *Library, key string) *started {
	t.Helper()
	s := &started{done: make(chan runtime.Result, 1)}
	if err := l.StartCheck(key, &s.out, func(r runtime.Result) { s.done <- r }); err != nil {
		t.Fatalf("StartCheck: %v", err)
	}
	return s
}

// wait returns the result passed to done.
func (s *started) wait(t *testing.T) runtime.Result {
	t.Helper()
	select {
	case r := <-s.done:
		return r
	case <-time.After(30 * time.Second):
		t.Fatal("check did not exit")
		return runtime.Result{}
	}
}

// runCheck runs the check of the course with key to the end.
func runCheck(t *testing.T, l *Library, key string) runtime.Result {
	t.Helper()
	return startCheck(t, l, key).wait(t)
}

func checkStatusOf(t *testing.T, l *Library, key string) string {
	t.Helper()
	o, err := l.Overview(key)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	return o.Check
}

func wantStatus(t *testing.T, l *Library, key, want string) {
	t.Helper()
	if got := checkStatusOf(t, l, key); got != want {
		t.Errorf("check status = %q, want %q", got, want)
	}
}

// 3.1 Check status

func TestCheckStatusWithoutCheck(t *testing.T) {
	l, _ := newLib(t)
	e := addLocal(t, l, courseFolder(t, "Go Basics"))

	wantStatus(t, l, e.Key, "")
}

func TestCheckStatus(t *testing.T) {
	src := checkCourse(t, "exit 0\n")
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)

	wantStatus(t, l, e.Key, CheckNotChecked)

	if r := runCheck(t, l, e.Key); r.ExitCode != 0 || r.Stopped {
		t.Fatalf("check result = %+v, want exit code 0", r)
	}
	wantStatus(t, l, e.Key, CheckPassed)
	if file := readFile(t, filepath.Join(dataDir, fileName)); !strings.Contains(file, `"checkHash": "`+hashOf([]byte("exit 0\n"))+`"`) {
		t.Errorf("library.json = %s, want the check hash", file)
	}
	// The "Unchanged check" scenario: the result survives a restart.
	wantStatus(t, openLib(t, dataDir), e.Key, CheckPassed)

	syncCheck(t, l, e.Key, src, "echo changed\n")
	wantStatus(t, l, e.Key, CheckChanged)
}

func TestCheckStatusBrokenManifest(t *testing.T) {
	src := checkCourse(t, "exit 0\n")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	writeFile(t, filepath.Join(src, manifest.CourseFile), "version: 1\ntitle: Broken\ncheck: check.sh\n")
	if _, err := l.Sync(e.Key); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	wantStatus(t, l, e.Key, "")
}

func TestCheckStatusOldLibraryFile(t *testing.T) {
	src := checkCourse(t, "exit 0\n")
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	old := `{"version": 1, "courses": [{"origin": "local", "source": ` + quote(src) +
		`, "key": "` + e.Key + `", "title": "Go Basics"}]}`
	writeFile(t, filepath.Join(dataDir, fileName), old)

	wantStatus(t, l, e.Key, CheckNotChecked)
}

func quote(s string) string {
	return `"` + strings.ReplaceAll(s, `\`, `\\`) + `"`
}

// A check file that cannot be read, here a link leading outside the course
// copy, is not checked and cannot be run.
func TestCheckUnreadable(t *testing.T) {
	src := checkCourse(t, "exit 0\n")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	runCheck(t, l, e.Key)
	check := filepath.Join(l.CopyPath(e.Key), "check.sh")
	outside := filepath.Join(t.TempDir(), "check.sh")
	writeFile(t, outside, "exit 0\n")
	if err := os.Remove(check); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, check); err != nil {
		t.Skipf("symbolic links are not available: %v", err)
	}

	wantStatus(t, l, e.Key, CheckNotChecked)
	err := l.StartCheck(e.Key, &syncBuffer{}, func(runtime.Result) {})
	wantCode(t, err, apperr.CheckReadFailed)
}

// 3.2 Running the check

// The "Failure after the check file changed" scenario.
func TestCheckFailureKeepsOtherHash(t *testing.T) {
	src := checkCourse(t, "exit 0\n")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	runCheck(t, l, e.Key)

	syncCheck(t, l, e.Key, src, "exit 1\n")
	if r := runCheck(t, l, e.Key); r.ExitCode != 1 {
		t.Fatalf("check result = %+v, want exit code 1", r)
	}
	wantStatus(t, l, e.Key, CheckChanged)

	syncCheck(t, l, e.Key, src, "exit 0\n")
	wantStatus(t, l, e.Key, CheckPassed)
}

// The "Learner broke a working environment" scenario: the check file is
// unchanged, the machine is not.
func TestCheckFailureOfSatisfiedCheck(t *testing.T) {
	src := checkCourse(t, "if [ -f broken ]; then exit 1; fi\n")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	runCheck(t, l, e.Key)
	wantStatus(t, l, e.Key, CheckPassed)

	writeFile(t, filepath.Join(l.CopyPath(e.Key), "broken"), "")
	if r := runCheck(t, l, e.Key); r.ExitCode != 1 {
		t.Fatalf("check result = %+v, want exit code 1", r)
	}
	wantStatus(t, l, e.Key, CheckNotChecked)
}

// The "Cancelled run" scenario.
func TestCheckStoppedChangesNothing(t *testing.T) {
	src := checkCourse(t, "if [ -f loop ]; then while true; do :; done; fi\n")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	runCheck(t, l, e.Key)
	writeFile(t, filepath.Join(l.CopyPath(e.Key), "loop"), "")

	s := startCheck(t, l, e.Key)
	l.StopCheck(e.Key)

	if r := s.wait(t); !r.Stopped {
		t.Errorf("check result = %+v, want stopped", r)
	}
	wantStatus(t, l, e.Key, CheckPassed)
}

func TestCheckRunning(t *testing.T) {
	src := checkCourse(t, loopUntilGo)
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	s := startCheck(t, l, e.Key)

	err := l.StartCheck(e.Key, &syncBuffer{}, func(runtime.Result) {})

	wantCode(t, err, apperr.CheckRunning)
	writeFile(t, filepath.Join(l.CopyPath(e.Key), "go"), "")
	if r := s.wait(t); r.ExitCode != 0 || r.Stopped {
		t.Errorf("check result = %+v, want exit code 0", r)
	}
	// The finished check no longer blocks a new one.
	runCheck(t, l, e.Key)
}

// The check modal closes on Cancel, so the learner can start the check again
// right after CancelCheck returns.
func TestCheckStartsRightAfterStop(t *testing.T) {
	src := checkCourse(t, loopUntilGo)
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	s := startCheck(t, l, e.Key)

	l.StopCheck(e.Key)

	next := startCheck(t, l, e.Key)
	s.wait(t)
	l.StopCheck(e.Key)
	next.wait(t)
}

func TestCheckRemovedEntryIgnored(t *testing.T) {
	src := checkCourse(t, loopUntilGo)
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)
	s := startCheck(t, l, e.Key)

	if err := l.update(func(f *libraryFile) error { f.Courses = nil; return nil }); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(l.CopyPath(e.Key), "go"), "")
	s.wait(t)

	if got := readFile(t, filepath.Join(dataDir, fileName)); strings.Contains(got, e.Key) {
		t.Errorf("library.json = %s, want the removed course to stay removed", got)
	}
}

func TestCheckEnvironment(t *testing.T) {
	t.Setenv("ACTIVITY_DIR", "activity")
	t.Setenv("WORKSPACE_DIR", "workspace")
	t.Setenv("CHIRON_INHERITED", "inherited")
	src := checkCourse(t, `echo "[$ACTIVITY_DIR][$WORKSPACE_DIR][$CHIRON_INHERITED]"; pwd`)
	l, _ := newLib(t)
	e := addLocal(t, l, src)

	s := startCheck(t, l, e.Key)
	s.wait(t)

	if got, want := s.out.String(), "[][][inherited]\n"+l.CopyPath(e.Key)+"\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestCheckWithoutCheckRefused(t *testing.T) {
	l, _ := newLib(t)
	e := addLocal(t, l, courseFolder(t, "Go Basics"))

	if err := l.StartCheck(e.Key, &syncBuffer{}, func(runtime.Result) {}); err == nil {
		t.Error("StartCheck of a course without check: want an error")
	}
	if err := l.StartCheck("not-added-0000", &syncBuffer{}, func(runtime.Result) {}); err == nil {
		t.Error("StartCheck of a missing course: want an error")
	} else {
		wantCode(t, err, apperr.CourseNotFound)
	}
}

// 3.3 The gate

func TestEnrollAndEnterWithoutCheck(t *testing.T) {
	l, _ := newLib(t)
	e := addLocal(t, l, courseFolder(t, "Go Basics"))

	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll: %v", err)
	}
	if _, err := l.Enter(e.Key); err != nil {
		t.Errorf("Enter: %v", err)
	}
}

func TestCheckGatesEnrollAndEnter(t *testing.T) {
	src := checkCourse(t, "exit 0\n")
	l, dataDir := newLib(t)
	e := addLocal(t, l, src)

	wantCode(t, l.Enroll(e.Key), apperr.CheckRequired)
	if exists(statePath(dataDir, e.Key)) {
		t.Fatal("course enrolled without a satisfied check")
	}

	runCheck(t, l, e.Key)
	if err := l.Enroll(e.Key); err != nil {
		t.Fatalf("Enroll after the check: %v", err)
	}
	if _, err := l.Enter(e.Key); err != nil {
		t.Fatalf("Enter after the check: %v", err)
	}

	// The "Version with a changed check" scenario, for a local-folder course.
	syncCheck(t, l, e.Key, src, "echo new\n")
	_, err := l.Enter(e.Key)
	wantCode(t, err, apperr.CheckRequired)

	// The "Required check survives a restart" scenario.
	_, err = openLib(t, dataDir).Enter(e.Key)
	wantCode(t, err, apperr.CheckRequired)
}

func TestSyncStopsRunningCheck(t *testing.T) {
	src := checkCourse(t, "if [ -f loop ]; then while true; do :; done; fi\n")
	l, _ := newLib(t)
	e := addLocal(t, l, src)
	runCheck(t, l, e.Key)
	writeFile(t, filepath.Join(l.CopyPath(e.Key), "loop"), "")
	s := startCheck(t, l, e.Key)

	if _, err := l.Sync(e.Key); err != nil {
		t.Fatalf("Sync: %v", err)
	}

	if r := s.wait(t); !r.Stopped {
		t.Errorf("check result = %+v, want stopped", r)
	}
	if exists(filepath.Join(l.CopyPath(e.Key), "loop")) {
		t.Error("course copy not replaced")
	}
	wantStatus(t, l, e.Key, CheckPassed)
}

func TestStopAll(t *testing.T) {
	l, _ := newLib(t)
	var runs []*started
	for _, title := range []string{"One", "Two"} {
		src := checkCourse(t, loopUntilGo)
		writeFile(t, filepath.Join(src, manifest.CourseFile), courseYAML(title)+"check: check.sh\n")
		runs = append(runs, startCheck(t, l, addLocal(t, l, src).Key))
	}

	l.StopAll()

	for _, s := range runs {
		if r := s.wait(t); !r.Stopped {
			t.Errorf("check result = %+v, want stopped", r)
		}
	}
}
