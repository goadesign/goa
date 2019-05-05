package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
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
	if len(args) == 0 {
		args = []interface{}{expr.ErrorResult}
	}
	dt, desc, fn := parseAttributeArgs(nil, args...)
	att := &expr.AttributeExpr{
		Description: desc,
		Type:        dt,
	}
	if fn != nil {
		eval.Execute(fn, att)
	}
	if att.Type == nil {
		att.Type = expr.ErrorResult
	}
	erro := &expr.ErrorExpr{AttributeExpr: att, Name: name}
	switch actual := eval.Current().(type) {
	case *expr.ServiceExpr:
		actual.Errors = append(actual.Errors, erro)
	case *expr.MethodExpr:
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
//    var _ = Service("divider", func() {
//        Error("request_timeout", func() {
//            Temporary()
//        })
//    })
func Temporary() {
	attr, ok := eval.Current().(*expr.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if attr.Meta == nil {
		attr.Meta = make(expr.MetaExpr)
	}
	attr.Meta["goa:error:temporary"] = nil
}

// Timeout qualifies an error type as describing errors due to timeouts.
//
// Timeout must appear in a Error expression.
//
// Timeout takes no argument.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Error("request_timeout", func() {
//            Timeout()
//        })
//    })
func Timeout() {
	attr, ok := eval.Current().(*expr.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if attr.Meta == nil {
		attr.Meta = make(expr.MetaExpr)
	}
	attr.Meta["goa:error:timeout"] = nil
}

// Fault qualifies an error type as describing errors due to a server-side
// fault.
//
// Fault must appear in a Error expression.
//
// Fault takes no argument.
//
// Example:
//
//    var _ = Service("divider", func() {
//         Error("internal_error", func() {
//                 Fault()
//         })
//    })
func Fault() {
	attr, ok := eval.Current().(*expr.AttributeExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	if attr.Meta == nil {
		attr.Meta = make(expr.MetaExpr)
	}
	attr.Meta["goa:error:fault"] = nil
}
