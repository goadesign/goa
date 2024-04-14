package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// Method defines a single service method.
//
// Method must appear in a Service expression.
//
// Method takes two arguments: the name of the method and the defining DSL.
//
// Example:
//
//	Method("add", func() {
//	    Description("The add method returns the sum of A and B")
//	    Docs(func() {
//	        Description("Add docs")
//	        URL("http//adder.goa.design/docs/endpoints/add")
//	    })
//	    Payload(Operands)
//	    Result(Sum)
//	    Error(ErrInvalidOperands)
//	})
func Method(name string, fn func()) {
	if name == "" {
		eval.ReportError("method name cannot be empty")
	}
	s, ok := eval.Current().(*expr.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	ep := &expr.MethodExpr{Name: name, Service: s, DSLFunc: fn}
	s.Methods = append(s.Methods, ep)
}

// Deprecated marks HTTP routes as deprecated in the generated OpenAPI specifications.
//
// Deprecated must appear in a Method HTTP expression.
//
// Deprecated takes no argument.
// Example:
//
//	Method("add", func() {
//	    HTTP(func() {
//	        GET("/")
//	        Deprecated()
//	    })
//	})
func Deprecated() {
	_, ok := eval.Current().(*expr.HTTPEndpointExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	Meta("openapi:deprecated", "true")
}
