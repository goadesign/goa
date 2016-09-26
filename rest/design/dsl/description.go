package dsl

import (
	"github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/eval"
	"github.com/goadesign/goa/rest/design"
)

// Description sets the expression description.
//
// Description may appear in API, Resource, Action, Files, Response, Type, MediaType or Attribute.
//
// Example:
//
//    var _ = API("cellar", func() {
//        Description("The wine cellar API")
//    })
//
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *design.ResponseExpr:
		expr.Description = d
	case *design.MediaTypeExpr:
		expr.Description = d
	case *design.FileServerExpr:
		expr.Description = d
	default:
		dsl.Description(d)
	}
}
