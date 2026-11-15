package library

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/git"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/source"
)

const (
	tagsPrefix = "refs/tags/"
	// hooksDir is the empty folder git uses as its hooks folder.
	hooksDir = "git-hooks"
)

// fetchBatch is how many tags one git fetch names, so a repository with many
// versions stays within the Windows command-line length limit (32767
// characters; a refspec of a version is about 40).
var fetchBatch = 100

// Remote is a repository about to be added: its course source and course
// versions, highest first, or the library entry of the same course if it is
// already added.
type Remote struct {
	Source   string   `json:"source"`
	Versions []string `json:"versions"`
	Existing *Entry   `json:"existing,omitempty"`
}

// RemoteVersions lists the course versions of the repository at addr without
// downloading it. A repository already in the library returns its entry and
// git does not run.
func (l *Library) RemoteVersions(ctx context.Context, addr string) (Remote, error) {
	src, err := source.NormalizeRepo(addr)
	if err != nil {
		return Remote{}, err
	}
	return l.remoteVersions(ctx, src)
}

// remoteVersions is RemoteVersions for a normalized source; tests call it
// with file:// sources.
func (l *Library) remoteVersions(ctx context.Context, src string) (Remote, error) {
	if e, ok, err := l.findSame(src); err != nil || ok {
		if ok {
			return Remote{Source: e.Source, Versions: []string{}, Existing: &e}, nil
		}
		return Remote{}, err
	}
	tags, err := l.remoteTags(ctx, "", src)
	if err != nil {
		return Remote{}, err
	}
	versions := git.Versions(tags)
	if len(versions) == 0 {
		return Remote{}, apperr.New(apperr.NoVersions)
	}
	return Remote{Source: src, Versions: versions}, nil
}

// remoteTags returns the tag names of the repository remote ("origin" or an
// address), running git in dir.
func (l *Library) remoteTags(ctx context.Context, dir, remote string) ([]string, error) {
	out, err := l.git.Run(ctx, dir, "ls-remote", "--tags", "--refs", remote)
	if err != nil {
		return nil, err
	}
	var tags []string
	for line := range strings.Lines(string(out)) {
		_, ref, ok := strings.Cut(strings.TrimSpace(line), "\t")
		if name, isTag := strings.CutPrefix(ref, tagsPrefix); ok && isTag {
			tags = append(tags, name)
		}
	}
	return tags, nil
}

// AddRepo clones the repository source, checks out version and adds it if
// its manifest at that version is valid. A source already in the library
// returns its entry. Cancelling ctx stops git and returns ctx.Err(); a
// cancelled or failed add leaves nothing on disk and nothing in the library.
func (l *Library) AddRepo(ctx context.Context, src, version string) (Entry, error) {
	norm, err := source.NormalizeRepo(src)
	if err != nil {
		return Entry{}, err
	}
	return l.addRepo(ctx, norm, version)
}

// addRepo is AddRepo for a normalized source; tests call it with file://
// sources.
func (l *Library) addRepo(ctx context.Context, src, version string) (Entry, error) {
	if !git.IsVersion(version) {
		return Entry{}, apperr.New(apperr.VersionNotFound, "version", version)
	}
	if e, ok, err := l.findSame(src); err != nil || ok {
		return e, err
	}
	key := source.Key(src)
	unlock := l.lockCourse(key)
	defer unlock()

	newDir := filepath.Join(l.coursesPath(), key+newSuffix+strconv.FormatInt(time.Now().UnixNano(), 10))
	c, err := l.cloneAt(ctx, src, newDir, version)
	if err != nil {
		l.removeQuietly(newDir)
		return Entry{}, err
	}

	e := Entry{Origin: OriginRepo, Source: src, Key: key, Title: c.Title, Type: c.Type}
	err = l.update(func(f *libraryFile) error {
		for _, other := range f.Courses {
			if other.Key == key {
				return apperr.New(apperr.StorageKeyTaken, "key", key, "source", other.Source)
			}
		}
		// A copy left by a course added before is replaced.
		l.removeAll(l.CopyPath(key))
		if err := rename(newDir, l.CopyPath(key)); err != nil {
			l.log.Error("move new course clone", "key", key, "err", err)
			return apperr.New(apperr.CopyFailed, "path", src)
		}
		f.Courses = append(f.Courses, e)
		return nil
	})
	if err != nil {
		l.removeQuietly(newDir)
		return Entry{}, err
	}
	l.log.Info("course added", "source", src, "key", key, "version", version)
	return e, nil
}

