package manifest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/olymp-software/chiron/internal/apperr"
)

// OverviewFile is the name of the overview file, matched case-insensitively.
const OverviewFile = "OVERVIEW.md"

// OverviewText is the text shown in the course overview: the overview file as
// Markdown, or else the course description as plain text. Both empty means no
// text.
type OverviewText struct {
	Markdown string `json:"markdown,omitempty"`
	Plain    string `json:"plain,omitempty"`
}

// ReadOverview returns the overview text of the course in root. More than one
// file matching OVERVIEW.md is a course error.
func ReadOverview(root string, c *Course) (OverviewText, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return OverviewText{}, apperr.New(apperr.CourseReadFailed, "path", ".")
	}
	var matches []string
	for _, e := range entries {
		if e.Type().IsRegular() && strings.EqualFold(e.Name(), OverviewFile) {
			matches = append(matches, e.Name())
		}
	}
	switch len(matches) {
	case 0:
		return OverviewText{Plain: c.Description}, nil
	case 1:
		b, err := os.ReadFile(filepath.Join(root, matches[0]))
		if err != nil {
			return OverviewText{}, apperr.New(apperr.CourseReadFailed, "path", matches[0])
		}
		return OverviewText{Markdown: string(b)}, nil
	default:
		return OverviewText{}, apperr.New(apperr.OverviewFileAmbiguous, "files", strings.Join(matches, ", "))
	}
}
