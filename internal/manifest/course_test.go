package manifest

import (
	"encoding/json"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/schemas"
)

const validCourse = `# yaml-language-server: $schema=https://example.com/course/v1.json
version: 1
title: Go Basics
description: Go fundamentals
check: scripts/check.sh
sections:
  - slug: basics
    title: First steps
    activities:
      - slug: hello
        title: Hello, Go
  - slug: advanced
    title: Advanced
    activities:
      - slug: loops
        title: Loops
`

func writeCourse(t *testing.T, yaml string, files ...string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, CourseFile), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		path := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func appErr(t *testing.T, err error) *apperr.Error {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("err = %v, want *apperr.Error", err)
	}
	return ae
}

func TestParseCourseValid(t *testing.T) {
	root := writeCourse(t, validCourse, "scripts/check.sh")

	c, err := ParseCourse(root)
	if err != nil {
		t.Fatalf("ParseCourse: %v", err)
	}

	want := &Course{
		Version:     1,
		Title:       "Go Basics",
		Description: "Go fundamentals",
		Check:       "scripts/check.sh",
		Sections: []Section{
			{Slug: "basics", Title: "First steps", Activities: []ActivityRef{{Slug: "hello", Title: "Hello, Go"}}},
			{Slug: "advanced", Title: "Advanced", Activities: []ActivityRef{{Slug: "loops", Title: "Loops"}}},
		},
	}
	if !reflect.DeepEqual(c, want) {
		t.Errorf("course = %+v\nwant %+v", c, want)
	}
}

func TestParseCourseMissing(t *testing.T) {
	_, err := ParseCourse(t.TempDir())

	if got := appErr(t, err); got.Code != apperr.ManifestMissing {
		t.Errorf("code = %s, want %s", got.Code, apperr.ManifestMissing)
	}
}

func TestParseCourseYAMLSyntax(t *testing.T) {
	root := writeCourse(t, "version: 1\ntitle: [unclosed\n")

	got := appErr(t, mustFail(t, root))
	if got.Code != apperr.YAMLSyntax || got.Params["line"] == "" {
		t.Errorf("error = %+v, want %s with a line", got, apperr.YAMLSyntax)
	}
}

func TestParseCourseNewerSchema(t *testing.T) {
	// Fields unknown to schema 1 must not be reported for a newer schema.
	root := writeCourse(t, "version: 2\ntitle: X\nauthor: Someone\n")

	got := appErr(t, mustFail(t, root))
	want := apperr.New(apperr.UnsupportedSchemaVersion, "version", "2")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("error = %+v, want %+v", got, want)
	}
}

func mustFail(t *testing.T, root string) error {
	t.Helper()
	c, err := ParseCourse(root)
	if err == nil {
		t.Fatalf("ParseCourse = %+v, want an error", c)
	}
	return err
}

