package dsl

import (
	"goa.design/goa/design"
	"goa.design/goa/eval"
)

// Error describes a method error return value. The description includes a
// unique name (in the scope of the method), an optional type, description and
// DSL that further describes the type. If no type is specified then the
// built-in ErrorResult type is used. The DSL syntax is identical to the
// Attribute DSL.
//
// Error must appear in the Service (to define error responses that apply to all
// the service methods) or Method expressions.
//
// See Attribute for details on the Error arguments.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Error("invalid_arguments") // Uses type ErrorResult
//
//        // Method which uses the default type for its response.
//        Method("divide", func() {
//            Payload(DivideRequest)
//            Error("div_by_zero", DivByZero, "Division by zero")
//        })
//    })
//
func Error(name string, args ...interface{}) {
	er := len(args) == 0
	if !er && len(args) == 1 {
		_, er = args[0].(func())
	}
	if er {
		args = []interface{}{design.ErrorResult}
	}
	dt, desc, fn := parseAttributeArgs(nil, args...)
	att := &design.AttributeExpr{
		Description: desc,
		Type:        dt,
	}
	if fn != nil {
		eval.Execute(fn, att)
	}
	erro := &design.ErrorExpr{AttributeExpr: att, Name: name}
	switch actual := eval.Current().(type) {
	case *design.ServiceExpr:
		actual.Errors = append(actual.Errors, erro)
	case *design.MethodExpr:
		actual.Errors = append(actual.Errors, erro)
	default:
		eval.IncompatibleDSL()
	}
}

// Temporary qualifies an error type as describing temporary (i.e. retryable)
// errors.
//
// Temporary must appear in a Error expression.
//
// Temporary takes no argument.
//
// Example:
//
// var _ = Service("divider", func() {
//      Error("request_timeout", func() {
//              Temporary()
//      })
// })
func Temporary() {
	attr, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if attr.Metadata == nil {
		attr.Metadata = make(design.MetadataExpr)
	}
	attr.Metadata["goa:error:temporary"] = nil
}

// Timeout qualifies an error type as describing errors due to timeouts.
//
// Timeout must appear in a Error expression.
//
// Timeout takes no argument.
//
// Example:
//
// var _ = Service("divider", func() {
//	Error("request_timeout", func() {
//		Timeout()
//	})
// })
func Timeout() {
	attr, ok := eval.Current().(*design.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if attr.Metadata == nil {
		attr.Metadata = make(design.MetadataExpr)
	}
	attr.Metadata["goa:error:timeout"] = nil
}
