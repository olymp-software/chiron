package apperr

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"testing"
)

// declaredCodes returns the names of the Code constants in apperr.go.
func declaredCodes(t *testing.T) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "apperr.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			if id, ok := vs.Type.(*ast.Ident); ok && id.Name == "Code" {
				for _, n := range vs.Values {
					lit := n.(*ast.BasicLit)
					names = append(names, lit.Value[1:len(lit.Value)-1])
				}
			}
		}
	}
	return names
}

func TestAllListsEveryCode(t *testing.T) {
	var all []string
	for _, c := range All {
		all = append(all, string(c))
	}
	declared := declaredCodes(t)
	if len(declared) == 0 {
		t.Fatal("no Code constants found")
	}
	if !slices.Equal(all, declared) {
		t.Errorf("All = %v\nwant the declared constants in order %v", all, declared)
	}
}

func TestTypeScriptUpToDate(t *testing.T) {
	got, err := os.ReadFile("../../frontend/src/errors/codes.ts")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(TypeScript()) {
		t.Error("frontend/src/errors/codes.ts is stale; run go generate ./internal/apperr")
	}
}

func TestErrorDescribesForLog(t *testing.T) {
	e := New(ManifestInvalid)
	e.Problems = []Problem{NewProblem(UnknownField, "sections[0]", "field", "autor")}

	want := "manifestInvalid; sections[0] unknownField map[field:autor]"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
