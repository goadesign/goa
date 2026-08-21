// This file performs the one allowed post-evaluation design mutation. It gives
// raw method object shapes stable semantic user-type wrappers and records the
// exact wrapper objects so later planning never infers compiler provenance from
// a user-controlled string.
package codegen

import (
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// normalizeRoots wraps raw method objects in every Goa design root and returns
// the exact compiler-created declarations with their closed method roles.
func normalizeRoots(roots []eval.Root) map[expr.UserType]MethodTypeIdentity {
	normalized := make(map[expr.UserType]MethodTypeIdentity)
	for _, root := range roots {
		if design, ok := root.(*expr.RootExpr); ok {
			normalizeRoot(design, normalized)
		}
	}
	return normalized
}

// normalizeRoot records every wrapper created for one design root.
func normalizeRoot(root *expr.RootExpr, normalized map[expr.UserType]MethodTypeIdentity) {
	for _, service := range root.Services {
		normalizeService(service, normalized)
	}
}

// normalizeService creates semantic wrappers for one service without
// consulting or mutating any Go name scope.
func normalizeService(service *expr.ServiceExpr, normalized map[expr.UserType]MethodTypeIdentity) {
	for _, method := range service.Methods {
		normalizeMethodAttribute(method.Payload, newMethodTypeIdentity(
			method.Name,
			methodPayloadTypeKind,
			expr.MethodPayloadExampleIdentity(method),
		), normalized)
		normalizeMethodAttribute(method.StreamingPayload, newMethodTypeIdentity(
			method.Name,
			methodStreamingPayloadTypeKind,
			expr.MethodStreamingPayloadExampleIdentity(method),
		), normalized)
		normalizeMethodAttribute(method.Result, newMethodTypeIdentity(
			method.Name,
			methodResultTypeKind,
			expr.MethodResultExampleIdentity(method),
		), normalized)
		if method.HasMixedResults() {
			normalizeMethodAttribute(method.StreamingResult, newMethodTypeIdentity(
				method.Name,
				methodStreamingResultTypeKind,
				expr.MethodStreamingResultExampleIdentity(method),
			), normalized)
		}
	}
}

// normalizeMethodAttribute records typed provenance only for the wrapper it
// creates. Existing named and non-object method types remain authored values.
func normalizeMethodAttribute(attribute *expr.AttributeExpr, identity MethodTypeIdentity, normalized map[expr.UserType]MethodTypeIdentity) {
	if attribute == nil {
		return
	}
	if userType, ok := attribute.Type.(expr.UserType); ok {
		exampleIdentity, generated := expr.GeneratedUserTypeExampleIdentity(userType)
		if generated && exampleIdentity == identity.exampleIdentity {
			normalized[userType.Origin()] = identity.bind(userType)
		}
		return
	}
	if _, ok := attribute.Type.(*expr.Object); !ok {
		return
	}
	wrapper := expr.NewGeneratedUserType(identity.Name(), expr.DupAtt(attribute), identity.exampleIdentity)
	attribute.Type = wrapper
	normalized[wrapper.Origin()] = identity.bind(wrapper)
}
