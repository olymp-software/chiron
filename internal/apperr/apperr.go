// Package apperr defines the errors the app shows to the learner. Every such
// error has a code known in advance; the frontend translates it, so Go never
// returns user-facing text.
package apperr

import (
	"fmt"
	"strings"
)

// Code identifies one kind of error. The frontend message key is errors.<code>.
type Code string

// Codes of errors shown to the learner. Every constant must be listed in All.
const (
	// Course manifest as a whole.
	ManifestMissing          Code = "manifestMissing"
	ManifestInvalid          Code = "manifestInvalid"
	UnsupportedSchemaVersion Code = "unsupportedSchemaVersion"
	YAMLSyntax               Code = "yamlSyntax"

	// Problems inside a manifest reported by the JSON Schema.
	MissingField  Code = "missingField"
	UnknownField  Code = "unknownField"
	WrongType     Code = "wrongType"
	InvalidSlug   Code = "invalidSlug"
	EmptyList     Code = "emptyList"
	EmptyString   Code = "emptyString"
	InvalidValue  Code = "invalidValue"
	SchemaProblem Code = "schemaProblem"

	// Problems inside a manifest found by semantic checks.
	DuplicateSectionSlug  Code = "duplicateSectionSlug"
	DuplicateActivitySlug Code = "duplicateActivitySlug"
	PathInvalid           Code = "pathInvalid"
	PathNotFound          Code = "pathNotFound"

	// Course files.
	OverviewFileAmbiguous Code = "overviewFileAmbiguous"
	CourseReadFailed      Code = "courseReadFailed"

	// Library and course sources.
	FolderNotFound     Code = "folderNotFound"
	NotAFolder         Code = "notAFolder"
	StorageKeyTaken    Code = "storageKeyTaken"
	CourseNotFound     Code = "courseNotFound"
	CopyFailed         Code = "copyFailed"
	LibraryNewerFormat Code = "libraryNewerFormat"
	LibraryReadFailed  Code = "libraryReadFailed"
	LibraryWriteFailed Code = "libraryWriteFailed"

	// Internal is a failure that is a bug in the app; details are in the log.
	Internal Code = "internal"
)

// All lists every code, in declaration order.
var All = []Code{
	ManifestMissing,
	ManifestInvalid,
	UnsupportedSchemaVersion,
	YAMLSyntax,
	MissingField,
	UnknownField,
	WrongType,
	InvalidSlug,
	EmptyList,
	EmptyString,
	InvalidValue,
	SchemaProblem,
	DuplicateSectionSlug,
	DuplicateActivitySlug,
	PathInvalid,
	PathNotFound,
	OverviewFileAmbiguous,
	CourseReadFailed,
	FolderNotFound,
	NotAFolder,
	StorageKeyTaken,
	CourseNotFound,
	CopyFailed,
	LibraryNewerFormat,
	LibraryReadFailed,
	LibraryWriteFailed,
	Internal,
}

// Params are named values interpolated into the translated message.
type Params map[string]string

// Problem is one item of an error that has several, e.g. one invalid field of
// a manifest. Path is the location inside the file, e.g. sections[0].slug.
type Problem struct {
	Code   Code   `json:"code"`
	Path   string `json:"path,omitempty"`
	Params Params `json:"params,omitempty"`
}

// Error is an error shown to the learner.
type Error struct {
	Code     Code      `json:"code"`
	Params   Params    `json:"params,omitempty"`
	Problems []Problem `json:"problems,omitempty"`
}

// New returns an error with the given code and params given as key, value
// pairs.
func New(code Code, kv ...string) *Error {
	return &Error{Code: code, Params: pairs(kv)}
}

// NewProblem returns a problem at path with params given as key, value pairs.
func NewProblem(code Code, path string, kv ...string) Problem {
	return Problem{Code: code, Path: path, Params: pairs(kv)}
}

func pairs(kv []string) Params {
	if len(kv)%2 != 0 {
		panic("apperr: odd number of params")
	}
	if len(kv) == 0 {
		return nil
	}
	p := make(Params, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		p[kv[i]] = kv[i+1]
	}
	return p
}

// Error describes the error for logs; it is not shown to the learner.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(string(e.Code))
	if len(e.Params) > 0 {
		fmt.Fprintf(&b, " %v", map[string]string(e.Params))
	}
	for _, p := range e.Problems {
		fmt.Fprintf(&b, "; %s %s", p.Path, p.Code)
		if len(p.Params) > 0 {
			fmt.Fprintf(&b, " %v", map[string]string(p.Params))
		}
	}
	return b.String()
}
