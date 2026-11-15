package library

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/git/gittest"
	"github.com/olymp-software/chiron/internal/manifest"
)

// twoActivitiesYAML is a course.yaml with the activities of courseYAML plus a
// new one, bye.
func twoActivitiesYAML(title string) string {
	return "version: 1\ntitle: " + title + "\nsections:\n" +
		"  - {slug: basics, title: Basics, activities: [{slug: hello, title: Hello}, {slug: bye, title: Bye}]}\n"
}

// folderFiles returns the files under dir as slash paths with their content.
func folderFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = readFile(t, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// testRepo is a course repository:
//   - v1.0.0: courseYAML("Go v1") with one activity;
//   - v1.1.0 and v1.1.1 on one commit: twoActivitiesYAML("Go v2");
//   - v1.2.0 on a commit no branch reaches: an invalid course.yaml;
//   - release-3 and v2.0.0-beta.1, which are not course versions.
type testRepo struct {
	*gittest.Repo
	v1, v2, broken string
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()
	r := &testRepo{Repo: gittest.New(t)}
	r.v1 = r.Commit(map[string]string{manifest.CourseFile: courseYAML("Go v1")})
	r.Tag("v1.0.0", r.v1)
	r.v2 = r.Commit(map[string]string{manifest.CourseFile: twoActivitiesYAML("Go v2")})
	r.Tag("v1.1.0", r.v2)
	r.Tag("v1.1.1", r.v2)
	r.broken = r.CommitDetached(map[string]string{manifest.CourseFile: "version: [\n"})
	r.Tag("v1.2.0", r.broken)
	r.Tag("release-3", r.v1)
	r.Tag("v2.0.0-beta.1", r.v2)
	return r
}

// repoWithFiles creates a repository whose only version v1.0.0 holds files.
func repoWithFiles(t *testing.T, files map[string]string) *gittest.Repo {
	t.Helper()
	r := gittest.New(t)
	r.Tag("v1.0.0", r.Commit(files))
	return r
}

func addRepo(t *testing.T, l *Library, url, version string) Entry {
	t.Helper()
	e, err := l.addRepo(context.Background(), url, version)
	if err != nil {
		t.Fatalf("addRepo %s: %v", version, err)
	}
	return e
}

func overview(t *testing.T, l *Library, key string) Overview {
	t.Helper()
	o, err := l.Overview(key)
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	return o
}

func courseOf(t *testing.T, l *Library, key string) Course {
	t.Helper()
	courses, err := l.Courses()
	if err != nil {
		t.Fatalf("Courses: %v", err)
	}
	for _, c := range courses {
		if c.Key == key {
			return c
		}
	}
	t.Fatalf("course %s not listed", key)
	return Course{}
}

func switchVersion(t *testing.T, l *Library, key, version string) Entry {
	t.Helper()
	e, err := l.SwitchVersion(context.Background(), key, version)
	if err != nil {
		t.Fatalf("SwitchVersion %s: %v", version, err)
	}
	return e
}

// noGit hides git from the library for the rest of the test.
func noGit(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
}

// tagCommit returns the commit a tag points at in the clone in dir.
func tagCommit(t *testing.T, r *gittest.Repo, dir, tag string) string {
	t.Helper()
	return strings.TrimSpace(r.Git(dir, "rev-parse", tag+"^{commit}"))
}

func wantNothingAdded(t *testing.T, l *Library, dataDir string) {
	t.Helper()
	if names := courseEntries(t, dataDir); len(names) != 0 {
		t.Errorf("courses folder = %v, want empty", names)
	}
	if courses, err := l.List(); err != nil || len(courses) != 0 {
		t.Errorf("library = %v, %v, want empty", courses, err)
	}
}

// 5.2 Remote versions

func TestRemoteVersions(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	got, err := l.remoteVersions(context.Background(), r.URL)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"v1.2.0", "v1.1.1", "v1.1.0", "v1.0.0"}
	if got.Source != r.URL || !slices.Equal(got.Versions, want) || got.Existing != nil {
		t.Errorf("remoteVersions = %+v, want source %s and versions %v", got, r.URL, want)
	}
}

