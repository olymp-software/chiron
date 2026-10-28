//go:build !windows

package source

// ComparisonForm returns the form of a course source under which equal
// sources are the same course. Paths are compared exactly.
func ComparisonForm(source string) string {
	return source
}
