package source

import "testing"

func TestSameSourceIgnoresCase(t *testing.T) {
	if !SameSource(`C:\Courses\go-basics`, `c:\courses\GO-BASICS`) {
		t.Error("sources differing only in case are different courses, want the same")
	}
	if SameSource(`C:\Courses\go-basics`, `C:\Courses\go-advanced`) {
		t.Error("different folders are the same course")
	}
}

func TestKeyIgnoresCase(t *testing.T) {
	if got, want := Key(`C:\Courses\Go Basics`), "go-basics-4fb70bd4"; got != want {
		t.Errorf("Key = %q, want %q (same as the lowercase spelling)", got, want)
	}
}
