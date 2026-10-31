package library

import (
	"errors"
	"slices"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
	"github.com/olymp-software/chiron/internal/progress"
)

// Activity is what the activity view shows. A broken activity still has one:
// its title and progression state, and the activity error instead of the
// content.
type Activity struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	// Dir is the activity folder relative to the course root:
	// <section>/<slug>.
	Dir             string        `json:"dir"`
	Type            string        `json:"type"`
	TheoryBlocks    []TheoryBlock `json:"theoryBlocks"`
	Last            bool          `json:"last"`
	CourseCompleted bool          `json:"courseCompleted"`
	Error           *apperr.Error `json:"error,omitempty"`
}

// TheoryBlock is a theory block of an activity; markdown is the only kind.
type TheoryBlock struct {
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

// Activity reads the available activity slug of the enrolled course with key.
// A locked activity (neither completed nor current) or one not in course.yaml
// is activityUnavailable; a broken activity is not an error of the call.
func (l *Library) Activity(key, slug string) (Activity, error) {
	e, err := l.entry(key)
	if err != nil {
		return Activity{}, err
	}
	// The copy must not be replaced by Sync between course.yaml and
	// activity.yaml.
	unlock := l.lockCourse(e.Key)
	defer unlock()
	root := l.CopyPath(e.Key)
	c, err := manifest.ParseCourse(root)
	if err != nil {
		return Activity{}, err
	}
	p, err := l.progress.Load(e.Key)
	if err != nil {
		return Activity{}, err
	}
	slugs := order(c)
	p = progress.Resolve(slugs, p)
	i := slices.Index(slugs, slug)
	if i < 0 || !available(p, slug) {
		return Activity{}, apperr.New(apperr.ActivityUnavailable, "slug", slug)
	}
	section, ref := find(c, slug)
	a := Activity{
		Slug:            slug,
		Title:           ref.Title,
		Dir:             section + "/" + slug,
		Last:            i == len(slugs)-1,
		CourseCompleted: p.CourseCompleted,
	}
	parsed, err := manifest.ParseActivity(root, section, slug)
	if err != nil {
		// Even an app failure stays an activity error, so Next keeps working.
		var ae *apperr.Error
		if !errors.As(err, &ae) {
			l.log.Error("parse activity", "key", e.Key, "slug", slug, "err", err)
			ae = apperr.New(apperr.Internal)
		}
		a.Error = ae
		return a, nil
	}
	a.Type = parsed.Type
	a.TheoryBlocks = make([]TheoryBlock, len(parsed.TheoryBlocks))
	for i, b := range parsed.TheoryBlocks {
		a.TheoryBlocks[i] = TheoryBlock{Kind: b.Kind, Content: b.Data.Content}
	}
	return a, nil
}

// CompleteActivity completes the available activity slug of the enrolled
// course with key, the effect of Next and Finish course. The first activity
// that is not completed becomes current.
// It returns the activity to open next ("" after the last) and the updated
// course tree. activity.yaml is not read, so a broken activity can be
// completed.
func (l *Library) CompleteActivity(key, slug string) (string, Tree, error) {
	e, err := l.entry(key)
	if err != nil {
		return "", Tree{}, err
	}
	// The progress must not be computed from an order Sync has just replaced.
	unlock := l.lockCourse(e.Key)
	defer unlock()
	c, err := manifest.ParseCourse(l.CopyPath(e.Key))
	if err != nil {
		return "", Tree{}, err
	}
	slugs := order(c)
	i := slices.Index(slugs, slug)
	unavailable := false
	p, err := l.progress.Update(e.Key, func(stored progress.Progress) progress.Progress {
		p := progress.Resolve(slugs, stored)
		if i < 0 || !available(p, slug) {
			unavailable = true
			return stored
		}
		if !slices.Contains(p.Completed, slug) {
			p.Completed = append(slices.Clone(p.Completed), slug)
		}
		// The current activity moves to the first one not completed.
		return progress.Resolve(slugs, p)
	})
	if err != nil {
		return "", Tree{}, err
	}
	if unavailable {
		return "", Tree{}, apperr.New(apperr.ActivityUnavailable, "slug", slug)
	}
	next := ""
	if i < len(slugs)-1 {
		next = slugs[i+1]
	}
	return next, tree(c, p), nil
}

// find returns the section slug and the reference of the activity slug in c.
func find(c *manifest.Course, slug string) (string, manifest.ActivityRef) {
	for _, s := range c.Sections {
		for _, a := range s.Activities {
			if a.Slug == slug {
				return s.Slug, a
			}
		}
	}
	return "", manifest.ActivityRef{}
}
