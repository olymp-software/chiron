package source

import "strings"

// pathComparisonForm returns the form of a local-folder source under which
// equal sources are the same course. Windows paths are case-insensitive.
func pathComparisonForm(source string) string {
	return strings.ToLower(source)
}
