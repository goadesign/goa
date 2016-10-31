package dsl

import (
	"testing"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

const (
	description = "test description"
)

func TestDescription(t *testing.T) {
	cases := map[string]struct {
		Expr     eval.Expression
		Desc     string
		DescFunc func(e eval.Expression) string
	}{
		"api":  {&design.APIExpr{}, description, apiDesc},
		"attr": {&design.AttributeExpr{}, description, attrDesc},
		"docs": {&design.DocsExpr{}, description, docsDesc},
	}

	for k, tc := range cases {
		eval.Context = &eval.DSLContext{}

		eval.Execute(func() { Description(tc.Desc) }, tc.Expr)

		if eval.Context.Errors != nil {
			t.Errorf("%s: Description failed unexpectedly with %s", k, eval.Context.Errors)
		}
		if tc.DescFunc(tc.Expr) != tc.Desc {
			t.Errorf("%s: Description not set on %+v, expected %s, got %+v", k, tc.Expr, tc.Desc, tc.DescFunc(tc.Expr))
		}
	}
}

func apiDesc(e eval.Expression) string  { return e.(*design.APIExpr).Description }
func attrDesc(e eval.Expression) string { return e.(*design.AttributeExpr).Description }
func docsDesc(e eval.Expression) string { return e.(*design.DocsExpr).Description }
