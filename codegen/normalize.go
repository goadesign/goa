// This file performs the one allowed post-evaluation design mutation. It gives
// raw method object shapes stable semantic user-type wrappers while leaving Go
// declaration naming to the generated service package catalog.
package codegen

import "goa.design/goa/v3/expr"

// NormalizeRoot wraps raw object payload, result, and streaming attributes in
// synthesized user types. The wrappers carry their natural preferred names and
// stable semantic identifiers; package planning later resolves Go collisions
// against declarations that are actually emitted in the service package.
//
// NormalizeRoot is idempotent and must run after prepare plugins and before
// generators read the design expression tree.
func NormalizeRoot(root *expr.RootExpr) {
	for _, service := range root.Services {
		normalizeService(service)
	}
}

// normalizeService creates semantic wrappers for the raw object attributes of
// one service without consulting or mutating any Go name scope.
func normalizeService(service *expr.ServiceExpr) {
	for _, method := range service.Methods {
		name := Goify(method.Name, true)
		normalizeMethodAttribute(method.Payload, name+"Payload", service.Name+"#"+name+"Payload")
		normalizeMethodAttribute(
			method.StreamingPayload,
			name+"StreamingPayload",
			service.Name+"#"+name+"StreamingPayload",
		)
		normalizeMethodAttribute(method.Result, name+"Result", service.Name+"#"+name+"Result")
		if method.HasMixedResults() {
			normalizeMethodAttribute(
				method.StreamingResult,
				name+"StreamingResult",
				service.Name+"#"+name+"StreamingResult",
			)
		}
	}
}

// normalizeMethodAttribute gives a raw method object its semantic identity.
// Existing named and non-object method types remain unchanged.
func normalizeMethodAttribute(attribute *expr.AttributeExpr, name, id string) {
	if attribute == nil {
		return
	}
	if _, ok := attribute.Type.(*expr.Object); !ok {
		return
	}
	attribute.Type = &expr.UserTypeExpr{
		AttributeExpr: expr.DupAtt(attribute),
		TypeName:      name,
		UID:           id,
	}
}
