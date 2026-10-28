package manifest

import (
	"encoding/json"
	"testing"

	"github.com/olymp-software/chiron/schemas"
)

func TestCourseSchemaCompiles(t *testing.T) {
	if _, err := courseSchema(); err != nil {
		t.Fatalf("compile course schema: %v", err)
	}
}

func TestCourseSchemaID(t *testing.T) {
	var doc struct {
		ID string `json:"$id"`
	}
	if err := json.Unmarshal(schemas.CourseV1, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ID != courseSchemaURL {
		t.Errorf("$id = %q, want %q", doc.ID, courseSchemaURL)
	}
}
