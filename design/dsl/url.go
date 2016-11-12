package dsl

import "github.com/goadesign/goa/design"
import "github.com/goadesign/goa/eval"

// URL sets the expression url.
//
// URL may appear in Docs
//
// Example:
//
//    Docs( func() {
//        URL("http://example.com")
//    })
//
func URL(u string) {
	switch expr := eval.Current().(type) {
	case *design.DocsExpr:
		expr.URL = u
	default:
		eval.IncompatibleDSL()
	}
}