func TestParseCourseProblems(t *testing.T) {
	tests := []struct {
		name  string
		yaml  string
		files []string
		want  []apperr.Problem
	}{
		{
			name: "unknown field",
			yaml: "version: 1\ntitle: X\nautor: Someone\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.UnknownField, "", "field", "autor")},
		},
		{
			name: "missing fields",
			yaml: "version: 1\nsections: [{slug: a, title: A, activities: [{slug: b}]}]\n",
			want: []apperr.Problem{
				apperr.NewProblem(apperr.MissingField, "", "field", "title"),
				apperr.NewProblem(apperr.MissingField, "sections[0].activities[0]", "field", "title"),
			},
		},
		{
			name: "empty section",
			yaml: "version: 1\ntitle: X\nsections: [{slug: a, title: A, activities: []}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyList, "sections[0].activities")},
		},
		{
			name: "no sections",
			yaml: "version: 1\ntitle: X\nsections: []\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyList, "sections")},
		},
		{
			name: "invalid slug",
			yaml: "version: 1\ntitle: X\nsections: [{slug: a, title: A, activities: [{slug: Hello_World, title: B}]}]\n",
			want: []apperr.Problem{
				apperr.NewProblem(apperr.InvalidSlug, "sections[0].activities[0].slug", "value", "Hello_World"),
			},
		},
		{
			name: "wrong type",
			yaml: "version: 1\ntitle: [X]\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.WrongType, "title", "want", "string", "got", "array")},
		},
		{
			name: "empty title",
			yaml: "version: 1\ntitle: ''\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyString, "title")},
		},
		{
			name: "version zero",
			yaml: "version: 0\ntitle: X\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.InvalidValue, "version", "value", "0", "want", "1")},
		},
		{
			name: "duplicate activity slug in different sections",
			yaml: "version: 1\ntitle: X\nsections:\n" +
				"  - {slug: basics, title: A, activities: [{slug: intro, title: I}]}\n" +
				"  - {slug: advanced, title: B, activities: [{slug: intro, title: I}]}\n",
			want: []apperr.Problem{
				apperr.NewProblem(apperr.DuplicateActivitySlug, "sections[1].activities[0].slug", "slug", "intro"),
			},
		},
		{
			name: "duplicate section slug",
			yaml: "version: 1\ntitle: X\nsections:\n" +
				"  - {slug: basics, title: A, activities: [{slug: a, title: I}]}\n" +
				"  - {slug: basics, title: B, activities: [{slug: b, title: I}]}\n",
			want: []apperr.Problem{
				apperr.NewProblem(apperr.DuplicateSectionSlug, "sections[1].slug", "slug", "basics"),
			},
		},
		{
			name: "check escapes the course",
			yaml: "version: 1\ntitle: X\ncheck: ../outside.sh\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.PathInvalid, "check", "path", "../outside.sh")},
		},
		{
			name:  "check with backslashes",
			yaml:  "version: 1\ntitle: X\ncheck: 'scripts\\check.sh'\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			files: []string{"scripts/check.sh"},
			want:  []apperr.Problem{apperr.NewProblem(apperr.PathInvalid, "check", "path", `scripts\check.sh`)},
		},
		{
			name: "check is absolute",
			yaml: "version: 1\ntitle: X\ncheck: 'C:/check.sh'\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.PathInvalid, "check", "path", "C:/check.sh")},
		},
		{
			name: "check missing",
			yaml: "version: 1\ntitle: X\ncheck: scripts/check.sh\nsections: [{slug: a, title: A, activities: [{slug: b, title: B}]}]\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.PathNotFound, "check", "path", "scripts/check.sh")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := writeCourse(t, tt.yaml, tt.files...)

			got := appErr(t, mustFail(t, root))
			if got.Code != apperr.ManifestInvalid {
				t.Fatalf("code = %s, want %s (%v)", got.Code, apperr.ManifestInvalid, got)
			}
			if !reflect.DeepEqual(got.Problems, tt.want) {
				t.Errorf("problems = %+v\nwant %+v", got.Problems, tt.want)
			}
		})
	}
}

func TestYAMLPath(t *testing.T) {
	if got := yamlPath([]string{"sections", "0", "activities", "1", "slug"}); got != "sections[0].activities[1].slug" {
		t.Errorf("yamlPath = %q", got)
	}
	if got := yamlPath(nil); got != "" {
		t.Errorf("yamlPath(nil) = %q, want empty", got)
	}
}

// structuralKeywords are schema keywords that never produce a problem of
// their own.
var structuralKeywords = map[string]bool{
	"$schema": true, "$id": true, "$defs": true, "$ref": true,
	"title": true, "description": true, "properties": true, "items": true,
}

// schemaKeywords collects the keywords used in a schema object and its
// subschemas.
func schemaKeywords(schema map[string]any, out map[string]bool) {
	for k, v := range schema {
		out[k] = true
		switch k {
		case "properties", "$defs":
			for _, sub := range v.(map[string]any) {
				schemaKeywords(sub.(map[string]any), out)
			}
		case "items", "additionalProperties":
			if sub, ok := v.(map[string]any); ok {
				schemaKeywords(sub, out)
			}
		}
	}
}

func TestEverySchemaKeywordHasCode(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(schemas.CourseV1, &schema); err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	schemaKeywords(schema, used)

	for _, k := range slices.Sorted(maps.Keys(used)) {
		if structuralKeywords[k] {
			continue
		}
		if _, ok := schemaKeywordCodes[k]; !ok {
			t.Errorf("course schema uses keyword %q without a code in schemaKeywordCodes", k)
		}
	}
}