func TestRemoteVersionsExistingRunsNoGit(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	noGit(t)
	got, err := l.remoteVersions(context.Background(), r.URL)
	if err != nil {
		t.Fatal(err)
	}
	if got.Existing == nil || got.Existing.Key != e.Key {
		t.Errorf("remoteVersions = %+v, want the existing entry", got)
	}
}

func TestRemoteVersionsNoTags(t *testing.T) {
	r := gittest.New(t)
	r.Tag("release-3", r.Commit(map[string]string{manifest.CourseFile: courseYAML("Go")}))
	l, _ := newLib(t)
	_, err := l.remoteVersions(context.Background(), r.URL)
	wantCode(t, err, apperr.NoVersions)
}

func TestRemoteVersionsMissingRepository(t *testing.T) {
	l, _ := newLib(t)
	_, err := l.remoteVersions(context.Background(), gittest.FileURL(filepath.Join(t.TempDir(), "missing.git")))
	wantCode(t, err, apperr.RepoNotFound)
}

func TestRemoteVersionsInvalidAddress(t *testing.T) {
	l, _ := newLib(t)
	noGit(t)
	_, err := l.RemoteVersions(context.Background(), "http://example.com/course.git")
	wantCode(t, err, apperr.RepoAddressInvalid)
}

func TestRemoteVersionsNoGit(t *testing.T) {
	l, _ := newLib(t)
	noGit(t)
	_, err := l.RemoteVersions(context.Background(), "https://example.com/team/course")
	wantCode(t, err, apperr.GitNotFound)
}

// 5.3 Fetching tags

func TestFetchKeepsKnownTags(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	// The Author moves v1.0.0 and releases v1.3.0.
	r.Tag("v1.0.0", r.v2)
	r.Tag("v1.3.0", r.v2)

	found, err := l.CheckVersions(context.Background(), e.Key)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(found, []string{"v1.3.0"}) {
		t.Errorf("found = %v, want [v1.3.0]", found)
	}
	dir := l.CopyPath(e.Key)
	if got := tagCommit(t, r.Repo, dir, "v1.0.0"); got != r.v1 {
		t.Errorf("v1.0.0 = %s, want the commit it had before (%s)", got, r.v1)
	}
	if got := tagCommit(t, r.Repo, dir, "v1.3.0"); got != r.v2 {
		t.Errorf("v1.3.0 = %s, want %s", got, r.v2)
	}
	if o := overview(t, l, e.Key); o.Version != "v1.0.0" || o.Title != "Go v1" {
		t.Errorf("overview version %q title %q, want v1.0.0 and the old title", o.Version, o.Title)
	}
}

// 5.4 Adding a repository course

func TestAddRepoCourse(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")

	want := Entry{Origin: OriginRepo, Source: r.URL, Key: e.Key, Title: "Go v1", Type: manifest.ContentCourse}
	if e != want {
		t.Errorf("entry = %+v, want %+v", e, want)
	}
	if courses, _ := l.List(); len(courses) != 1 || courses[0] != want {
		t.Errorf("library = %+v, want the entry", courses)
	}
	// The clone has every version, also the one no branch reaches.
	o := overview(t, l, e.Key)
	if o.Version != "v1.0.0" || !slices.Equal(o.Versions, []string{"v1.2.0", "v1.1.1", "v1.1.0", "v1.0.0"}) {
		t.Errorf("version %q, versions %v", o.Version, o.Versions)
	}
	if o.Error != nil {
		t.Errorf("overview error = %v", o.Error)
	}
}

func TestAddRepoSameSourceTwice(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	first := addRepo(t, l, r.URL, "v1.0.0")
	second := addRepo(t, l, r.URL, "v1.1.0")
	if second != first {
		t.Errorf("second add = %+v, want the existing entry %+v", second, first)
	}
	if courses, _ := l.List(); len(courses) != 1 {
		t.Errorf("library has %d entries, want 1", len(courses))
	}
}

func TestAddRepoWorkshop(t *testing.T) {
	r := repoWithFiles(t, folderFiles(t, workshopSource(t, "", "", "", nil)))
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	if e.Type != manifest.ContentWorkshop || e.Title != "Rollouts" {
		t.Errorf("entry = %+v, want the workshop Rollouts", e)
	}
}

