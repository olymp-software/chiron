// Package schemas embeds the JSON Schemas of course files. The same files are
// published for authors' editors.
package schemas

import _ "embed"

// CourseV1 is the JSON Schema of course.yaml, schema version 1.
//
//go:embed course/v1.json
var CourseV1 []byte

// ActivityV1 is the JSON Schema of activity.yaml, schema version 1.
//
//go:embed activity/v1.json
var ActivityV1 []byte
