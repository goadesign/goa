// This file supplies exact semantic owners to protobuf-shaping unit tests.
package codegen

import "goa.design/goa/v3/expr"

// testGRPCMessageExampleIdentity returns a distinct request-message owner for
// the named test case without introducing production fallback identity rules.
func testGRPCMessageExampleIdentity(name string) expr.ExampleIdentity {
	service := &expr.ServiceExpr{Name: "test"}
	method := &expr.MethodExpr{Name: name, Service: service}
	return expr.GRPCRequestMessageExampleIdentity(method)
}
