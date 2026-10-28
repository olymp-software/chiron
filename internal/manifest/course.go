// Package manifest reads course.yaml. Every problem it finds is reported as an
// apperr code with the YAML path of the offending value.
package manifest

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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

// courseSchemaVersion is the highest course.yaml schema version the app
// understands.
const courseSchemaVersion = 1

// Course is a valid course.yaml.
type Course struct {
	Version     int       `yaml:"version"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Check       string    `yaml:"check"`
	Sections    []Section `yaml:"sections"`
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

// ParseCourse reads and validates course.yaml in the course root. Errors are
// *apperr.Error.
func ParseCourse(root string) (*Course, error) {
	data, err := os.ReadFile(filepath.Join(root, CourseFile))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, apperr.New(apperr.ManifestMissing)
	}
	if err != nil {
		return nil, apperr.New(apperr.CourseReadFailed, "path", CourseFile)
	}

	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, yamlSyntaxError(err)
	}
	if m, ok := doc.(map[string]any); ok {
		if v, ok := m["version"].(int); ok && v > courseSchemaVersion {
			return nil, apperr.New(apperr.UnsupportedSchemaVersion, "version", strconv.Itoa(v))
		}
	}

	if problems, err := validateSchema(doc); err != nil {
		return nil, err
	} else if len(problems) > 0 {
		return nil, invalid(problems)
	}

	var c Course
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, yamlSyntaxError(err)
	}
	if problems := c.check(root); len(problems) > 0 {
		return nil, invalid(problems)
	}
	return &c, nil
}

func invalid(problems []apperr.Problem) *apperr.Error {
	e := apperr.New(apperr.ManifestInvalid)
	e.Problems = problems
	return e
}

var yamlLine = regexp.MustCompile(`line (\d+)`)

func yamlSyntaxError(err error) *apperr.Error {
	if m := yamlLine.FindStringSubmatch(err.Error()); m != nil {
		return apperr.New(apperr.YAMLSyntax, "line", m[1])
	}
	return apperr.New(apperr.YAMLSyntax)
}

// validateSchema validates the decoded YAML against the course schema and
// returns its problems sorted by path.
func validateSchema(doc any) ([]apperr.Problem, error) {
	schema, err := courseSchema()
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

// jsonCompatible converts maps with non-string keys, which YAML allows, to
// maps with string keys.
func jsonCompatible(v any) any {
	switch v := v.(type) {
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
