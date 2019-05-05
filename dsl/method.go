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
//    Method("add", func() {
//        Description("The add method returns the sum of A and B")
//        Docs(func() {
//            Description("Add docs")
//            URL("http//adder.goa.design/docs/endpoints/add")
//        })
//        Payload(Operands)
//        Result(Sum)
//        Error(ErrInvalidOperands)
//    })
//
func Method(name string, fn func()) {
	s, ok := eval.Current().(*expr.ServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	ep := &expr.MethodExpr{Name: name, Service: s, DSLFunc: fn}
	s.Methods = append(s.Methods, ep)
}
