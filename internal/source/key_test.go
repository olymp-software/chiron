package source

import (
	"regexp"
	"strings"
	"testing"
)

// The expected keys are literal on purpose: stored data is found by the key,
// so the algorithm must never change. Sources are lowercase so the keys are
// the same on case-insensitive Windows.
func TestKeyLiteral(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{source: `c:\courses\go basics`, want: "go-basics-4fb70bd4"},
		{source: "github.com/owner/repo", want: "repo-620c3a93"},
		{source: "/home/learner/курс", want: "course-c9841a64"},
	}
	for _, tt := range tests {
		if got := Key(tt.source); got != tt.want {
			t.Errorf("Key(%q) = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestKeyCharacters(t *testing.T) {
	valid := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	for _, source := range []string{
		`C:\Courses\Go Basics`,
		`C:\Courses\Go Basics\`,
		"github.com/owner/repo",
		"/home/learner/My_Course (v2)!",
		"/",
	} {
		if got := Key(source); !valid.MatchString(got) {
			t.Errorf("Key(%q) = %q, want only [a-z0-9-]", source, got)
		}
	}
}

func TestKeyLabel(t *testing.T) {
	tests := []struct {
		source string
		want   string
	}{
		{source: `C:\Courses\Go Basics`, want: "go-basics"},
		{source: `C:\Courses\Go Basics\`, want: "go-basics"},
		{source: "/home/learner/My_Course (v2)!", want: "my-course-v2"},
		{source: "/home/learner/---", want: "course"},
		{source: "/", want: "course"},
		{source: "/home/" + strings.Repeat("a", 31) + "-bcd", want: strings.Repeat("a", 31)},
		{source: "/home/" + strings.Repeat("x", 40), want: strings.Repeat("x", 32)},
	}
	for _, tt := range tests {
		if got := label(tt.source); got != tt.want {
			t.Errorf("label(%q) = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestKeySameLastSegment(t *testing.T) {
	a, b := Key(`c:\a\course`), Key(`c:\b\course`)
	if a == b {
		t.Errorf("Key = %q for both folders, want different keys", a)
	}
	if !strings.HasPrefix(a, "course-") || !strings.HasPrefix(b, "course-") {
		t.Errorf("keys = %q, %q, want the label course", a, b)
	}
}
