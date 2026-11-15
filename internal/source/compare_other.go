//go:build !windows

package source

// pathComparisonForm returns the form of a local-folder source under which
// equal sources are the same course. Paths are compared exactly.
func pathComparisonForm(source string) string {
	return source
}
