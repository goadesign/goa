package dsl

import (
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestDescription(t *testing.T) {
	const (
		description = "test description"
	)

	cases := map[string]struct {
		Expr     eval.Expression
		Desc     string
		DescFunc func(e eval.Expression) string
	}{
		"api":  {&expr.APIExpr{}, description, apiDesc},
		"attr": {&expr.AttributeExpr{}, description, attrDesc},
		"docs": {&expr.DocsExpr{}, description, docsDesc},
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

func apiDesc(e eval.Expression) string  { return e.(*expr.APIExpr).Description }
func attrDesc(e eval.Expression) string { return e.(*expr.AttributeExpr).Description }
func docsDesc(e eval.Expression) string { return e.(*expr.DocsExpr).Description }
