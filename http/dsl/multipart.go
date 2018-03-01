package dsl

import (
	"goa.design/goa/eval"
	httpdesign "goa.design/goa/http/design"
)

// MultipartRequest defines the HTTP request for the endpoint to be a
// multipart content type.
//
// MultipartRequest must appear in a HTTP endpoint expression.
//
func MultipartRequest() {
	e, ok := eval.Current().(*httpdesign.EndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	e.MultipartRequest = true
}
