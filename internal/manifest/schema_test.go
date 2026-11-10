package manifest

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
	"github.com/olymp-software/chiron/schemas"
)

func TestCourseSchemaCompiles(t *testing.T) {
	if _, err := courseSchema(); err != nil {
		t.Fatalf("compile course schema: %v", err)
	}
}

func TestCourseSchemaID(t *testing.T) {
	assertSchemaID(t, schemas.CourseV1, courseSchemaURL)
}

func TestCourseActivitySchemaCompiles(t *testing.T) {
	if _, err := courseActivitySchema(); err != nil {
		t.Fatalf("compile course activity schema: %v", err)
	}
}

func TestCourseActivitySchemaID(t *testing.T) {
	assertSchemaID(t, schemas.CourseActivityV1, courseActivitySchemaURL)
}

func TestWorkshopSchemaCompiles(t *testing.T) {
	if _, err := workshopSchema(); err != nil {
		t.Fatalf("compile workshop schema: %v", err)
	}
}

func TestWorkshopSchemaID(t *testing.T) {
	assertSchemaID(t, schemas.WorkshopV1, workshopSchemaURL)
}

func TestWorkshopActivitySchemaCompiles(t *testing.T) {
	if _, err := workshopActivitySchema(); err != nil {
		t.Fatalf("compile workshop activity schema: %v", err)
	}
}

func TestWorkshopActivitySchemaID(t *testing.T) {
	assertSchemaID(t, schemas.WorkshopActivityV1, workshopActivitySchemaURL)
}

func TestWorkshopActivitySchema(t *testing.T) {
	const blocks = `"theory_blocks": [{"kind": "markdown", "data": {"content": "# Hi"}}]`
	tests := []struct {
		name string
		doc  string
		want []apperr.Problem
	}{
		{
			name: "only theory blocks",
			doc:  `{"version": 1, ` + blocks + `}`,
		},
		{
			name: "every field",
			doc: `{"version": 1, ` + blocks + `, "solution": "text", "check": "check.sh",
				"controls": [{"label": "Up", "script": "up.sh", "concurrency": "server"}],
				"files": [{"path": "starter"}]}`,
		},
		{
			name: "type",
			doc:  `{"version": 1, "type": "theory", ` + blocks + `}`,
			want: []apperr.Problem{apperr.NewProblem(apperr.UnknownField, "", "field", "type")},
		},
		{
			name: "data",
			doc:  `{"version": 1, "data": {}, ` + blocks + `}`,
			want: []apperr.Problem{apperr.NewProblem(apperr.UnknownField, "", "field", "data")},
		},
		{
			name: "empty solution",
			doc:  `{"version": 1, "solution": "", ` + blocks + `}`,
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyString, "solution")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc any
			if err := json.Unmarshal([]byte(tt.doc), &doc); err != nil {
				t.Fatal(err)
			}

			got, err := validateSchema(workshopActivitySchema, doc)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("problems = %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func assertSchemaID(t *testing.T, data []byte, want string) {
	t.Helper()
	var doc struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ID != want {
		t.Errorf("$id = %q, want %q", doc.ID, want)
	}
}
