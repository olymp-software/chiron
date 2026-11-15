package source

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/olymp-software/chiron/internal/apperr"
)

const (
	httpsPrefix = "https://"
	sshPrefix   = "ssh://"
	gitSuffix   = ".git"
	// githubShorthand starts the github.com/owner/repo shorthand.
	githubShorthand = "github.com/"
)

// scpLike matches an scp-like ssh address user@host:path. Unlike git, the
// user is required, so a Windows path such as C:\x never matches.
var scpLike = regexp.MustCompile(`^([^@/\\:]+)@([^:/\\]+):/?(.+)$`)

// NormalizeRepo returns the course source of a repository address: an
// https:// or ssh:// address with the scheme and host lowercased and trailing
// slashes removed; the rest stays as typed. It accepts https://…, ssh://…,
// scp-like user@host:path and the github.com/owner/repo shorthand; anything
// else, including http://, git:// and file://, gives RepoAddressInvalid.
func NormalizeRepo(addr string) (string, error) {
	invalid := apperr.New(apperr.RepoAddressInvalid, "address", addr)
	a := strings.TrimSpace(addr)
	if strings.HasPrefix(strings.ToLower(a), githubShorthand) {
		a = httpsPrefix + a
	}
	if m := scpLike.FindStringSubmatch(a); m != nil && !strings.Contains(a, "://") {
		a = sshPrefix + m[1] + "@" + m[2] + "/" + m[3]
	}
	scheme, rest, ok := strings.Cut(a, "://")
	if !ok {
		return "", invalid
	}
	scheme = strings.ToLower(scheme)
	if scheme != "https" && scheme != "ssh" {
		return "", invalid
	}
	authority, path, _ := strings.Cut(rest, "/")
	path = strings.TrimRight(path, "/")
	user, host, hasUser := strings.Cut(authority, "@")
	if !hasUser {
		user, host = "", authority
	}
	// A password or token in the address would be stored and logged with the
	// course source; credentials belong to git's credential helpers. The error
	// is logged too, so it carries the address without the password.
	if name, _, hasPassword := strings.Cut(user, ":"); hasUser && hasPassword {
		masked := scheme + "://" + name + ":***@" + strings.ToLower(host) + "/" + path
		return "", apperr.New(apperr.RepoAddressInvalid, "address", masked)
	}
	if host == "" || path == "" || hasUser && user == "" || strings.ContainsAny(a, " \t\\?#") {
		return "", invalid
	}
	if hasUser {
		user += "@"
	}
	source := scheme + "://" + user + strings.ToLower(host) + "/" + path
	if u, err := url.Parse(source); err != nil || u.Hostname() == "" {
		return "", invalid
	}
	return source, nil
}

// isRepoSource reports whether source is a repository course source, as
// NormalizeRepo returns it.
func isRepoSource(source string) bool {
	return strings.HasPrefix(source, httpsPrefix) || strings.HasPrefix(source, sshPrefix)
}

// ComparisonForm returns the form of a course source under which equal
// sources are the same course. Repository sources are compared the same way
// on every OS: lowercased, without a trailing .git. Local-folder sources
// follow the OS's path rules.
func ComparisonForm(source string) string {
	if isRepoSource(source) {
		return strings.TrimSuffix(strings.ToLower(source), gitSuffix)
	}
	return pathComparisonForm(source)
}
