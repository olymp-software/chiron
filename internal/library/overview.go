package library

import (
	"errors"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/internal/manifest"
)

// Overview is what the course overview shows. A broken course still has an
// overview: its stored title and the course error instead of the text.
type Overview struct {
	Entry
	Enrolled bool `json:"enrolled"`
	// Completed is the completion stored in the progress, as in the library.
	Completed bool                  `json:"completed"`
	Text      manifest.OverviewText `json:"text"`
	// Check is the environment check status: "" (no check or invalid
	// course.yaml), CheckNotChecked, CheckPassed or CheckChanged.
	Check string `json:"check"`
	// Controls are the labels of the workshop controls of an enrolled workshop.
	Controls []string      `json:"controls,omitempty"`
	Error    *apperr.Error `json:"error,omitempty"`
}

// Overview reads the course overview from the course copy. Only a missing
// course or an unreadable library is returned as an error.
func (l *Library) Overview(key string) (Overview, error) {
	e, err := l.entry(key)
	if err != nil {
		return Overview{}, err
	}
	o := Overview{Entry: e}
	o.Enrolled, o.Completed = l.enrollment(e.Key)
	root := l.CopyPath(e.Key)
	c, err := manifest.Parse(root)
	if err == nil {
		o.Check = l.checkStatus(e, c)
		if o.Enrolled {
			for _, ct := range c.Controls {
				o.Controls = append(o.Controls, ct.Label)
			}
		}
		o.Text, err = manifest.ReadOverview(root, c)
	}
	if err != nil {
		var ae *apperr.Error
		if !errors.As(err, &ae) {
			l.log.Error("read course overview", "key", e.Key, "err", err)
			ae = apperr.New(apperr.CourseReadFailed, "path", ".")
		}
		o.Error = ae
	}
	return o, nil
}
