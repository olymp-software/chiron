//go:build !windows

package source

import "testing"

func TestSameSourceIsExact(t *testing.T) {
	if SameSource("/courses/go-basics", "/courses/GO-BASICS") {
		t.Error("sources differing in case are the same course, want different")
	}
}
