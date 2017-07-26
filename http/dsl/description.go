package dsl

import (
	"goa.design/goa.v2/dsl"
	"goa.design/goa.v2/eval"
	"goa.design/goa.v2/http/design"
)

// Description sets the expression description.
//
// Description may appear in API, Service, Endpoint, Files, Response, Type, ResultType or Attribute.
//
// Example:
//
//    var _ = API("cellar", func() {
//        Description("The wine cellar API")
//    })
//
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *design.HTTPResponseExpr:
		expr.Description = d
	case *design.FileServerExpr:
		expr.Description = d
	default:
		dsl.Description(d)
	}
}
