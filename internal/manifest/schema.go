package manifest

import (
	"bytes"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/olymp-software/chiron/schemas"
)

const courseSchemaURL = "https://github.com/olymp-software/chiron/schemas/course/v1.json"

var courseSchema = sync.OnceValues(func() (*jsonschema.Schema, error) {
	return compileSchema(courseSchemaURL, schemas.CourseV1)
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
