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
	Dir string `json:"dir"`
	// Type is the activity type, empty in a workshop.
	Type         string        `json:"type"`
	TheoryBlocks []TheoryBlock `json:"theoryBlocks"`
	// Quiz is set for a quiz activity that is not broken.
	Quiz *Quiz `json:"quiz,omitempty"`
	// Workshop is set for an activity of a workshop that is not broken.
	Workshop        *WorkshopActivity `json:"workshop,omitempty"`
	Last            bool              `json:"last"`
	CourseCompleted bool              `json:"courseCompleted"`
	Error           *apperr.Error     `json:"error,omitempty"`
}

// TheoryBlock is a theory block of an activity; markdown is the only kind.
type TheoryBlock struct {
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

// Quiz is the questions of a quiz activity, correct options included: the
// frontend evaluates answers, which are not protected.
type Quiz struct {
	Questions []Question `json:"questions"`
}

// Question is one quiz question; choice is the only kind, so its data is
// flattened into the question.
type Question struct {
	Kind        string   `json:"kind"`
	Prompt      string   `json:"prompt"`
	Explanation string   `json:"explanation"`
	Multiple    bool     `json:"multiple"`
	Options     []Option `json:"options"`
}

// Option is one selectable item of a choice question.
type Option struct {
	Text     string `json:"text"`
	Correct  bool   `json:"correct"`
	Feedback string `json:"feedback"`
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
	c, err := manifest.Parse(root)
	if err != nil {
		return Activity{}, err
	}
	section, title, last, completed, err := l.availableActivity(e.Key, c, slug)
	if err != nil {
		return Activity{}, err
	}
	a := Activity{
		Slug:            slug,
		Title:           title,
		Dir:             section + "/" + slug,
		Last:            last,
		CourseCompleted: completed,
	}
	if c.Type == manifest.ContentWorkshop {
		err = l.fillWorkshopActivity(e.Key, root, section, slug, &a)
	} else {
		var parsed *manifest.Activity
		parsed, err = manifest.ParseActivity(root, section, slug)
		if err == nil {
			a.Type = parsed.Type
			a.TheoryBlocks = theoryBlocks(parsed.TheoryBlocks)
			a.Quiz = quiz(parsed.Quiz)
		}
	}
	if err != nil {
		// Even an app failure stays an activity error, so Next keeps working.
		var ae *apperr.Error
		if !errors.As(err, &ae) {
			l.log.Error("parse activity", "key", e.Key, "slug", slug, "err", err)
			ae = apperr.New(apperr.Internal)
		}
		a.Error = ae
	}
	return a, nil
}

// availableActivity returns the section and title of the activity slug of c,
// whether it is the last one and whether the course is completed. A locked
// activity or one not in the manifest is activityUnavailable.
func (l *Library) availableActivity(key string, c *manifest.Course, slug string) (section, title string, last, completed bool, err error) {
	p, err := l.progress.Load(key)
	if err != nil {
		return "", "", false, false, err
	}
	slugs := order(c)
	p = progress.Resolve(slugs, p)
	i := slices.Index(slugs, slug)
	if i < 0 || !available(p, slug) {
		return "", "", false, false, apperr.New(apperr.ActivityUnavailable, "slug", slug)
	}
	section, ref := find(c, slug)
	return section, ref.Title, i == len(slugs)-1, p.CourseCompleted, nil
}

func theoryBlocks(blocks []manifest.TheoryBlock) []TheoryBlock {
	out := make([]TheoryBlock, len(blocks))
	for i, b := range blocks {
		out[i] = TheoryBlock{Kind: b.Kind, Content: b.Data.Content}
	}
	return out
}

// quiz converts the quiz data of an activity, nil for other types.
func quiz(q *manifest.Quiz) *Quiz {
	if q == nil {
		return nil
	}
	out := &Quiz{Questions: make([]Question, len(q.Questions))}
	for i, mq := range q.Questions {
		options := make([]Option, len(mq.Data.Options))
		for j, o := range mq.Data.Options {
			options[j] = Option{Text: o.Text, Correct: o.Correct, Feedback: o.Feedback}
		}
		out.Questions[i] = Question{
			Kind:        mq.Kind,
			Prompt:      mq.Prompt,
			Explanation: mq.Explanation,
			Multiple:    mq.Data.Multiple,
			Options:     options,
		}
	}
	return out
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
	c, err := manifest.Parse(l.CopyPath(e.Key))
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
	return next, l.tree(e.Key, c, p), nil
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
