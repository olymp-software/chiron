package git

import (
	"slices"
	"testing"
)

func TestIsVersion(t *testing.T) {
	for tag, want := range map[string]bool{
		"v0.0.0":        true,
		"v1.2.3":        true,
		"v10.20.30":     true,
		"v1.10.0":       true,
		"1.2.3":         false,
		"v01.2.3":       false,
		"v1.02.3":       false,
		"v1.2.03":       false,
		"v1.2":          false,
		"v1.2.3.4":      false,
		"v2.0.0-beta.1": false,
		"v1.2.3+build":  false,
		"release-3":     false,
		"V1.2.3":        false,
		"":              false,
	} {
		if got := IsVersion(tag); got != want {
			t.Errorf("IsVersion(%q) = %v, want %v", tag, got, want)
		}
	}
}

func TestVersions(t *testing.T) {
	got := Versions([]string{"v1.0.0", "v1.2.0", "release-3", "v1.10.0", "v2.0.0-beta.1", "v1.2.0", "v0.9.9"})
	want := []string{"v1.10.0", "v1.2.0", "v1.0.0", "v0.9.9"}
	if !slices.Equal(got, want) {
		t.Errorf("Versions = %q, want %q", got, want)
	}
	if got := Versions(nil); got == nil || len(got) != 0 {
		t.Errorf("Versions(nil) = %#v, want empty", got)
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v1.10.0", "v1.2.0", 1},
		{"v1.2.0", "v1.10.0", -1},
		{"v2.0.0", "v1.99.99", 1},
		{"v1.0.1", "v1.0.1", 0},
		{"", "v0.0.0", -1},
		{"v0.0.0", "", 1},
	}
	for _, tt := range tests {
		if got := CompareVersions(tt.a, tt.b); got != tt.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestHighest(t *testing.T) {
	if got := Highest([]string{"v1.0.0", "v1.0.1", "latest"}); got != "v1.0.1" {
		t.Errorf("Highest = %q, want v1.0.1", got)
	}
	if got := Highest([]string{"release-3"}); got != "" {
		t.Errorf("Highest = %q, want empty", got)
	}
}
