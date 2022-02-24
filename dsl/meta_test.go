package dsl_test

import (
	"testing"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestMetaData(t *testing.T) {
	cases := map[string]struct {
		Expr        eval.Expression
		Name        string
		Values      []string
		MetaFunc    func(e eval.Expression) expr.MetaExpr
		Invocations int
	}{
		"userType":   {&expr.UserTypeExpr{AttributeExpr: &expr.AttributeExpr{}}, "openapi:summary", []string{"Short summary of what endpoint does"}, userTypeMeta, 1},
		"api":        {&expr.APIExpr{}, "metadata", []string{"some metadata"}, apiExprMeta, 2},
		"attribute":  {&expr.AttributeExpr{}, "attribute_meta", []string{"attr meta", "more attr meta"}, attributeMeta, 2},
		"method":     {&expr.MethodExpr{Name: "testmethod"}, "method", []string{"method meta"}, methodMeta, 2},
		"resultType": {&expr.ResultTypeExpr{UserTypeExpr: &expr.UserTypeExpr{AttributeExpr: &expr.AttributeExpr{}}}, "resultTypeMeta", []string{"result type meta"}, resultTypeMeta, 2},
	}

	for k, tc := range cases {
		t.Run(k, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			for i := tc.Invocations; i > 0; i-- {
				eval.Execute(func() {
					Meta(tc.Name, tc.Values...)
				}, tc.Expr)
			}
			if eval.Context.Errors != nil {
				t.Errorf("%s: Meta failed unexpectedly with %s", k, eval.Context.Errors)
			}
			meta := tc.MetaFunc(tc.Expr)
			if _, ok := meta[tc.Name]; !ok {
				t.Errorf("%s: expected %s to be present", k, tc.Name)
			}
			if len(meta[tc.Name]) != (len(tc.Values) * tc.Invocations) {
				t.Errorf("%s: expected the number of meta values to match %d got %d ", k, len(tc.Values), len(meta[tc.Name]))
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
func apiExprMeta(e eval.Expression) expr.MetaExpr    { return e.(*expr.APIExpr).Meta }
func userTypeMeta(e eval.Expression) expr.MetaExpr   { return e.(*expr.UserTypeExpr).Meta }
func attributeMeta(e eval.Expression) expr.MetaExpr  { return e.(*expr.AttributeExpr).Meta }
func methodMeta(e eval.Expression) expr.MetaExpr     { return e.(*expr.MethodExpr).Meta }
func resultTypeMeta(e eval.Expression) expr.MetaExpr { return e.(*expr.ResultTypeExpr).Meta }
