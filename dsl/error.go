package dsl

import (
	"goa.design/goa.v2/design"
	"goa.design/goa.v2/eval"
)

// Error describes an endpoint error response. The description includes a unique
// name (in the scope of the endpoint), an optional type, description and DSL
// that further describes the type. If no type is specified then the goa
// ErrorMedia type is used. The DSL syntax is identical to the Attribute DSL.
// Transport specific DSL may further describe the mapping between the error
// type attributes and the serialized response.
//
// goa has a few predefined error names for the common cases, see ErrBadRequest
// for example.
//
// Error may appear in the Service (to define error responses that apply to all
// the service endpoints) or Endpoint expressions.
// See Attribute for details on the Error arguments.
//
// Example:
//
//    var _ = Service("divider", func() {
//        Error("invalid_arguments") // Uses type ErrorMedia
//
//        // Endpoint which uses the default type for its response.
//        Endpoint("divide", func() {
//            Payload(DivideRequest)
//            Error("div_by_zero", DivByZero, "Division by zero")
//        })
//    })
//
func Error(name string, args ...interface{}) {
	if len(args) == 0 {
		args = []interface{}{design.ErrorMedia}
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
	case *design.EndpointExpr:
		actual.Errors = append(actual.Errors, erro)
	default:
		eval.IncompatibleDSL()
	}
}