// cloneAt clones src into dir with every course version, checks out version
// and parses the manifest there.
func (l *Library) cloneAt(ctx context.Context, src, dir, version string) (*manifest.Course, error) {
	if _, err := l.git.Run(ctx, "", "clone", "-q", "--no-checkout", "--no-tags", src, dir); err != nil {
		return nil, err
	}
	fetched, err := l.fetchTags(ctx, dir, nil)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(fetched, version) {
		return nil, apperr.New(apperr.VersionNotFound, "version", version)
	}
	if _, err := l.git.Run(ctx, dir, "checkout", "-q", "--detach", tagsPrefix+version); err != nil {
		return nil, err
	}
	return manifest.Parse(dir)
}

// fetchTags fetches the course versions of origin that are not in known into
// the clone in dir and returns them, highest first. Known tags are never named
// in a refspec, so they keep their commits even if the Author moved them.
func (l *Library) fetchTags(ctx context.Context, dir string, known []string) ([]string, error) {
	tags, err := l.remoteTags(ctx, dir, "origin")
	if err != nil {
		return nil, err
	}
	missing := slices.DeleteFunc(git.Versions(tags), func(v string) bool { return slices.Contains(known, v) })
	for batch := range slices.Chunk(missing, fetchBatch) {
		args := []string{"fetch", "-q", "--no-tags", "origin"}
		for _, v := range batch {
			args = append(args, tagsPrefix+v+":"+tagsPrefix+v)
		}
		if _, err := l.git.Run(ctx, dir, args...); err != nil {
			return nil, err
		}
	}
	return missing, nil
}

// localVersions returns the course versions in the clone in dir, highest
// first.
func (l *Library) localVersions(ctx context.Context, dir string) ([]string, error) {
	out, err := l.git.Run(ctx, dir, "for-each-ref", "--format=%(refname)", tagsPrefix)
	if err != nil {
		return nil, err
	}
	var tags []string
	for line := range strings.Lines(string(out)) {
		if name, ok := strings.CutPrefix(strings.TrimSpace(line), tagsPrefix); ok {
			tags = append(tags, name)
		}
	}
	return git.Versions(tags), nil
}

// selectedVersion returns the highest course version at the clone's HEAD, ""
// if none points at it.
func (l *Library) selectedVersion(ctx context.Context, dir string) (string, error) {
	out, err := l.git.Run(ctx, dir, "tag", "--points-at", "HEAD")
	if err != nil {
		return "", err
	}
	return git.Highest(strings.Fields(string(out))), nil
}

// versions returns the selected version and the course versions of a
// repository course from its clone.
func (l *Library) versions(ctx context.Context, key string) (selected string, all []string, err error) {
	dir := l.CopyPath(key)
	if all, err = l.localVersions(ctx, dir); err != nil {
		return "", nil, err
	}
	if selected, err = l.selectedVersion(ctx, dir); err != nil {
		return "", nil, err
	}
	return selected, all, nil
}

// outdated reports whether a repository course has a locally known version
// higher than its selected one. A failure, such as no git, is logged and
// counts as not outdated.
func (l *Library) outdated(e Entry) bool {
	selected, all, err := l.versions(context.Background(), e.Key)
	if err != nil {
		l.log.Warn("read course versions", "key", e.Key, "err", err)
		return false
	}
	return len(all) > 0 && git.CompareVersions(all[0], selected) > 0
}

