package rest

import (
	"bytes"
	"testing"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/files"
	"goa.design/goa.v2/design"
	rest "goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/eval"
)

// RunRestDSL returns the rest DSL root resulting from running the given DSL.
func RunRestDSL(t *testing.T, dsl func()) *rest.RootExpr {
	// reset all roots and codegen data structures
	eval.Reset()
	design.Root = new(design.RootExpr)
	rest.Root = &rest.RootExpr{Design: design.Root}
	eval.Register(design.Root)
	eval.Register(rest.Root)
	design.Root.API = &design.APIExpr{
		Name:    "test api",
		Servers: []*design.ServerExpr{{URL: "http://localhost"}},
	}
	files.Services = make(files.ServicesData)
	Resources = make(ResourcesData)

	// run DSL (first pass)
	if !eval.Execute(dsl, nil) {
		t.Fatal(eval.Context.Error())
	}

	// run DSL (second pass)
	if err := eval.RunDSL(); err != nil {
		t.Fatal(err)
	}

	// return generated root
	return rest.Root
}

// SectionCode generates and formats the code for the given section.
func SectionCode(t *testing.T, section *codegen.Section) string {
	var code bytes.Buffer
	if err := section.Write(&code); err != nil {
		t.Fatal(err)
	}
	return codegen.FormatTestCode(t, "package foo\n"+code.String())
}
