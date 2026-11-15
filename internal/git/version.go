package git

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
)

// versionRe matches a course version tag: vX.Y.Z without leading zeros or a
// pre-release part.
var versionRe = regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

// version is a parsed course version.
type version [3]uint64

// parseVersion parses a vX.Y.Z tag name.
func parseVersion(tag string) (version, bool) {
	m := versionRe.FindStringSubmatch(tag)
	if m == nil {
		return version{}, false
	}
	var v version
	for i := range v {
		n, err := strconv.ParseUint(m[i+1], 10, 64)
		if err != nil {
			return version{}, false
		}
		v[i] = n
	}
	return v, true
}

// IsVersion reports whether tag is a course version.
func IsVersion(tag string) bool {
	_, ok := parseVersion(tag)
	return ok
}

// CompareVersions compares two course versions like cmp.Compare. A tag that
// is not a version is lower than every version.
func CompareVersions(a, b string) int {
	va, okA := parseVersion(a)
	vb, okB := parseVersion(b)
	if !okA || !okB {
		return cmp.Compare(boolInt(okA), boolInt(okB))
	}
	for i := range va {
		if c := cmp.Compare(va[i], vb[i]); c != 0 {
			return c
		}
	}
	return 0
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Versions returns the course versions among tags, highest first, without
// duplicates.
func Versions(tags []string) []string {
	vs := []string{}
	for _, t := range tags {
		if IsVersion(t) && !slices.Contains(vs, t) {
			vs = append(vs, t)
		}
	}
	slices.SortFunc(vs, func(a, b string) int { return CompareVersions(b, a) })
	return vs
}

// Highest returns the highest course version among tags, "" if there is none.
func Highest(tags []string) string {
	if vs := Versions(tags); len(vs) > 0 {
		return vs[0]
	}
	return ""
}
