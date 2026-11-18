// This file validates prepared URL bindings against a service method's payload.
// HTTP, JSON-RPC, and transport plugins use the same path and query rules before
// generating decoders; unsupported types and missing fields become design errors.
package expr

import (
	"fmt"
	"reflect"

	"goa.design/goa/v3/eval"
)

// ValidateParams checks prepared path and query bindings against the method's
// payload. It resolves the payload types without changing the bindings and
// returns errors for unsupported parameter types or missing payload fields.
// Plugins that supply their own request envelope can use this stage without
// validating HTTP bodies, responses, streaming, or security mappings.
func (e *HTTPEndpointExpr) ValidateParams() *eval.ValidationErrors {
	if e.Params.IsEmpty() {
		return nil
	}

	var (
		pparams = DupMappedAtt(e.PathParams())
		qparams = DupMappedAtt(e.QueryParams())
	)
	// We have to figure out the actual type for the params because the actual
	// type is initialized only during the finalize phase. In the validation
	// phase, all param types are string type by default unless specified
	// explicitly.
	initAttr(pparams, e.MethodExpr.Payload)
	initAttr(qparams, e.MethodExpr.Payload)

	invalidTypeErr := func(verr *eval.ValidationErrors, e *HTTPEndpointExpr, name string) {
		verr.Add(e, "path parameter %s cannot be an object, path parameter types must be primitive, array or map (query string only)", name)
	}
	verr := new(eval.ValidationErrors)
	WalkMappedAttr(pparams, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
		switch {
		case IsObject(a.Type), IsMap(a.Type), IsUnion(a.Type):
			invalidTypeErr(verr, e, name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array path parameter %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("path parameter %s", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	WalkMappedAttr(qparams, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
		switch {
		case IsObject(a.Type), IsUnion(a.Type):
			invalidTypeErr(verr, e, name)
		case IsArray(a.Type):
			arr := AsArray(a.Type)
			if !IsPrimitive(arr.ElemType.Type) {
				verr.Add(e, "elements of array query parameter %q must be primitive", name)
			}
		default:
			ctx := fmt.Sprintf("query parameter %s", name)
			verr.Merge(a.Validate(ctx, e))
		}
		return nil
	})
	if e.MethodExpr.Payload != nil {
		switch e.MethodExpr.Payload.Type.(type) {
		case *Object, UserType:
			WalkMappedAttr(pparams, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Path parameter %q not found in payload.", name)
					return nil
				}
				e.validateMappedDefault(verr, "path parameter", name, a)
				return nil
			})
			WalkMappedAttr(qparams, func(name, _ string, a *AttributeExpr) error { // nolint: errcheck
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Query string parameter %q not found in payload.", name)
					return nil
				}
				e.validateMappedDefault(verr, "query parameter", name, a)
				return nil
			})
		case *Array:
			if len(*AsObject(pparams.Type))+len(*AsObject(qparams.Type)) > 1 {
				verr.Add(e, "Payload type is array but HTTP endpoint defines multiple parameters. At most one parameter must be defined and it must be an array.")
			}
		case *Map:
			if len(*AsObject(pparams.Type))+len(*AsObject(qparams.Type)) > 1 {
				verr.Add(e, "Payload type is map but HTTP endpoint defines multiple parameters. At most one query string parameter must be defined and it must be a map.")
			}
		}
	}
	return verr
}

// validateMappedDefault reports a default authored on an HTTP path parameter,
// query parameter, header or cookie that differs from the default of the
// object payload attribute it maps. Object payloads own field defaults:
// generated decoders apply the payload default, so a different transport
// default would be ignored at runtime while still appearing in OpenAPI.
func (e *HTTPEndpointExpr) validateMappedDefault(verr *eval.ValidationErrors, kind, name string, a *AttributeExpr) {
	if a.DefaultValue == nil {
		return
	}
	payloadDefault := e.MethodExpr.Payload.GetDefault(name)
	if payloadDefault == nil {
		verr.Add(e, "%s %q defines a default value but payload attribute %q does not; object payloads own field defaults, so define the Default on the payload attribute instead.", kind, name, name)
		return
	}
	if !reflect.DeepEqual(a.DefaultValue, payloadDefault) {
		verr.Add(e, "%s %q default value %#v differs from payload attribute %q default value %#v; object payloads own field defaults, so define the Default on the payload attribute only.", kind, name, a.DefaultValue, name, payloadDefault)
	}
}
