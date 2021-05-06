package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Redirect indicates that HTTP requests reply to the request with a redirect.
// The logic is the same as the standard http package Redirect function.
//
// Redirect must appear in a HTTP endpoint expression or a HTTP file server
// expression.
//
// Redirect accepts 2 arguments. The first argument is the URL that is being
// redirected to. The second argument is the HTTP status code.
//
// Example:
//
//    var _ = Service("service", func() {
//        Method("method", func() {
//            HTTP(func() {
//                GET("/resources")
//                Redirect("/redirect/dest", StatusMovedPermanently)
//            })
//        })
//    })
//
//    var _ = Service("service", func() {
//        Files("/file.json", "/path/to/file.json", func() {
//            Redirect("/redirect/dest", StatusMovedPermanently)
//        })
//    })
//
func Redirect(url string, code int) {
	redirect := &expr.HTTPRedirectExpr{
		URL:        url,
		StatusCode: code,
	}
	switch actual := eval.Current().(type) {
	case *expr.HTTPEndpointExpr:
		redirect.Parent = actual
		actual.Redirect = redirect
	case *expr.HTTPFileServerExpr:
		redirect.Parent = actual
		actual.Redirect = redirect
	default:
		eval.IncompatibleDSL()
	}
}
