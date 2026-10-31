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
	assertSchemaID(t, schemas.CourseV1, courseSchemaURL)
}

func TestActivitySchemaCompiles(t *testing.T) {
	if _, err := activitySchema(); err != nil {
		t.Fatalf("compile activity schema: %v", err)
	}
}

func TestActivitySchemaID(t *testing.T) {
	assertSchemaID(t, schemas.ActivityV1, activitySchemaURL)
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
