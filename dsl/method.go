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
func Method(name string, fn func()) *expr.MethodExpr {
	if name == "" {
		eval.ReportError("method name cannot be empty")
	}
	s, ok := eval.Current().(*expr.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return nil
	}
	me := &expr.MethodExpr{
		Name:    name,
		Service: s,
		Stream:  expr.NoStreamKind,
		DSLFunc: fn,
	}
	s.Methods = append(s.Methods, me)
	return me
}

// Idempotent marks a method as safe to retry with the exact same input.
//
// Idempotent must appear in a Method expression.
//
// Marking a method idempotent is a service contract: replaying the same
// invocation must have the same externally visible effect as invoking the
// method once. Transport generators use this contract to advertise or
// configure retry behavior where the transport supports it.
//
// Idempotent takes no argument.
//
// Example:
//
//	Method("show", func() {
//	    Idempotent()
//	    Payload(func() {
//	        Attribute("id", String)
//	        Required("id")
//	    })
//	    Result(Book)
//	})
func Idempotent() {
	method, ok := eval.Current().(*expr.MethodExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	method.Idempotent = true
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
