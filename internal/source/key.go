package source

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	maxLabel     = 32
	hashLen      = 8
	defaultLabel = "course"
)

// Key returns the storage key of a course source: a readable label from the
// source's last segment, "-", and a short hash of the source's comparison
// form. It contains only [a-z0-9-]. The algorithm must never change: stored
// data is found by it.
func Key(source string) string {
	sum := sha256.Sum256([]byte(ComparisonForm(source)))
	return label(source) + "-" + hex.EncodeToString(sum[:])[:hashLen]
}

// label returns the last segment of source, lowercased, with every run of
// characters outside [a-z0-9] replaced by "-", trimmed of "-" and cut to 32
// characters. Both / and \ separate segments, so Windows paths give the same
// label on every OS.
func label(source string) string {
	segment := strings.TrimRight(source, `/\`)
	if i := strings.LastIndexAny(segment, `/\`); i >= 0 {
		segment = segment[i+1:]
	}
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(segment) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > maxLabel {
		s = strings.TrimRight(s[:maxLabel], "-")
	}
	if s == "" {
		return defaultLabel
	}
	return s
}
