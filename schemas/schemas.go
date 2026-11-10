// Package schemas embeds the JSON Schemas of course files. The same files are
// published for authors' editors.
package schemas

import _ "embed"

// CourseV1 is the JSON Schema of course.yaml, schema version 1.
//
//go:embed course/v1.json
var CourseV1 []byte

// CourseActivityV1 is the JSON Schema of activity.yaml of a course, schema
// version 1.
//
//go:embed course-activity/v1.json
var CourseActivityV1 []byte

// WorkshopV1 is the JSON Schema of workshop.yaml, schema version 1.
//
//go:embed workshop/v1.json
var WorkshopV1 []byte

// WorkshopActivityV1 is the JSON Schema of activity.yaml of a workshop, schema
// version 1.
//
//go:embed workshop-activity/v1.json
var WorkshopActivityV1 []byte
