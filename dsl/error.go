package dsl

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	pkg "goa.design/goa/v3/pkg"
)

const (
	// The constants below make it possible for the service specific code to
	// return error names that are consistent with the names used by the
	// generated request and response payload validation code.
	//
	// Usage:
	//
	// var _ = Service("divider", func() {
	//     Error(MissingField)
	//     Error(InvalidRange)
	//
	//     Payload(func() {
	//        Attribute("numerator", Int)
	//        Attribute("denominator", Int, func() {
	//            Minimum(1)
	//        })
	//        Required("numerator", "denominator")
	//     })
	//
	//     HTTP(func() {
	//        Response(MissingField, StatusBadRequest)
	//        Response(InvalidRange, StatusBadRequest)
	//     })
	//
	//     GRPC(func() {
	//         Response(MissingField, CodeInvalidArgument)
	//         Response(InvalidRange, CodeInvalidArgument)
	//     })
	// })

	// InvalidFieldType is the error name for invalid field type errors.
	InvalidFieldType = pkg.InvalidFieldType
	// MissingField is the error name for missing field errors.
	MissingField = pkg.MissingField
	// InvalidEnumValue is the error name for invalid enum value errors.
	InvalidEnumValue = pkg.InvalidEnumValue
	// InvalidFormat is the error name for invalid format errors.
	InvalidFormat = pkg.InvalidFormat
	// InvalidPattern is the error name for invalid pattern errors.
	InvalidPattern = pkg.InvalidPattern
	// InvalidRange is the error name for invalid range errors.
	InvalidRange = pkg.InvalidRange
	// InvalidLength is the error name for invalid length errors.
	InvalidLength = pkg.InvalidLength
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
	case *expr.APIExpr:
		expr.Root.Errors = append(expr.Root.Errors, erro)
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
	attr.AddMeta("goa:error:temporary")
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
	attr.AddMeta("goa:error:timeout")
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
	attr.AddMeta("goa:error:fault")
}
