package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olymp-software/chiron/internal/apperr"
)

// check returns the problems the schema cannot express: slug uniqueness and
// the check and control paths in the course root.
func (c *Course) check(root string) []apperr.Problem {
	var problems []apperr.Problem
	sections := map[string]bool{}
	activities := map[string]bool{}
	for i, s := range c.Sections {
		path := fmt.Sprintf("sections[%d].slug", i)
		if sections[s.Slug] {
			problems = append(problems, apperr.NewProblem(apperr.DuplicateSectionSlug, path, "slug", s.Slug))
		}
		sections[s.Slug] = true
		for j, a := range s.Activities {
			path := fmt.Sprintf("sections[%d].activities[%d].slug", i, j)
			if activities[a.Slug] {
				problems = append(problems, apperr.NewProblem(apperr.DuplicateActivitySlug, path, "slug", a.Slug))
			}
			activities[a.Slug] = true
		}
	}
	if c.Check != "" {
		if p, ok := checkPath(root, "check", c.Check, false); !ok {
			problems = append(problems, p)
		}
	}
	return append(problems, checkControls(root, c.Controls)...)
}

// checkControls returns the problems of control scripts, paths in base.
func checkControls(base string, controls []Control) []apperr.Problem {
	var problems []apperr.Problem
	for i, c := range controls {
		if p, ok := checkPath(base, fmt.Sprintf("controls[%d].script", i), c.Script, false); !ok {
			problems = append(problems, p)
		}
	}
	return problems
}

// check returns the problems the schema cannot express: paths inside the
// activity folder dir.
func (a *WorkshopActivity) check(dir string) []apperr.Problem {
	var problems []apperr.Problem
	if a.Check != "" {
		if p, ok := checkPath(dir, "check", a.Check, false); !ok {
			problems = append(problems, p)
		}
	}
	problems = append(problems, checkControls(dir, a.Controls)...)
	for i, f := range a.Files {
		if p, ok := checkPath(dir, fmt.Sprintf("files[%d].path", i), f.Path, true); !ok {
			problems = append(problems, p)
		}
	}
	return problems
}

// check returns the problems the schema cannot express: paths inside the
// activity folder dir and the number of correct options.
func (a *Activity) check(dir string) []apperr.Problem {
	problems := checkControls(dir, a.Controls)
	if a.Coding != nil {
		for i, w := range a.Coding.Workspace {
			if p, ok := checkPath(dir, fmt.Sprintf("data.workspace[%d].path", i), w.Path, true); !ok {
				problems = append(problems, p)
			}
		}
	}
	if a.Quiz != nil {
		for i, q := range a.Quiz.Questions {
			correct := 0
			for _, o := range q.Data.Options {
				if o.Correct {
					correct++
				}
			}
			path := fmt.Sprintf("data.questions[%d].data.options", i)
			switch {
			case !q.Data.Multiple && correct != 1:
				problems = append(problems, apperr.NewProblem(apperr.OneCorrectOptionRequired, path))
			case q.Data.Multiple && correct == 0:
				problems = append(problems, apperr.NewProblem(apperr.CorrectOptionRequired, path))
			}
		}
	}
	return problems
}

// checkPath checks a path from a course file: relative to base, / separated,
// no .., and an existing file in base, or a folder if dirOK.
func checkPath(base, field, path string, dirOK bool) (apperr.Problem, bool) {
	if !validRelPath(path) {
		return apperr.NewProblem(apperr.PathInvalid, field, "path", path), false
	}
	info, err := os.Stat(filepath.Join(base, filepath.FromSlash(path)))
	if err == nil && (info.Mode().IsRegular() || dirOK && info.IsDir()) {
		return apperr.Problem{}, true
	}
	return apperr.NewProblem(apperr.PathNotFound, field, "path", path), false
}

func validRelPath(path string) bool {
	if strings.Contains(path, `\`) || strings.HasPrefix(path, "/") || filepath.VolumeName(path) != "" {
		return false
	}
	// The base folder itself is not a path in it.
	if filepath.Clean(path) == "." {
		return false
	}
	// A drive letter is not a volume name on Linux, but still absolute on Windows.
	if len(path) >= 2 && path[1] == ':' {
		return false
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == ".." {
			return false
		}
	}
	return true
}
