// This file wraps unnamed method payload and result objects in user types
// after evaluation. It records each wrapper so later code finds the same
// generated type without trusting a user-provided string.
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
	apiName := ""
	if root.API != nil {
		apiName = root.API.Name
	}
	for _, service := range root.Services {
		normalizeService(apiName, service, normalized)
	}
}

// normalizeService wraps each unnamed object payload and result in a generated
// user type without reading or changing generated Go names.
func normalizeService(apiName string, service *expr.ServiceExpr, normalized map[expr.UserType]MethodTypeIdentity) {
	for _, method := range service.Methods {
		normalizeMethodAttribute(method.Payload, newMethodTypeIdentity(
			apiName,
			method.Name,
			methodPayloadTypeKind,
			expr.MethodPayloadExampleIdentity(method),
		), normalized)
		normalizeMethodAttribute(method.StreamingPayload, newMethodTypeIdentity(
			apiName,
			method.Name,
			methodStreamingPayloadTypeKind,
			expr.MethodStreamingPayloadExampleIdentity(method),
		), normalized)
		normalizeMethodAttribute(method.Result, newMethodTypeIdentity(
			apiName,
			method.Name,
			methodResultTypeKind,
			expr.MethodResultExampleIdentity(method),
		), normalized)
		if method.HasMixedResults() {
			normalizeMethodAttribute(method.StreamingResult, newMethodTypeIdentity(
				apiName,
				method.Name,
				methodStreamingResultTypeKind,
				expr.MethodStreamingResultExampleIdentity(method),
			), normalized)
		}
	}
}

// normalizeMethodAttribute records which generated wrapper was created for an
// unnamed object. Existing named and non-object method types remain unchanged.
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
