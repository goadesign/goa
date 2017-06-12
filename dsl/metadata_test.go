package dsl_test

import (
	"testing"

	"goa.design/goa.v2/design"
	. "goa.design/goa.v2/dsl"
	"goa.design/goa.v2/eval"
)

func TestMetaData(t *testing.T) {
	cases := map[string]struct {
		Expr        eval.Expression
		Name        string
		Values      []string
		MetaFunc    func(e eval.Expression) design.MetadataExpr
		Invocations int
	}{
		"userType":  {&design.UserTypeExpr{AttributeExpr: &design.AttributeExpr{}}, "swagger:summary", []string{"Short summary of what action does"}, userTypeMeta, 1},
		"api":       {&design.APIExpr{}, "metadata", []string{"some metadata"}, apiExprMeta, 2},
		"attribute": {&design.AttributeExpr{}, "attribute_meta", []string{"attr meta", "more attr meta"}, attributeMeta, 2},
		"method":    {&design.MethodExpr{Name: "testmethod"}, "method", []string{"method meta"}, methodMeta, 2},
		"mediaType": {&design.MediaTypeExpr{UserTypeExpr: &design.UserTypeExpr{AttributeExpr: &design.AttributeExpr{}}}, "mediaTypeMeta", []string{"media type meta"}, mediaTypeMeta, 2},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			for i := tc.Invocations; i > 0; i-- {
				eval.Execute(func() {
					Metadata(tc.Name, tc.Values...)
				}, tc.Expr)
			}
			if eval.Context.Errors != nil {
				t.Errorf("%s: Metadata failed unexpectedly with %s", k, eval.Context.Errors)
			}
			meta := tc.MetaFunc(tc.Expr)
			if _, ok := meta[tc.Name]; !ok {
				t.Errorf("%s: expected %s to be present", k, tc.Name)
			}
			if len(meta[tc.Name]) != (len(tc.Values) * tc.Invocations) {
				t.Errorf("%s: expected the number of metadata values to match %d got %d ", k, len(tc.Values), len(meta[tc.Name]))
			}
			for _, caseVal := range tc.Values {
				if !hasValue(meta[tc.Name], caseVal) {
					t.Errorf("%s: meta data %v did not conatin expected value %v", k, meta[tc.Name], caseVal)
				}
			}
		})
	}
}

func hasValue(vals []string, val string) bool {
	for _, v := range vals {
		if v == val {
			return true
		}
	}
	return false
}
func apiExprMeta(e eval.Expression) design.MetadataExpr   { return e.(*design.APIExpr).Metadata }
func userTypeMeta(e eval.Expression) design.MetadataExpr  { return e.(*design.UserTypeExpr).Metadata }
func attributeMeta(e eval.Expression) design.MetadataExpr { return e.(*design.AttributeExpr).Metadata }
func methodMeta(e eval.Expression) design.MetadataExpr    { return e.(*design.MethodExpr).Metadata }
func mediaTypeMeta(e eval.Expression) design.MetadataExpr { return e.(*design.MediaTypeExpr).Metadata }
