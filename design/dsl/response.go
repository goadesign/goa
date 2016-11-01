package dsl

import (
	"fmt"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/eval"
)

// Response describes and endpoint response type. Transport specific DSL may
// provide a mapping between the type attributes and the response state when the
// type is an object (e.g. which object attributes are written to the HTTP
// response headers and which ones to the body).
//
// Response may appear in a Endpoint expression.
//
// Response takes one or two arguments. The first argument is either a reference
// to a type, the name of a type or a DSL function.
// If the first argument is a type or the name of a type then an optional DSL
// may be passed as second argument that further specializes the type by
// providing additional validations (e.g. list of required attributes)
//
// Examples:
//
// Endpoint("add", func() {
//     // Define response using primitive type
//     Response(Int32)
// })
//
// Endpoint("add", func() {
//     // Define response using object defined inline
//     Response(func() {
//         Attribute("value", Int32, "Resulting sum")
//         Required("value")
//     })
// })
//
// Endpoint("add", func() {
//     // Define response type using user type
//     Response(Sum) // this works too: Response("Sum")
// })
//
// Endpoint("add", func() {
//     // Specify required attributes on user type
//     Response(Sum, func() {
//         Required("value")
//     })
// })
//
func Response(val interface{}, dsls ...func()) {
	att := endpointTypeDSL(val, dsls...)
	if att == nil {
		return
	}
	e := eval.Current().(*design.EndpointExpr)
	sn := camelize(e.Service.Name)
	en := camelize(e.Name)
	e.Response = &design.UserTypeExpr{
		AttributeExpr: att,
		TypeName:      fmt.Sprintf("%s%sResponse", en, sn),
	}
}