func TestAddRepoInvalidManifest(t *testing.T) {
	r := newTestRepo(t)
	l, dataDir := newLib(t)
	_, err := l.addRepo(context.Background(), r.URL, "v1.2.0")
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("err = %v, want the course error", err)
	}
	wantNothingAdded(t, l, dataDir)
}

func TestAddRepoUnknownVersion(t *testing.T) {
	r := newTestRepo(t)
	l, dataDir := newLib(t)
	for _, v := range []string{"v9.9.9", "release-3", "v2.0.0-beta.1"} {
		_, err := l.addRepo(context.Background(), r.URL, v)
		wantCode(t, err, apperr.VersionNotFound)
	}
	wantNothingAdded(t, l, dataDir)
}

func TestAddRepoCancelled(t *testing.T) {
	r := newTestRepo(t)
	l, dataDir := newLib(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := l.addRepo(ctx, r.URL, "v1.0.0"); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	wantNothingAdded(t, l, dataDir)
}

func TestRecoverDeletesInterruptedClone(t *testing.T) {
	r := newTestRepo(t)
	_, dataDir := newLib(t)
	leftover := filepath.Join(dataDir, coursesDir, "repo-0123abcd"+newSuffix+"42")
	r.Git("", "clone", "-q", "--no-checkout", r.URL, leftover)

	l := openLib(t, dataDir)

	wantNothingAdded(t, l, dataDir)
}

func TestAddRepoKeepsLineEndings(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	writeFile(t, cfg, "[core]\n\tautocrlf = true\n")
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)
	r := repoWithFiles(t, map[string]string{
		manifest.CourseFile: courseYAML("Go"),
		"check.sh":          "echo one\necho two\n",
	})
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	if got := readFile(t, filepath.Join(l.CopyPath(e.Key), "check.sh")); got != "echo one\necho two\n" {
		t.Errorf("check.sh = %q, want LF line endings", got)
	}
}

// 6.1 Versions

func TestSelectedVersionIsHighestAtHead(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.1.0")
	if o := overview(t, l, e.Key); o.Version != "v1.1.1" {
		t.Errorf("version = %q, want v1.1.1", o.Version)
	}
}

func TestOutdated(t *testing.T) {
	r := repoWithFiles(t, map[string]string{manifest.CourseFile: courseYAML("Go")})
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	if courseOf(t, l, e.Key).Outdated {
		t.Error("outdated on the only version")
	}

	// Two tags on the selected commit: the higher one is selected.
	r.Tag("v1.0.1", r.Head())
	if _, err := l.CheckVersions(context.Background(), e.Key); err != nil {
		t.Fatal(err)
	}
	if courseOf(t, l, e.Key).Outdated {
		t.Error("outdated because of a second tag on the selected commit")
	}

	// A newer tag the app has not fetched yet does not count.
	r.Tag("v1.1.0", r.Commit(map[string]string{"notes.md": "new"}))
	if courseOf(t, l, e.Key).Outdated {
		t.Error("outdated before the newer tag was fetched")
	}
	if _, err := l.CheckVersions(context.Background(), e.Key); err != nil {
		t.Fatal(err)
	}
	if !courseOf(t, l, e.Key).Outdated {
		t.Error("not outdated with a newer local tag")
	}
	switchVersion(t, l, e.Key, "v1.1.0")
	if courseOf(t, l, e.Key).Outdated {
		t.Error("outdated on the latest version")
	}
}

func TestVersionsWithoutGit(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	noGit(t)
	if courseOf(t, l, e.Key).Outdated {
		t.Error("outdated without git")
	}
	o := overview(t, l, e.Key)
	if o.Error != nil || o.Title != "Go v1" || o.Version != "" || len(o.Versions) != 0 {
		t.Errorf("overview = %+v, want it loaded without versions", o)
	}
}

func TestLocalFolderCourseNeedsNoGit(t *testing.T) {
	l, _ := newLib(t)
	noGit(t)
	e := addLocal(t, l, courseFolder(t, "Go"))
	if c := courseOf(t, l, e.Key); c.Outdated {
		t.Error("local-folder course outdated")
	}
	o := overview(t, l, e.Key)
	if o.Error != nil || o.Version != "" || o.Versions == nil || len(o.Versions) != 0 {
		t.Errorf("overview = %+v, want no versions", o)
	}
	syncCourse(t, l, e.Key)
}

