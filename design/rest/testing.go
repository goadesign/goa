package rest

import (
	"testing"

	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// RunRestDSL returns the rest DSL root resulting from running the given DSL.
func RunRestDSL(t *testing.T, dsl func()) *RootExpr {
	// reset all roots and codegen data structures
	eval.Reset()
	design.Root = new(design.RootExpr)
	Root = &RootExpr{Design: design.Root}
	eval.Register(design.Root)
	eval.Register(Root)
	design.Root.API = &design.APIExpr{
		Name:    "test api",
		Servers: []*design.ServerExpr{{URL: "http://localhost"}},
	}

	// run DSL (first pass)
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}

	// run DSL (second pass)
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}

	// return generated root
	return Root
}
