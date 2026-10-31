package library

import (
	"slices"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/progress"
)

// Tree is what the course tree shows: the sections and activities of
// course.yaml, the current activity and whether the course is completed.
type Tree struct {
	Title     string        `json:"title"`
	Current   string        `json:"current"`
	Completed bool          `json:"completed"`
	Sections  []TreeSection `json:"sections"`
}

// TreeSection is a section of the course tree.
type TreeSection struct {
	Slug       string         `json:"slug"`
	Title      string         `json:"title"`
	Activities []TreeActivity `json:"activities"`
}

// TreeActivity is an activity of the course tree. A locked activity is
// neither completed nor current and cannot be opened.
type TreeActivity struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Locked    bool   `json:"locked"`
}

// Enroll enrolls the course with key with its first activity as the current
// activity. course.yaml must be valid and the environment check, if declared,
// satisfied. An enrolled course keeps its progress.
func (l *Library) Enroll(key string) error {
	e, c, err := l.course(key)
	if err != nil {
		return err
	}
	if !l.checkSatisfied(e, c) {
		return apperr.New(apperr.CheckRequired)
	}
	if err := l.progress.Enroll(key, order(c)[0]); err != nil {
		return err
	}
	l.log.Info("course enrolled", "key", key)
	return nil
}

// Enter decides whether the enrolled course with key may be entered and
// returns its current activity, moved into course.yaml if the activity is
// gone.
func (l *Library) Enter(key string) (string, error) {
	if _, err := l.entry(key); err != nil {
		return "", err
	}
	if !l.enrolled(key) {
		return "", apperr.New(apperr.NotEnrolled)
	}
	e, c, err := l.course(key)
	if err != nil {
		return "", err
	}
	if !l.checkSatisfied(e, c) {
		return "", apperr.New(apperr.CheckRequired)
	}
	slugs := order(c)
	p, err := l.progress.Update(key, func(p progress.Progress) progress.Progress {
		return progress.Resolve(slugs, p)
	})
	if err != nil {
		return "", err
	}
	return p.Current, nil
}

// Tree returns the course tree of the enrolled course with key.
func (l *Library) Tree(key string) (Tree, error) {
	_, c, err := l.course(key)
	if err != nil {
		return Tree{}, err
	}
	p, err := l.progress.Load(key)
	if err != nil {
		return Tree{}, err
	}
	return tree(c, p), nil
}

// tree builds the course tree of c with progress p.
func tree(c *manifest.Course, p progress.Progress) Tree {
	p = progress.Resolve(order(c), p)
	t := Tree{
		Title:     c.Title,
		Current:   p.Current,
		Completed: p.CourseCompleted,
		Sections:  make([]TreeSection, len(c.Sections)),
	}
	for si, s := range c.Sections {
		ts := TreeSection{Slug: s.Slug, Title: s.Title, Activities: make([]TreeActivity, len(s.Activities))}
		for j, a := range s.Activities {
			ts.Activities[j] = TreeActivity{
				Slug:      a.Slug,
				Title:     a.Title,
				Completed: slices.Contains(p.Completed, a.Slug),
				Locked:    !available(p, a.Slug),
			}
		}
		t.Sections[si] = ts
	}
	return t
}

// available reports whether the activity slug may be opened with resolved
// progress p: it is completed or current.
func available(p progress.Progress, slug string) bool {
	return slug == p.Current || slices.Contains(p.Completed, slug)
}

// course returns the library entry of the course with key and its
// course.yaml.
func (l *Library) course(key string) (Entry, *manifest.Course, error) {
	e, err := l.entry(key)
	if err != nil {
		return Entry{}, nil, err
	}
	unlock := l.lockCourse(e.Key)
	defer unlock()
	c, err := manifest.ParseCourse(l.CopyPath(e.Key))
	return e, c, err
}

// order returns the activity slugs of c in course order.
func order(c *manifest.Course) []string {
	var slugs []string
	for _, s := range c.Sections {
		for _, a := range s.Activities {
			slugs = append(slugs, a.Slug)
		}
	}
	return slugs
}