// 6.2 Version switch

func TestSwitchVersion(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")

	got := switchVersion(t, l, e.Key, "v1.1.0")

	if got.Title != "Go v2" {
		t.Errorf("title = %q, want Go v2", got.Title)
	}
	if o := overview(t, l, e.Key); o.Version != "v1.1.1" || o.Title != "Go v2" {
		t.Errorf("overview version %q title %q", o.Version, o.Title)
	}
	// Back to an older version.
	switchVersion(t, l, e.Key, "v1.0.0")
	if o := overview(t, l, e.Key); o.Version != "v1.0.0" || o.Title != "Go v1" {
		t.Errorf("overview after switching back: version %q title %q", o.Version, o.Title)
	}
}

func TestSwitchVersionDiscardsLocalChanges(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	dir := l.CopyPath(e.Key)
	writeFile(t, filepath.Join(dir, manifest.CourseFile), "edited")
	writeFile(t, filepath.Join(dir, "untracked", "file.txt"), "x")

	switchVersion(t, l, e.Key, "v1.1.0")

	if got := readFile(t, filepath.Join(dir, manifest.CourseFile)); got != twoActivitiesYAML("Go v2") {
		t.Errorf("course.yaml = %q, want the committed one", got)
	}
	if exists(filepath.Join(dir, "untracked")) {
		t.Error("untracked files survived")
	}
}

func TestSwitchVersionBrokenManifest(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")

	got := switchVersion(t, l, e.Key, "v1.2.0")

	if got.Title != "Go v1" {
		t.Errorf("title = %q, want the old one", got.Title)
	}
	o := overview(t, l, e.Key)
	if o.Error == nil || o.Version != "v1.2.0" || len(o.Versions) != 4 {
		t.Errorf("overview = %+v, want a course error at v1.2.0 with the version list", o)
	}
	switchVersion(t, l, e.Key, "v1.0.0")
	if o := overview(t, l, e.Key); o.Error != nil || o.Version != "v1.0.0" {
		t.Errorf("overview after switching back = %+v", o)
	}
}

func TestSwitchVersionUnknown(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	_, err := l.SwitchVersion(context.Background(), e.Key, "v9.0.0")
	wantCode(t, err, apperr.VersionNotFound)
}

func TestSwitchVersionOffline(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	if err := os.RemoveAll(r.Bare); err != nil {
		t.Fatal(err)
	}
	switchVersion(t, l, e.Key, "v1.1.0")
}

func TestSwitchVersionUncompletesCourse(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	if err := l.Enroll(e.Key); err != nil {
		t.Fatal(err)
	}
	completeActivity(t, l, e.Key, "hello")
	if !courseOf(t, l, e.Key).Completed {
		t.Fatal("course not completed")
	}

	switchVersion(t, l, e.Key, "v1.1.0")

	if c := courseOf(t, l, e.Key); c.Completed || !c.Enrolled {
		t.Errorf("course = %+v, want enrolled and not completed", c)
	}
}

func TestSwitchVersionStopsCheck(t *testing.T) {
	src := checkCourse(t, "while true; do :; done\n")
	r := repoWithFiles(t, folderFiles(t, src))
	r.Tag("v1.1.0", r.Head())
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	s := startCheck(t, l, e.Key)

	switchVersion(t, l, e.Key, "v1.1.0")

	if res := s.wait(t); !res.Stopped {
		t.Errorf("check result = %+v, want stopped", res)
	}
}

func TestSwitchVersionStopsControls(t *testing.T) {
	r := repoWithFiles(t, folderFiles(t, workshopSource(t, "", twoControls, "", twoControlFiles())))
	r.Tag("v1.1.0", r.Head())
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	if err := l.Enroll(e.Key); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(l.WorkshopFolder(e.Key), "mine"), "kept")
	s := startControl(t, l, e.Key, "first", 0)

	switchVersion(t, l, e.Key, "v1.1.0")

	if res := s.wait(t); !res.Stopped {
		t.Errorf("control result = %+v, want stopped", res)
	}
	if got := readFile(t, filepath.Join(l.WorkshopFolder(e.Key), "mine")); got != "kept" {
		t.Errorf("workshop folder file = %q, want it untouched", got)
	}
}

