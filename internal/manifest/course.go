// Package manifest reads course.yaml, workshop.yaml and activity.yaml. Every
// problem it finds is reported as an apperr code with the YAML path of the
// offending value.
package manifest

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"go.yaml.in/yaml/v3"

	"github.com/olymp-software/chiron/internal/apperr"
)

// CourseFile is the name of the course manifest in the course root.
const CourseFile = "course.yaml"

// WorkshopFile is the name of the workshop manifest in the workshop root.
const WorkshopFile = "workshop.yaml"

// Content types, defined by the manifest file in the root.
const (
	ContentCourse   = "course"
	ContentWorkshop = "workshop"
)

// Highest schema versions the app understands; they are independent.
const (
	courseSchemaVersion   = 1
	workshopSchemaVersion = 1
)

// Course is a valid course.yaml or workshop.yaml.
type Course struct {
	// Type is the content type, ContentCourse or ContentWorkshop.
	Type        string    `yaml:"-"`
	Version     int       `yaml:"version"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Check       string    `yaml:"check"`
	Sections    []Section `yaml:"sections"`
	// Controls run from the course overview; only workshop.yaml has them.
	Controls []Control `yaml:"controls"`
}

// Section is a section of the course manifest.
type Section struct {
	Slug        string        `yaml:"slug"`
	Title       string        `yaml:"title"`
	Description string        `yaml:"description"`
	Activities  []ActivityRef `yaml:"activities"`
}

// ActivityRef is an activity as referenced from the course manifest.
type ActivityRef struct {
	Slug  string `yaml:"slug"`
	Title string `yaml:"title"`
}

// Parse reads and validates the manifest in the root; which file is present
// defines the content type. Errors are *apperr.Error.
func Parse(root string) (*Course, error) {
	course, err := readManifest(root, CourseFile)
	if err != nil {
		return nil, err
	}
	workshop, err := readManifest(root, WorkshopFile)
	if err != nil {
		return nil, err
	}
	switch {
	case course != nil && workshop != nil:
		return nil, apperr.New(apperr.ManifestConflict)
	case course != nil:
		return parseManifest(root, course, CourseFile, ContentCourse, courseSchemaVersion, courseSchema)
	case workshop != nil:
		return parseManifest(root, workshop, WorkshopFile, ContentWorkshop, workshopSchemaVersion, workshopSchema)
	default:
		return nil, apperr.New(apperr.ManifestMissing)
	}
}

// readManifest returns the content of a manifest in the root, nil if there is
// no such file.
func readManifest(root, file string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, file))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, apperr.New(apperr.CourseReadFailed, "path", file)
	}
	return data, nil
}

func parseManifest(root string, data []byte, file, contentType string, version int,
	schema func() (*jsonschema.Schema, error)) (*Course, error) {
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, yamlSyntaxError(file, err)
	}
	if m, ok := doc.(map[string]any); ok {
		if v, ok := m["version"].(int); ok && v > version {
			return nil, unsupportedVersion(file, v)
		}
	}

	if problems, err := validateSchema(schema, doc); err != nil {
		return nil, err
	} else if len(problems) > 0 {
		return nil, invalid(file, problems)
	}

	var c Course
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, yamlSyntaxError(file, err)
	}
	c.Type = contentType
	if problems := c.check(root); len(problems) > 0 {
		return nil, invalid(file, problems)
	}
	return &c, nil
}

// invalid reports the problems of file, a course-relative path.
func invalid(file string, problems []apperr.Problem) *apperr.Error {
	e := apperr.New(apperr.ManifestInvalid, "file", file)
	e.Problems = problems
	return e
}

func unsupportedVersion(file string, v int) *apperr.Error {
	return apperr.New(apperr.UnsupportedSchemaVersion, "file", file, "version", strconv.Itoa(v))
}

var yamlLine = regexp.MustCompile(`line (\d+)`)

func yamlSyntaxError(file string, err error) *apperr.Error {
	if m := yamlLine.FindStringSubmatch(err.Error()); m != nil {
		return apperr.New(apperr.YAMLSyntax, "file", file, "line", m[1])
	}
	return apperr.New(apperr.YAMLSyntax, "file", file)
}

// validateSchema validates the decoded YAML against a schema and returns its
// problems sorted by path.
func validateSchema(compiled func() (*jsonschema.Schema, error), doc any) ([]apperr.Problem, error) {
	schema, err := compiled()
	if err != nil {
		return nil, err
	}
	// The validator expects values as decoded from JSON.
	b, err := json.Marshal(jsonCompatible(doc))
	if err != nil {
		return nil, err
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	err = schema.Validate(inst)
	var ve *jsonschema.ValidationError
	if errors.As(err, &ve) {
		var problems []apperr.Problem
		collectProblems(ve, &problems)
		slices.SortStableFunc(problems, func(a, b apperr.Problem) int {
			return cmp.Compare(a.Path, b.Path)
		})
		return problems, nil
	}
	return nil, err
}

// jsonCompatible converts what YAML allows and JSON does not: maps with
// non-string keys to maps with string keys, and .nan and .inf to strings.
func jsonCompatible(v any) any {
	switch v := v.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Sprint(v)
		}
		return v
	case map[string]any:
		for k, e := range v {
			v[k] = jsonCompatible(e)
		}
		return v
	case map[any]any:
		m := make(map[string]any, len(v))
		for k, e := range v {
			m[fmt.Sprint(k)] = jsonCompatible(e)
		}
		return m
	case []any:
		for i, e := range v {
			v[i] = jsonCompatible(e)
		}
		return v
	default:
		return v
	}
}

// yamlPath formats a JSON pointer location as a YAML path like
// sections[0].activities[1].slug.
func yamlPath(location []string) string {
	var b strings.Builder
	for _, tok := range location {
		if _, err := strconv.Atoi(tok); err == nil {
			fmt.Fprintf(&b, "[%s]", tok)
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(tok)
	}
	return b.String()
}
