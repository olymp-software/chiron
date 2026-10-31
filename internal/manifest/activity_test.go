package manifest

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
)

const activityFile = "basics/hello/activity.yaml"

// writeActivity writes basics/hello/activity.yaml in a new course root and
// creates files relative to the activity folder; a path ending in / is a
// folder.
func writeActivity(t *testing.T, yaml string, files ...string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "basics", "hello")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ActivityFile), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		path := filepath.Join(dir, filepath.FromSlash(f))
		if strings.HasSuffix(f, "/") {
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func mustFailActivity(t *testing.T, root string) *apperr.Error {
	t.Helper()
	a, err := ParseActivity(root, "basics", "hello")
	if err == nil {
		t.Fatalf("ParseActivity = %+v, want an error", a)
	}
	return appErr(t, err)
}

const theoryBlocks = "theory_blocks:\n  - kind: markdown\n    data: {content: '# Hi'}\n"

func TestParseActivityValid(t *testing.T) {
	tests := []struct {
		name  string
		yaml  string
		files []string
		want  *Activity
	}{
		{
			name: "theory",
			yaml: "version: 1\ntype: theory\ntheory_blocks:\n" +
				"  - kind: markdown\n    data: {content: '# One'}\n" +
				"  - kind: markdown\n    data: {content: '# Two'}\n" +
				"controls:\n  - {label: Check, script: check.sh, concurrency: docker}\n",
			files: []string{"check.sh"},
			want: &Activity{
				Version: 1,
				Type:    TypeTheory,
				TheoryBlocks: []TheoryBlock{
					{Kind: KindMarkdown, Data: MarkdownData{Content: "# One"}},
					{Kind: KindMarkdown, Data: MarkdownData{Content: "# Two"}},
				},
				Controls: []Control{{Label: "Check", Script: "check.sh", Concurrency: "docker"}},
			},
		},
		{
			name: "quiz",
			yaml: "version: 1\ntype: quiz\n" + theoryBlocks +
				"data:\n  questions:\n" +
				"    - kind: choice\n      prompt: Pick\n      explanation: Because\n" +
				"      data:\n        multiple: true\n        options:\n" +
				"          - {text: A, correct: true}\n          - {text: B, feedback: No}\n" +
				"    - kind: choice\n      data:\n        options:\n" +
				"          - {text: C}\n          - {text: D, correct: true}\n",
			want: &Activity{
				Version:      1,
				Type:         TypeQuiz,
				TheoryBlocks: []TheoryBlock{{Kind: KindMarkdown, Data: MarkdownData{Content: "# Hi"}}},
				Quiz: &Quiz{Questions: []Question{
					{Kind: "choice", Prompt: "Pick", Explanation: "Because", Data: Choice{
						Multiple: true,
						Options:  []Option{{Text: "A", Correct: true}, {Text: "B", Feedback: "No"}},
					}},
					{Kind: "choice", Data: Choice{Options: []Option{{Text: "C"}, {Text: "D", Correct: true}}}},
				}},
			},
		},
		{
			name: "coding",
			yaml: "version: 1\ntype: coding\n" + theoryBlocks +
				"controls:\n  - {label: Run, script: scripts/run.sh}\n" +
				"data:\n  workspace:\n    - path: main.go\n    - path: src\n",
			files: []string{"main.go", "src/", "scripts/run.sh"},
			want: &Activity{
				Version:      1,
				Type:         TypeCoding,
				TheoryBlocks: []TheoryBlock{{Kind: KindMarkdown, Data: MarkdownData{Content: "# Hi"}}},
				Controls:     []Control{{Label: "Run", Script: "scripts/run.sh"}},
				Coding:       &Coding{Workspace: []WorkspaceItem{{Path: "main.go"}, {Path: "src"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := writeActivity(t, tt.yaml, tt.files...)

			a, err := ParseActivity(root, "basics", "hello")
			if err != nil {
				t.Fatalf("ParseActivity: %v", err)
			}
			if !reflect.DeepEqual(a, tt.want) {
				t.Errorf("activity = %+v\nwant %+v", a, tt.want)
			}
		})
	}
}

func TestParseActivityFileMissing(t *testing.T) {
	root := writeActivity(t, "")
	if err := os.Remove(filepath.Join(root, "basics", "hello", ActivityFile)); err != nil {
		t.Fatal(err)
	}
	want := apperr.New(apperr.ActivityFileMissing, "file", activityFile)

	if got := mustFailActivity(t, root); !reflect.DeepEqual(got, want) {
		t.Errorf("missing file: error = %+v, want %+v", got, want)
	}
	if err := os.RemoveAll(filepath.Join(root, "basics")); err != nil {
		t.Fatal(err)
	}
	if got := mustFailActivity(t, root); !reflect.DeepEqual(got, want) {
		t.Errorf("missing folder: error = %+v, want %+v", got, want)
	}
}

func TestParseActivityYAMLSyntax(t *testing.T) {
	root := writeActivity(t, "version: 1\ntype: [unclosed\n")

	got := mustFailActivity(t, root)
	if got.Code != apperr.YAMLSyntax || got.Params["line"] == "" || got.Params["file"] != activityFile {
		t.Errorf("error = %+v, want %s with a line in %s", got, apperr.YAMLSyntax, activityFile)
	}
}

func TestParseActivityNewerApp(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want *apperr.Error
	}{
		{
			// Fields unknown to schema 1 must not be reported for a newer schema.
			name: "version 2",
			yaml: "version: 2\ntype: theory\nhints: []\n",
			want: apperr.New(apperr.UnsupportedSchemaVersion, "file", activityFile, "version", "2"),
		},
		{
			name: "unknown type",
			yaml: "version: 1\ntype: terminal\n" + theoryBlocks,
			want: apperr.New(apperr.UnsupportedActivityType, "file", activityFile, "type", "terminal"),
		},
		{
			name: "unknown block kind",
			yaml: "version: 1\ntype: theory\ntheory_blocks:\n  - kind: diagram\n    data: {}\n",
			want: apperr.New(apperr.UnsupportedBlockKind, "file", activityFile, "kind", "diagram"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := writeActivity(t, tt.yaml)

			if got := mustFailActivity(t, root); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("error = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseActivityProblems(t *testing.T) {
	const quiz = "version: 1\ntype: quiz\n" + theoryBlocks
	const coding = "version: 1\ntype: coding\n" + theoryBlocks
	tests := []struct {
		name  string
		yaml  string
		files []string
		want  []apperr.Problem
	}{
		{
			name: "content instead of theory_blocks",
			yaml: "version: 1\ntype: theory\ncontent: '# Hi'\n",
			want: []apperr.Problem{
				apperr.NewProblem(apperr.MissingField, "", "field", "theory_blocks"),
				apperr.NewProblem(apperr.UnknownField, "", "field", "content"),
			},
		},
		{
			name: "empty theory_blocks",
			yaml: "version: 1\ntype: theory\ntheory_blocks: []\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyList, "theory_blocks")},
		},
		{
			name: "missing theory_blocks",
			yaml: "version: 1\ntype: quiz\ndata:\n  questions:\n    - kind: choice\n      data: {options: [{text: A, correct: true}, {text: B}]}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.MissingField, "", "field", "theory_blocks")},
		},
		{
			name: "block without data",
			yaml: "version: 1\ntype: theory\ntheory_blocks:\n  - kind: markdown\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.MissingField, "theory_blocks[0]", "field", "data")},
		},
		{
			name: "several problems",
			yaml: "version: 1\ntype: theory\ntitel: Hello\ntheory_blocks:\n  - kind: markdown\n",
			want: []apperr.Problem{
				apperr.NewProblem(apperr.UnknownField, "", "field", "titel"),
				apperr.NewProblem(apperr.MissingField, "theory_blocks[0]", "field", "data"),
			},
		},
		{
			name: "empty markdown content",
			yaml: "version: 1\ntype: theory\ntheory_blocks:\n  - kind: markdown\n    data: {content: ''}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyString, "theory_blocks[0].data.content")},
		},
		{
			name: "theory with data",
			yaml: "version: 1\ntype: theory\n" + theoryBlocks + "data: {}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.UnknownField, "", "field", "data")},
		},
		{
			name: "type of wrong type",
			yaml: "version: 1\ntype: 5\n" + theoryBlocks,
			want: []apperr.Problem{
				apperr.NewProblem(apperr.InvalidValue, "type", "value", "5", "want", "theory, quiz, coding"),
			},
		},
		{
			name: "quiz without data",
			yaml: quiz,
			want: []apperr.Problem{apperr.NewProblem(apperr.MissingField, "", "field", "data")},
		},
		{
			name: "no questions",
			yaml: quiz + "data: {questions: []}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyList, "data.questions")},
		},
		{
			name: "single option",
			yaml: quiz + "data:\n  questions:\n    - kind: choice\n      data: {options: [{text: A, correct: true}]}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyList, "data.questions[0].data.options")},
		},
		{
			name: "two correct radio options",
			yaml: quiz + "data:\n  questions:\n    - kind: choice\n" +
				"      data: {options: [{text: A, correct: true}, {text: B, correct: true}]}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.OneCorrectOptionRequired, "data.questions[0].data.options")},
		},
		{
			name: "no correct checkbox option",
			yaml: quiz + "data:\n  questions:\n    - kind: choice\n" +
				"      data: {multiple: true, options: [{text: A}, {text: B}]}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.CorrectOptionRequired, "data.questions[0].data.options")},
		},
		{
			name:  "workspace item as string",
			yaml:  coding + "data:\n  workspace:\n    - main.go\n",
			files: []string{"main.go"},
			want: []apperr.Problem{
				apperr.NewProblem(apperr.WrongType, "data.workspace[0]", "want", "object", "got", "string"),
			},
		},
		{
			name: "empty workspace",
			yaml: coding + "data: {workspace: []}\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.EmptyList, "data.workspace")},
		},
		{
			name: "missing workspace folder",
			yaml: coding + "data:\n  workspace:\n    - path: src\n",
			want: []apperr.Problem{apperr.NewProblem(apperr.PathNotFound, "data.workspace[0].path", "path", "src")},
		},
		{
			name: "workspace is the activity folder",
			yaml: coding + "data:\n  workspace:\n    - path: .\n    - path: ./\n",
			want: []apperr.Problem{
				apperr.NewProblem(apperr.PathInvalid, "data.workspace[0].path", "path", "."),
				apperr.NewProblem(apperr.PathInvalid, "data.workspace[1].path", "path", "./"),
			},
		},
		{
			name: "infinite version",
			yaml: "version: .inf\ntype: theory\n" + theoryBlocks,
			want: []apperr.Problem{
				apperr.NewProblem(apperr.InvalidValue, "version", "value", "+Inf", "want", "1"),
			},
		},
		{
			name:  "script outside the activity folder",
			yaml:  "version: 1\ntype: theory\n" + theoryBlocks + "controls:\n  - {label: Test, script: ../shared/test.sh}\n",
			files: []string{"../shared/test.sh"},
			want: []apperr.Problem{
				apperr.NewProblem(apperr.PathInvalid, "controls[0].script", "path", "../shared/test.sh"),
			},
		},
		{
			name:  "script is a folder",
			yaml:  "version: 1\ntype: theory\n" + theoryBlocks + "controls:\n  - {label: Test, script: scripts}\n",
			files: []string{"scripts/"},
			want: []apperr.Problem{
				apperr.NewProblem(apperr.PathNotFound, "controls[0].script", "path", "scripts"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := writeActivity(t, tt.yaml, tt.files...)

			got := mustFailActivity(t, root)
			if got.Code != apperr.ManifestInvalid || got.Params["file"] != activityFile {
				t.Fatalf("error = %v, want %s in %s", got, apperr.ManifestInvalid, activityFile)
			}
			if !reflect.DeepEqual(got.Problems, tt.want) {
				t.Errorf("problems = %+v\nwant %+v", got.Problems, tt.want)
			}
		})
	}
}
