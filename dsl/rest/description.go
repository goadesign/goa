package rest

import (
	"goa.design/goa.v2/design/rest"
	"goa.design/goa.v2/dsl"
	"goa.design/goa.v2/eval"
)

// Description sets the expression description.
//
// Description may appear in API, Resource, Action, Files, Response, Type, ResultType or Attribute.
//
// Example:
//
//    var _ = API("cellar", func() {
//        Description("The wine cellar API")
//    })
//
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *rest.HTTPResponseExpr:
		expr.Description = d
	case *rest.FileServerExpr:
		expr.Description = d
	default:
		dsl.Description(d)
	}
}
