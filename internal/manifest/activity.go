package manifest

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"go.yaml.in/yaml/v3"

	"github.com/olymp-software/chiron/internal/apperr"
)

// ActivityFile is the name of the activity file in the activity folder.
const ActivityFile = "activity.yaml"

// Highest activity.yaml schema versions the app understands; they are
// independent.
const (
	activitySchemaVersion         = 1
	workshopActivitySchemaVersion = 1
)

// Activity types.
const (
	TypeTheory = "theory"
	TypeQuiz   = "quiz"
	TypeCoding = "coding"
)

// KindMarkdown is the only theory block kind.
const KindMarkdown = "markdown"

// Activity is a valid activity.yaml.
type Activity struct {
	Version      int           `yaml:"version"`
	Type         string        `yaml:"type"`
	TheoryBlocks []TheoryBlock `yaml:"theory_blocks"`
	Controls     []Control     `yaml:"controls"`
	// Quiz is the data of a quiz activity, nil for other types.
	Quiz *Quiz `yaml:"-"`
	// Coding is the data of a coding activity, nil for other types.
	Coding *Coding `yaml:"-"`
}

// WorkshopActivity is a valid activity.yaml of a workshop.
type WorkshopActivity struct {
	Version      int             `yaml:"version"`
	TheoryBlocks []TheoryBlock   `yaml:"theory_blocks"`
	Solution     string          `yaml:"solution"`
	Check        string          `yaml:"check"`
	Controls     []Control       `yaml:"controls"`
	Files        []WorkspaceItem `yaml:"files"`
}

// TheoryBlock is one item of theory_blocks.
type TheoryBlock struct {
	Kind string       `yaml:"kind"`
	Data MarkdownData `yaml:"data"`
}

// MarkdownData is the data of a markdown theory block.
type MarkdownData struct {
	Content string `yaml:"content"`
}

// Control is a button that runs a script from the activity folder.
type Control struct {
	Label       string `yaml:"label"`
	Script      string `yaml:"script"`
	Concurrency string `yaml:"concurrency"`
}

// Quiz is the data of a quiz activity.
type Quiz struct {
	Questions []Question `yaml:"questions"`
}

// Question is one quiz question; choice is the only kind.
type Question struct {
	Kind        string `yaml:"kind"`
	Prompt      string `yaml:"prompt"`
	Data        Choice `yaml:"data"`
	Explanation string `yaml:"explanation"`
}

// Choice is the data of a choice question.
type Choice struct {
	Multiple bool     `yaml:"multiple"`
	Options  []Option `yaml:"options"`
}

// Option is one selectable item of a choice question.
type Option struct {
	Text     string `yaml:"text"`
	Correct  bool   `yaml:"correct"`
	Feedback string `yaml:"feedback"`
}

// Coding is the data of a coding activity.
type Coding struct {
	Workspace []WorkspaceItem `yaml:"workspace"`
}

// WorkspaceItem is a file or folder copied to the learner folder.
type WorkspaceItem struct {
	Path string `yaml:"path"`
}

// ActivityPath returns the course-relative path of the activity file of the
// activity slug in section.
func ActivityPath(section, slug string) string {
	return path.Join(section, slug, ActivityFile)
}

// ParseActivity reads and validates the activity.yaml of the activity slug in
// section. Errors are *apperr.Error; the whole-file ones carry the
// course-relative file path.
func ParseActivity(courseRoot, section, slug string) (*Activity, error) {
	file := ActivityPath(section, slug)
	dir := filepath.Join(courseRoot, section, slug)
	data, doc, err := readActivity(courseRoot, file)
	if err != nil {
		return nil, err
	}
	if err := precheckActivity(file, doc); err != nil {
		return nil, err
	}

	if problems, err := validateSchema(courseActivitySchema, doc); err != nil {
		return nil, err
	} else if len(problems) > 0 {
		return nil, invalid(file, problems)
	}

	var raw struct {
		Activity `yaml:",inline"`
		Data     yaml.Node `yaml:"data"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, yamlSyntaxError(file, err)
	}
	a := raw.Activity
	switch a.Type {
	case TypeQuiz:
		a.Quiz = &Quiz{}
		err = raw.Data.Decode(a.Quiz)
	case TypeCoding:
		a.Coding = &Coding{}
		err = raw.Data.Decode(a.Coding)
	}
	if err != nil {
		return nil, yamlSyntaxError(file, err)
	}
	if problems := a.check(dir); len(problems) > 0 {
		return nil, invalid(file, problems)
	}
	return &a, nil
}

// ParseWorkshopActivity reads and validates the activity.yaml of the activity
// slug in section of a workshop. Errors are as in ParseActivity.
func ParseWorkshopActivity(root, section, slug string) (*WorkshopActivity, error) {
	file := ActivityPath(section, slug)
	data, doc, err := readActivity(root, file)
	if err != nil {
		return nil, err
	}
	if err := precheckWorkshopActivity(file, doc); err != nil {
		return nil, err
	}

	if problems, err := validateSchema(workshopActivitySchema, doc); err != nil {
		return nil, err
	} else if len(problems) > 0 {
		return nil, invalid(file, problems)
	}

	var a WorkshopActivity
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil, yamlSyntaxError(file, err)
	}
	if problems := a.check(filepath.Join(root, section, slug)); len(problems) > 0 {
		return nil, invalid(file, problems)
	}
	return &a, nil
}

// readActivity returns the content of the activity file at the course-relative
// path file and its decoded YAML.
func readActivity(root, file string) ([]byte, any, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, apperr.New(apperr.ActivityFileMissing, "file", file)
	}
	if err != nil {
		return nil, nil, apperr.New(apperr.CourseReadFailed, "path", file)
	}
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, nil, yamlSyntaxError(file, err)
	}
	return data, doc, nil
}

// precheckActivity reports what requires a newer app version: a higher
// schema version, an unknown type or an unknown theory block kind. Values of
// the wrong type are left to the schema.
func precheckActivity(file string, doc any) error {
	m, ok := doc.(map[string]any)
	if !ok {
		return nil
	}
	if v, ok := m["version"].(int); ok && v > activitySchemaVersion {
		return unsupportedVersion(file, v)
	}
	if t, ok := m["type"].(string); ok && t != TypeTheory && t != TypeQuiz && t != TypeCoding {
		return apperr.New(apperr.UnsupportedActivityType, "file", file, "type", t)
	}
	return precheckBlocks(file, m)
}

// precheckWorkshopActivity is precheckActivity for a workshop, where any type
// is left to the schema, which forbids the field.
func precheckWorkshopActivity(file string, doc any) error {
	m, ok := doc.(map[string]any)
	if !ok {
		return nil
	}
	if v, ok := m["version"].(int); ok && v > workshopActivitySchemaVersion {
		return unsupportedVersion(file, v)
	}
	return precheckBlocks(file, m)
}

// precheckBlocks reports an unknown theory block kind.
func precheckBlocks(file string, m map[string]any) error {
	blocks, _ := m["theory_blocks"].([]any)
	for _, b := range blocks {
		block, _ := b.(map[string]any)
		if k, ok := block["kind"].(string); ok && k != KindMarkdown {
			return apperr.New(apperr.UnsupportedBlockKind, "file", file, "kind", k)
		}
	}
	return nil
}
