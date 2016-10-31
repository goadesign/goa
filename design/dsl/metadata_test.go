package dsl_test

import (
	"testing"

	"github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
)

func TestMetaData(t *testing.T) {
	cases := map[string]struct {
		Expr     eval.Expression
		Name     string
		Values   []string
		MetaFunc func(e eval.Expression) design.MetadataExpr
	}{
		"userType": {&design.UserTypeExpr{AttributeExpr: &design.AttributeExpr{}}, "swagger:summary", []string{"Short summary of what action does"}, userTypeMeta},
		"api":      {&design.APIExpr{}, "metadata", []string{"some metadata"}, apiExprMeta},
	}

	for k, tc := range cases {
		eval.Context = &eval.DSLContext{}
		eval.Execute(func() {
			Metadata(tc.Name, tc.Values...)
		}, tc.Expr)
		if eval.Context.Errors != nil {
			t.Errorf("%s: Description failed unexpectedly with %s", k, eval.Context.Errors)
		}
		meta := tc.MetaFunc(tc.Expr)
		if _, ok := meta[tc.Name]; !ok {
			t.Fatalf("expected %s to be present", tc.Name)
		}
		if len(meta[tc.Name]) != len(tc.Values) {
			t.Fatal("expected the number of metadata values to match ")
		}
	}
}
func apiExprMeta(e eval.Expression) design.MetadataExpr  { return e.(*design.APIExpr).Metadata }
func userTypeMeta(e eval.Expression) design.MetadataExpr { return e.(*design.UserTypeExpr).Metadata }
