package dsl

import (
	"testing"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestView(t *testing.T) {
	viewName, view2Name := "test", "test2"
	baseRT := &expr.ResultTypeExpr{
		UserTypeExpr: &expr.UserTypeExpr{
			AttributeExpr: &expr.AttributeExpr{
				Type: &expr.Object{
					{Name: "att", Attribute: &expr.AttributeExpr{Type: expr.String}},
					{Name: "att2", Attribute: &expr.AttributeExpr{Type: expr.String}},
				},
			},
			TypeName: "test",
		},
	}
	viewedRT := expr.Dup(baseRT).(*expr.ResultTypeExpr)
	viewedRT.Views = []*expr.ViewExpr{{
		Name:          viewName,
		AttributeExpr: viewedRT.AttributeExpr,
	}}
	viewDSL := func() {
		View(viewName, func() {
			Attribute("att")
		})
	}
	view2DSL := func() {
		View(view2Name, func() {
			Attribute("att2")
		})
	}
	allViewsDSL := func() {
		View(viewName, func() {
			Attribute("att")
		})
		View(view2Name, func() {
			Attribute("att2")
		})
	}
	cases := []struct {
		name              string
		rt                *expr.ResultTypeExpr
		dsl               func()
		expectedViews     []string
		expectedViewAttrs map[string][]string
		expectedErr       string
	}{
		{"noop", baseRT, func() {}, nil, nil, ""},
		{"view", baseRT, viewDSL, []string{viewName}, map[string][]string{viewName: {"att"}}, ""},
		{"view2", baseRT, view2DSL, []string{view2Name}, map[string][]string{view2Name: {"att2"}}, ""},
		{"all views", baseRT, allViewsDSL, []string{viewName, view2Name}, map[string][]string{viewName: {"att"}, view2Name: {"att2"}}, ""},
		{"duplicate view", baseRT, func() { viewDSL(); viewDSL() }, nil, nil, `[result_type_test.go:29] multiple expressions for view "test" in result type "test" in attribute`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			rt := expr.Dup(c.rt).(*expr.ResultTypeExpr)
			eval.Execute(c.dsl, rt)
			err := eval.Context.Errors
			if len(c.expectedErr) > 0 {
				if err == nil {
					t.Errorf("got no error, expected %s", c.expectedErr)
				} else if got, want := err.Error(), c.expectedErr; got != want {
					t.Errorf("got error %s, expected %s", got, want)
				}
				return
			}
			if err != nil {
				t.Errorf("got error %s, expected none", err)
			}
			if got, want := len(rt.Views), len(c.expectedViews); got != want {
				t.Errorf("got %d views, expected %d", got, want)
			}
			for _, view := range c.expectedViews {
				found := false
				for _, v := range rt.Views {
					if v.Name == view {
						found = true
						for _, attr := range c.expectedViewAttrs[view] {
							found2 := false
							for _, attr2 := range *v.AttributeExpr.Type.(*expr.Object) {
								if attr2.Name == attr {
									found2 = true
									break
								}
							}
							if !found2 {
								t.Errorf("attribute %s not found in view %s", attr, view)
							}
						}
					}
				}
				if !found {
					t.Errorf("view %s not found", view)
				}
			}
		})
	}
}
