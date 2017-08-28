package dsl

import (
	"goa.design/goa/dsl"
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

// Description sets the expression description.
//
// Description must appear in API, Service, Endpoint, Files, Response, Type,
// ResultType or Attribute.
//
// Description accepts a single argument which is the description value.
//
// Example:
//
//    var _ = API("cellar", func() {
//        Description("The wine cellar API")
//    })
//
func Description(d string) {
	switch expr := eval.Current().(type) {
	case *httpdesign.HTTPResponseExpr:
		expr.Description = d
	case *httpdesign.FileServerExpr:
		expr.Description = d
	default:
		dsl.Description(d)
	}
}
