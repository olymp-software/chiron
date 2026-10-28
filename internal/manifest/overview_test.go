package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
)

func writeRootFile(t *testing.T, root, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// caseSensitive reports whether root can hold files differing only in case.
func caseSensitive(t *testing.T, root string) bool {
	t.Helper()
	writeRootFile(t, root, "probe", "")
	_, err := os.Stat(filepath.Join(root, "PROBE"))
	if err := os.Remove(filepath.Join(root, "probe")); err != nil {
		t.Fatal(err)
	}
	return err != nil
}

func TestReadOverview(t *testing.T) {
	course := &Course{Description: "Go fundamentals"}
	tests := []struct {
		name   string
		files  map[string]string
		course *Course
		want   OverviewText
	}{
		{
			name:   "overview file wins over description",
			files:  map[string]string{"OVERVIEW.md": "# Go"},
			course: course,
			want:   OverviewText{Markdown: "# Go"},
		},
		{
			name:   "lowercase file name",
			files:  map[string]string{"overview.md": "# go"},
			course: course,
			want:   OverviewText{Markdown: "# go"},
		},
		{
			name:   "only description",
			files:  map[string]string{"README.md": "# readme"},
			course: course,
			want:   OverviewText{Plain: "Go fundamentals"},
		},
		{
			name:   "neither",
			course: &Course{},
			want:   OverviewText{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for name, content := range tt.files {
				writeRootFile(t, root, name, content)
			}

			got, err := ReadOverview(root, tt.course)
			if err != nil {
				t.Fatalf("ReadOverview: %v", err)
			}
			if got != tt.want {
				t.Errorf("overview = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestReadOverviewTwoFiles(t *testing.T) {
	root := t.TempDir()
	if !caseSensitive(t, root) {
		t.Skip("file system is case-insensitive")
	}
	writeRootFile(t, root, "OVERVIEW.md", "")
	writeRootFile(t, root, "overview.md", "")

	_, err := ReadOverview(root, &Course{})

	got := appErr(t, err)
	want := apperr.New(apperr.OverviewFileAmbiguous, "files", "OVERVIEW.md, overview.md")
	if got.Code != want.Code || got.Params["files"] != want.Params["files"] {
		t.Errorf("error = %+v, want %+v", got, want)
	}
}
