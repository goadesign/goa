package testing

import (
	"testing"

	"goa.design/goa.v2/design"
	rest "goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/eval"
)

// RunRestDSL returns the rest DSL root resulting from running the given DSL.
func RunRestDSL(t *testing.T, dsl func()) *rest.RootExpr {
	eval.Reset()
	design.Root = new(design.RootExpr)
	rest.Root = &rest.RootExpr{Design: design.Root}
	eval.Register(design.Root)
	eval.Register(rest.Root)
	design.Root.API = &design.APIExpr{
		Name:    "test api",
		Servers: []*design.ServerExpr{{URL: "http://localhost"}},
	}
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}
	return rest.Root
}
