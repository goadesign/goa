package testing

import (
	"testing"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// RunDSL returns the DSL root resulting from running the given DSL.
func RunDSL(t *testing.T, dsl func()) *design.RootExpr {
	eval.Reset()
	design.Root = new(design.RootExpr)
	eval.Register(design.Root)
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
	return design.Root
}
