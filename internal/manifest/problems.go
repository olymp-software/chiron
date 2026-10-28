package manifest

import (
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"

	"github.com/olymp-software/chiron/internal/apperr"
)

// schemaKeywordCodes maps every validation keyword the course schemas use to
// the code its problems are reported with. A test fails if a schema uses a
// keyword missing here.
var schemaKeywordCodes = map[string]apperr.Code{
	"required":             apperr.MissingField,
	"additionalProperties": apperr.UnknownField,
	"type":                 apperr.WrongType,
	"pattern":              apperr.InvalidSlug, // the schemas use pattern only for slugs
	"minItems":             apperr.EmptyList,
	"minLength":            apperr.EmptyString,
	"const":                apperr.InvalidValue,
}

// collectProblems appends a problem for every leaf of a validation error tree.
func collectProblems(ve *jsonschema.ValidationError, out *[]apperr.Problem) {
	if len(ve.Causes) > 0 {
		for _, c := range ve.Causes {
			collectProblems(c, out)
		}
		return
	}
	path := yamlPath(ve.InstanceLocation)
	switch k := ve.ErrorKind.(type) {
	case *kind.Required:
		for _, f := range k.Missing {
			*out = append(*out, apperr.NewProblem(apperr.MissingField, path, "field", f))
		}
	case *kind.AdditionalProperties:
		for _, f := range k.Properties {
			*out = append(*out, apperr.NewProblem(apperr.UnknownField, path, "field", f))
		}
	case *kind.Type:
		*out = append(*out, apperr.NewProblem(apperr.WrongType, path,
			"want", strings.Join(k.Want, ", "), "got", k.Got))
	case *kind.Pattern:
		*out = append(*out, apperr.NewProblem(apperr.InvalidSlug, path, "value", k.Got))
	case *kind.MinItems:
		*out = append(*out, apperr.NewProblem(apperr.EmptyList, path))
	case *kind.MinLength:
		*out = append(*out, apperr.NewProblem(apperr.EmptyString, path))
	case *kind.Const:
		*out = append(*out, apperr.NewProblem(apperr.InvalidValue, path,
			"value", fmt.Sprint(k.Got), "want", fmt.Sprint(k.Want)))
	default:
		*out = append(*out, apperr.NewProblem(apperr.SchemaProblem, path,
			"keyword", strings.Join(ve.ErrorKind.KeywordPath(), "/")))
	}
}
