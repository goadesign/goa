package dsl

import (
	goa "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/rest/design"
)

// Description sets the expression description.
//
// Description may appear in Service, EndpointGroup, Endpoint, Type or Field.
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *design.ResponseExpr:
		expr.Description = d
	case *design.MediaTypeExpr:
		expr.Description = d
	case *design.FileServerExpr:
		expr.Description = d
	default:
		goa.Description(d)
	}
}
