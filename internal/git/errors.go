package git

import (
	"strings"

	"github.com/olymp-software/chiron/internal/apperr"
)

// outputLines is how many last lines of git's error output a GitFailed error
// carries.
const outputLines = 10

// patterns maps stable git and ssh messages, in English because git runs with
// LC_ALL=C, to error codes. The first match wins.
var patterns = []struct {
	code apperr.Code
	text []string
}{
	{apperr.RepoAccessDenied, []string{
		"Authentication failed",
		"could not read Username",
		"could not read Password",
		"terminal prompts disabled",
		"Permission denied (",
	}},
	{apperr.RepoNotFound, []string{
		"Repository not found",
		"does not appear to be a git repository",
		"The requested URL returned error: 404",
	}},
	{apperr.SshHostUnknown, []string{
		"Host key verification failed",
	}},
	{apperr.RepoUnreachable, []string{
		"Could not resolve host",
		"Failed to connect",
		"Connection timed out",
		"Network is unreachable",
		"Connection refused",
	}},
}

// classify maps git's error output to an error for the learner. Unknown
// output gives GitFailed with its last lines.
func classify(stderr string) *apperr.Error {
	for _, p := range patterns {
		for _, t := range p.text {
			if strings.Contains(stderr, t) {
				return apperr.New(p.code)
			}
		}
	}
	return apperr.New(apperr.GitFailed, "output", lastLines(stderr, outputLines))
}

// lastLines returns the last n non-empty lines of s.
func lastLines(s string, n int) string {
	var lines []string
	for l := range strings.Lines(strings.ReplaceAll(s, "\r\n", "\n")) {
		if l = strings.TrimRight(l, "\r\n"); strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