// 6.3 Check for new versions

func TestCheckVersions(t *testing.T) {
	r := repoWithFiles(t, map[string]string{manifest.CourseFile: courseYAML("Go")})
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	r.Tag("v1.1.0", r.Commit(map[string]string{"notes.md": "new"}))

	found, err := l.CheckVersions(context.Background(), e.Key)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(found, []string{"v1.1.0"}) {
		t.Errorf("found = %v, want [v1.1.0]", found)
	}
	o := overview(t, l, e.Key)
	if o.Version != "v1.0.0" || !slices.Equal(o.Versions, []string{"v1.1.0", "v1.0.0"}) {
		t.Errorf("version %q versions %v, want v1.0.0 selected and v1.1.0 listed", o.Version, o.Versions)
	}

	found, err = l.CheckVersions(context.Background(), e.Key)
	if err != nil || len(found) != 0 {
		t.Errorf("second check = %v, %v, want nothing new", found, err)
	}
}

func TestCheckVersionsUnreachable(t *testing.T) {
	r := repoWithFiles(t, map[string]string{manifest.CourseFile: courseYAML("Go")})
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	if err := os.RemoveAll(r.Bare); err != nil {
		t.Fatal(err)
	}

	_, err := l.CheckVersions(context.Background(), e.Key)

	wantCode(t, err, apperr.RepoNotFound)
	if o := overview(t, l, e.Key); o.Version != "v1.0.0" || !slices.Equal(o.Versions, []string{"v1.0.0"}) {
		t.Errorf("version %q versions %v, want them unchanged", o.Version, o.Versions)
	}
}

func TestCheckVersionsLocalFolderCourse(t *testing.T) {
	l, _ := newLib(t)
	e := addLocal(t, l, courseFolder(t, "Go"))
	_, err := l.CheckVersions(context.Background(), e.Key)
	wantCode(t, err, apperr.Internal)
}

// 6.4 Repair on startup

func TestRecoverRepairsClone(t *testing.T) {
	r := newTestRepo(t)
	l, dataDir := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	dir := l.CopyPath(e.Key)
	writeFile(t, filepath.Join(dir, manifest.CourseFile), "edited")
	writeFile(t, filepath.Join(dir, "untracked.txt"), "x")

	l = openLib(t, dataDir)

	if got := readFile(t, filepath.Join(dir, manifest.CourseFile)); got != courseYAML("Go v1") {
		t.Errorf("course.yaml = %q, want the committed one", got)
	}
	if exists(filepath.Join(dir, "untracked.txt")) {
		t.Error("untracked file survived")
	}
	if o := overview(t, l, e.Key); o.Version != "v1.0.0" {
		t.Errorf("version = %q, want v1.0.0", o.Version)
	}
}

func TestRecoverWithoutGit(t *testing.T) {
	r := newTestRepo(t)
	l, dataDir := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	writeFile(t, filepath.Join(l.CopyPath(e.Key), "untracked.txt"), "x")
	orphan := filepath.Join(dataDir, coursesDir, "orphan-0123abcd")
	writeFile(t, filepath.Join(orphan, "file"), "x")
	noGit(t)

	openLib(t, dataDir)

	if exists(orphan) {
		t.Error("orphan copy not deleted")
	}
	if !exists(filepath.Join(l.CopyPath(e.Key), "untracked.txt")) {
		t.Error("clone changed without git")
	}
}

// 6.5 Removal

func TestRemoveRepoCourse(t *testing.T) {
	r := newTestRepo(t)
	l, _ := newLib(t)
	e := addRepo(t, l, r.URL, "v1.0.0")
	// git writes pack files read-only.
	r.Git(l.CopyPath(e.Key), "gc", "-q")

	remove(t, l, e.Key, false, false)

	wantNotInLibrary(t, l, e.Key)
	if exists(l.CopyPath(e.Key)) {
		t.Error("clone survived")
	}
}
