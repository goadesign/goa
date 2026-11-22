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
			WalkMappedAttr(pparams, func(name, _ string, _ *AttributeExpr) error { // nolint: errcheck
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Path parameter %q not found in payload.", name)
				}
				return nil
			})
			WalkMappedAttr(qparams, func(name, _ string, _ *AttributeExpr) error { // nolint: errcheck
				if e.MethodExpr.Payload.Find(name) == nil {
					verr.Add(e, "Query string parameter %q not found in payload.", name)
				}
				return nil
			})
			e.validateMappedDefaults(verr, "path parameter", e.PathParams())
			e.validateMappedDefaults(verr, "query parameter", e.QueryParams())
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

// validateMappedDefaults reports defaults authored on HTTP path parameters,
// query parameters, headers or cookies that differ from the default of the
// object payload attribute they map. Object payloads own field defaults:
// generated decoders apply the payload default, so a different transport
// default would be ignored at runtime while still appearing in OpenAPI.
//
// mapped must be the authored mapping, before initAttr copies payload
// defaults onto it, so only defaults written on the transport are checked.
// Attributes missing from the payload are reported by the callers.
func (e *HTTPEndpointExpr) validateMappedDefaults(verr *eval.ValidationErrors, kind string, mapped *MappedAttributeExpr) {
	WalkMappedAttr(mapped, func(name, elem string, a *AttributeExpr) error { // nolint: errcheck
		if a.DefaultValue == nil {
			return nil
		}
		field, payloadDefault, found := payloadFieldDefault(e.MethodExpr.Payload, name, make(map[*AttributeExpr]struct{}))
		if !found {
			return nil
		}
		if payloadDefault == nil {
			verr.Add(e, "%s %q defines a default value but payload attribute %q does not; object payloads own field defaults, so define the Default on the payload attribute instead.", kind, elem, name)
			return nil
		}
		if !defaultValuesEqual(field.Type, reflect.ValueOf(a.DefaultValue), reflect.ValueOf(payloadDefault)) {
			verr.Add(e, "%s %q default value %#v differs from payload attribute %q default value %#v; object payloads own field defaults, so remove the Default from the %s or make it match.", kind, elem, a.DefaultValue, name, payloadDefault, kind)
		}
		return nil
	})
}

// payloadFieldDefault returns the field the object attribute parent declares
// or inherits under name and the default it carries once Finalize resolves
// inheritance. Validation runs before Finalize, so it follows the same order
// as resolveInheritance on parent and on the attribute of its user type:
// Bases replace declared fields with later Bases taking precedence, and
// References only fill a missing field type or default. A field typed by a
// named non-object type falls back to that type's default, like GetDefault.
// found is false when neither parent nor its Bases declare the field. seen
// holds the attributes being resolved to stop on cycles.
func payloadFieldDefault(parent *AttributeExpr, name string, seen map[*AttributeExpr]struct{}) (field *AttributeExpr, value any, found bool) {
	if _, ok := seen[parent]; ok {
		return nil, nil, false
	}
	seen[parent] = struct{}{}
	defer delete(seen, parent)
	for i := len(parent.Bases) - 1; i >= 0; i-- {
		base, ok := parent.Bases[i].(UserType)
		if !ok {
			continue
		}
		if field, value, found := payloadFieldDefault(base.Attribute(), name, seen); found {
			return field, value, true
		}
	}
	if ut, ok := parent.Type.(UserType); ok {
		field, value, found = payloadFieldDefault(ut.Attribute(), name, seen)
	} else if object := AsObject(parent.Type); object != nil {
		field = object.Attribute(name)
		if field != nil {
			value, found = field.DefaultValue, true
		}
	}
	if !found {
		return nil, nil, false
	}
	for _, reference := range parent.References {
		ref, ok := reference.(UserType)
		if !ok {
			continue
		}
		refField, refValue, refFound := payloadFieldDefault(ref.Attribute(), name, seen)
		if !refFound {
			continue
		}
		if field.Type == nil {
			field = &AttributeExpr{Type: refField.Type}
		}
		if value == nil {
			value = refValue
		}
	}
	if value == nil {
		if ut, ok := field.Type.(UserType); ok && !IsObject(ut) {
			value = ut.Attribute().DefaultValue
		}
	}
	return field, value, true
}
