package source

import "strings"

// ComparisonForm returns the form of a course source under which equal
// sources are the same course. Windows paths are case-insensitive.
func ComparisonForm(source string) string {
	return strings.ToLower(source)
}
