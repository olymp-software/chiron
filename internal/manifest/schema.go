package manifest

import (
	"bytes"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/olymp-software/chiron/schemas"
)

const (
	courseSchemaURL           = "https://github.com/olymp-software/chiron/schemas/course/v1.json"
	courseActivitySchemaURL   = "https://github.com/olymp-software/chiron/schemas/course-activity/v1.json"
	workshopSchemaURL         = "https://github.com/olymp-software/chiron/schemas/workshop/v1.json"
	workshopActivitySchemaURL = "https://github.com/olymp-software/chiron/schemas/workshop-activity/v1.json"
)

var courseSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compileSchema(courseSchemaURL, schemas.CourseV1)
})

var courseActivitySchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compileSchema(courseActivitySchemaURL, schemas.CourseActivityV1)
})

var workshopSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compileSchema(workshopSchemaURL, schemas.WorkshopV1)
})

var workshopActivitySchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compileSchema(workshopActivitySchemaURL, schemas.WorkshopActivityV1)
})

// compileSchema compiles an embedded schema registered under url, so the
// compiler never fetches anything.
func compileSchema(url string, data []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(url, doc); err != nil {
		return nil, err
	}
	return c.Compile(url)
}
