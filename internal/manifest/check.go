package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olymp-software/chiron/internal/apperr"
)

// check returns the problems the schema cannot express: slug uniqueness and
// the check path.
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
		if p, ok := checkPath(root, "check", c.Check); !ok {
			problems = append(problems, p)
		}
	}
	return problems
}

// checkPath checks a path from a course file: relative, / separated, no ..,
// and an existing file in the course root.
func checkPath(root, field, path string) (apperr.Problem, bool) {
	if !validRelPath(path) {
		return apperr.NewProblem(apperr.PathInvalid, field, "path", path), false
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil || !info.Mode().IsRegular() {
		return apperr.NewProblem(apperr.PathNotFound, field, "path", path), false
	}
	return apperr.Problem{}, true
}

func validRelPath(path string) bool {
	if strings.Contains(path, `\`) || strings.HasPrefix(path, "/") || filepath.VolumeName(path) != "" {
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