// SwitchVersion checks out version in the clone of a repository course,
// discarding local changes and untracked files, after stopping the course's
// scripts. The target is not validated: the stored title and content type are
// updated only if its manifest is valid.
func (l *Library) SwitchVersion(ctx context.Context, key, version string) (Entry, error) {
	e, err := l.repoEntry(key)
	if err != nil {
		return Entry{}, err
	}
	unlock := l.lockCourse(e.Key)
	defer unlock()

	dir := l.CopyPath(e.Key)
	all, err := l.localVersions(ctx, dir)
	if err != nil {
		return Entry{}, err
	}
	if !slices.Contains(all, version) {
		return Entry{}, apperr.New(apperr.VersionNotFound, "version", version)
	}
	l.StopCheck(e.Key)
	l.runs.stopCourse(e.Key)
	if _, err := l.git.Run(ctx, dir, "checkout", "-q", "--force", "--detach", tagsPrefix+version); err != nil {
		return Entry{}, err
	}
	l.log.Info("course version switched", "source", e.Source, "key", e.Key, "version", version)
	// HEAD has moved, so the stored title, type and progress follow it even
	// if leftovers cannot be cleaned; Recover cleans them on the next start.
	_, cleanErr := l.git.Run(ctx, dir, "clean", "-q", "-ffdx")
	refreshed, err := l.refresh(e)
	if err != nil {
		return refreshed, err
	}
	return refreshed, cleanErr
}

// CheckVersions fetches the course versions of a repository course that are
// not known locally and returns them, highest first. The selected version
// does not change. Cancelling ctx stops git and returns ctx.Err().
func (l *Library) CheckVersions(ctx context.Context, key string) ([]string, error) {
	e, err := l.repoEntry(key)
	if err != nil {
		return nil, err
	}
	// The course lock is not taken: the check may wait on the network for
	// long, and the course must stay usable meanwhile. The fetch only adds
	// tags, never touches the worktree, and git locks the refs it writes, so
	// it is safe beside a version switch or a running script.
	dir := l.CopyPath(e.Key)
	known, err := l.localVersions(ctx, dir)
	if err != nil {
		return nil, err
	}
	found, err := l.fetchTags(ctx, dir, known)
	if err != nil {
		return nil, err
	}
	l.log.Info("course versions checked", "key", e.Key, "new", found)
	return found, nil
}

// repoEntry returns the library entry of a repository course.
func (l *Library) repoEntry(key string) (Entry, error) {
	e, err := l.entry(key)
	if err != nil {
		return Entry{}, err
	}
	if e.Origin != OriginRepo {
		l.log.Error("not a repository course", "key", key)
		return Entry{}, apperr.New(apperr.Internal)
	}
	return e, nil
}

// repairClones brings the clone of every repository course back to its
// checked-out commit, so an interrupted version switch leaves no files of two
// versions. Without git it is skipped.
func (l *Library) repairClones(courses []Entry) {
	ctx := context.Background()
	for _, c := range courses {
		dir := l.CopyPath(c.Key)
		if c.Origin != OriginRepo {
			continue
		}
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		_, err := l.git.Run(ctx, dir, "reset", "-q", "--hard")
		if err == nil {
			_, err = l.git.Run(ctx, dir, "clean", "-q", "-ffdx")
		}
		var ae *apperr.Error
		if errors.As(err, &ae) && ae.Code == apperr.GitNotFound {
			l.log.Warn("course clones not repaired: git not found")
			return
		}
		if err != nil {
			l.log.Error("repair course clone", "key", c.Key, "err", err)
		}
	}
}

// removeQuietly deletes a folder the learner never saw, retrying while the OS
// reports it busy.
func (l *Library) removeQuietly(path string) {
	if err := removeRetrying(path); err != nil {
		l.log.Error("remove folder", "path", path, "err", err)
	}
}
