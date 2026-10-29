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
	Enrolled bool                  `json:"enrolled"`
	Text     manifest.OverviewText `json:"text"`
	Error    *apperr.Error         `json:"error,omitempty"`
}

// Overview reads the course overview from the course copy. Only a missing
// course or an unreadable library is returned as an error.
func (l *Library) Overview(key string) (Overview, error) {
	e, err := l.entry(key)
	if err != nil {
		return Overview{}, err
	}
	o := Overview{Entry: e, Enrolled: l.enrolled(e.Key)}
	root := l.CopyPath(e.Key)
	c, err := manifest.ParseCourse(root)
	if err == nil {
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
