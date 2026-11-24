// This file validates prepared URL bindings against a service method's payload.
// HTTP, JSON-RPC, and transport plugins use the same path and query rules before
// generating decoders; unsupported types and missing fields become design errors.
package expr

import (
	"fmt"
	"reflect"

	"goa.design/goa/v3/eval"
)

type (
	// shadowField records the effective type, default value, and metadata of an
	// AttributeExpr during simulated inheritance resolution so validation never
	// mutates the unfinalized design AST.
	shadowField struct {
		typ          DataType
		defaultValue any
		meta         MetaExpr
	}

	// payloadDefaultResolver simulates AttributeExpr finalization in RootExpr
	// WalkSets order so pre-Finalize HTTP validation observes the exact field
	// defaults that GetDefault returns after Finalize.
	payloadDefaultResolver struct {
		finalized           map[*AttributeExpr]bool
		inheritanceResolved map[*AttributeExpr]bool
		resolving           map[*AttributeExpr]bool
		objects             map[*Object][]*NamedAttributeExpr
		fields              map[*AttributeExpr]*shadowField
	}
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

// newPayloadDefaultResolver simulates AttributeExpr finalization across the
// design's types, result types, and methods following RootExpr.WalkSets order
// without mutating the design AST.
func newPayloadDefaultResolver(e *HTTPEndpointExpr) *payloadDefaultResolver {
	r := &payloadDefaultResolver{
		finalized:           make(map[*AttributeExpr]bool),
		inheritanceResolved: make(map[*AttributeExpr]bool),
		resolving:           make(map[*AttributeExpr]bool),
		objects:             make(map[*Object][]*NamedAttributeExpr),
		fields:              make(map[*AttributeExpr]*shadowField),
	}
	if e.MethodExpr.Service != nil && e.MethodExpr.Service.design != nil {
		design := e.MethodExpr.Service.design
		for _, t := range design.Types {
			r.finalize(t.Attribute())
		}
		for _, rt := range design.ResultTypes {
			r.finalize(rt.Attribute())
		}
		for _, svc := range design.Services {
			for _, m := range svc.Methods {
				r.finalizeMethod(m)
			}
		}
	}
	r.finalize(e.MethodExpr.Payload)
	return r
}

// transportDefaultAttr pairs the resolved payload field type with any metadata
// authored directly on the HTTP mapping, matching initAttrFromDesign.
func transportDefaultAttr(mapped, payload *AttributeExpr) *AttributeExpr {
	if mapped.Meta == nil {
		return payload
	}
	copied := *payload
	copied.Meta = mapped.Meta
	return &copied
}

// findNamedAttr returns the attribute for name in entries, or nil when absent.
func findNamedAttr(entries []*NamedAttributeExpr, name string) *AttributeExpr {
	for _, nat := range entries {
		if nat.Name == name {
			return nat.Attribute
		}
	}
	return nil
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
func (e *HTTPEndpointExpr) validateMappedDefaults(
	verr *eval.ValidationErrors,
	kind string,
	mapped *MappedAttributeExpr,
) {
	var resolver *payloadDefaultResolver
	WalkMappedAttr(mapped, func(name, elem string, a *AttributeExpr) error { // nolint: errcheck
		if a.DefaultValue == nil {
			return nil
		}
		if resolver == nil {
			resolver = newPayloadDefaultResolver(e)
		}
		field, payloadDefault, found := resolver.fieldDefault(e.MethodExpr.Payload, name)
		if !found {
			return nil
		}
		if payloadDefault == nil {
			verr.Add(e, "%s %q defines a default value but payload attribute %q does not; object payloads own field defaults, so define the Default on the payload attribute instead.", kind, elem, name)
			return nil
		}
		leftAttr := transportDefaultAttr(a, field)
		equal, comparable := defaultValuesEqual(
			leftAttr,
			field,
			reflect.ValueOf(a.DefaultValue),
			reflect.ValueOf(payloadDefault),
		)
		if !comparable && (!IsArray(field.Type) || reflect.DeepEqual(a.DefaultValue, payloadDefault)) {
			return nil
		}
		if !equal {
			verr.Add(e, "%s %q default value %#v differs from payload attribute %q default value %#v; object payloads own field defaults, so remove the Default from the %s or make it match.", kind, elem, a.DefaultValue, name, payloadDefault, kind)
		}
		return nil
	})
}

// fieldDefault returns the resolved field attribute and default value that
// payload owns under name after simulated finalization, matching GetDefault.
func (r *payloadDefaultResolver) fieldDefault(
	payload *AttributeExpr,
	name string,
) (*AttributeExpr, any, bool) {
	if payload == nil || !r.isObject(r.effectiveType(payload)) {
		return nil, nil, false
	}
	att := findNamedAttr(r.objectEntries(payload), name)
	if att == nil {
		return nil, nil, false
	}
	sf := r.fieldFor(att)
	field := &AttributeExpr{
		Type: sf.typ,
		Meta: sf.meta,
	}
	value := sf.defaultValue
	if value == nil {
		if ut, ok := sf.typ.(UserType); ok && !r.isObject(ut) && ut.Attribute() != nil {
			value = r.fieldFor(ut.Attribute()).defaultValue
		}
	}
	return field, value, true
}

// finalizeMethod simulates MethodExpr.Finalize on an earlier method in the
// design.
func (r *payloadDefaultResolver) finalizeMethod(m *MethodExpr) {
	if m == nil {
		return
	}
	r.finalize(m.Payload)
	r.finalize(m.StreamingPayload)
	if m.StreamingResult != nil {
		r.finalize(m.StreamingResult)
	}
	r.finalize(m.Result)
}

// finalize simulates AttributeExpr.Finalize without mutating AST nodes.
func (r *payloadDefaultResolver) finalize(a *AttributeExpr) {
	if a == nil || a.finalized || r.finalized[a] {
		return
	}
	r.resolveInheritance(a)
	if r.finalized[a] {
		return
	}
	r.finalized[a] = true
	typ := r.effectiveType(a)
	if ut, ok := typ.(UserType); ok {
		r.finalize(ut.Attribute())
	}
	switch {
	case r.isObject(typ):
		for _, nat := range r.objectEntries(a) {
			r.finalize(nat.Attribute)
		}
	case IsUnion(typ):
		for _, nat := range AsUnion(typ).Values {
			r.finalize(nat.Attribute)
		}
	case IsArray(typ):
		r.finalize(AsArray(typ).ElemType)
	case IsMap(typ):
		m := AsMap(typ)
		r.finalize(m.ElemType)
		r.finalize(m.KeyType)
	}
}

// resolveInheritance simulates AttributeExpr.resolveInheritance, applying
// References in their current state before resolving and merging Bases.
func (r *payloadDefaultResolver) resolveInheritance(a *AttributeExpr) {
	if a == nil || a.finalized || r.finalized[a] || r.inheritanceResolved[a] || r.resolving[a] {
		return
	}
	r.resolving[a] = true
	defer delete(r.resolving, a)

	typ := r.effectiveType(a)
	if ut, ok := typ.(UserType); ok && ut.Attribute() != nil {
		r.resolveInheritance(ut.Attribute())
	}
	if r.isObject(typ) {
		for _, ref := range a.References {
			if ru, ok := ref.(UserType); ok {
				r.inherit(a, ru.Attribute())
			}
		}
		for _, base := range a.Bases {
			ru, ok := base.(UserType)
			if !ok || ru.Attribute() == nil {
				continue
			}
			r.resolveInheritance(ru.Attribute())
			r.merge(a, ru.Attribute())
		}
	}
	r.inheritanceResolved[a] = true
}

// merge simulates AttributeExpr.Merge onto shadow object entries.
func (r *payloadDefaultResolver) merge(a, other *AttributeExpr) {
	if a == nil || other == nil {
		return
	}
	if !r.isObject(r.effectiveType(a)) || !r.isObject(r.effectiveType(other)) {
		return
	}
	right := r.objectEntries(other)
	if r.effectiveType(a) == Empty && len(right) > 0 {
		r.fieldFor(a).typ = &Object{}
	}
	for _, nat := range right {
		r.setObjectEntry(a, nat.Name, nat.Attribute)
	}
}

// inherit simulates AttributeExpr.Inherit onto shadow fields.
func (r *payloadDefaultResolver) inherit(a, parent *AttributeExpr) {
	if !r.shouldInherit(a, parent) {
		return
	}
	if r.effectiveType(a) == Empty && len(r.objectEntries(parent)) > 0 {
		r.fieldFor(a).typ = &Object{}
	}
	r.inheritRecursive(a, parent, make(map[*AttributeExpr]struct{}))
}

// inheritRecursive copies missing default values and types from matching fields
// on parent into a's shadow fields.
func (r *payloadDefaultResolver) inheritRecursive(
	a, parent *AttributeExpr,
	seen map[*AttributeExpr]struct{},
) {
	if !r.shouldInherit(a, parent) {
		return
	}
	parentEntries := r.objectEntries(parent)
	for _, nat := range r.objectEntries(a) {
		patt := findNamedAttr(parentEntries, nat.Name)
		if patt == nil {
			continue
		}
		attField := r.fieldFor(nat.Attribute)
		pattField := r.fieldFor(patt)
		if attField.defaultValue == nil && pattField.defaultValue != nil {
			attField.defaultValue = pattField.defaultValue
			if !hasCustomFieldType(&AttributeExpr{Type: attField.typ, Meta: attField.meta}) &&
				hasCustomFieldType(&AttributeExpr{Type: pattField.typ, Meta: pattField.meta}) {
				attField.meta = pattField.meta
			}
		}
		if attField.typ == nil {
			attField.typ = pattField.typ
		} else if r.shouldInherit(nat.Attribute, patt) {
			if _, ok := seen[nat.Attribute]; ok {
				continue
			}
			seen[nat.Attribute] = struct{}{}
			for _, child := range r.objectEntries(nat.Attribute) {
				if childParent := findNamedAttr(r.objectEntries(patt), child.Name); childParent != nil {
					r.inheritRecursive(child.Attribute, childParent, seen)
				}
			}
		}
	}
}

// shouldInherit reports whether both attributes are objects in the shadow model.
func (r *payloadDefaultResolver) shouldInherit(a, parent *AttributeExpr) bool {
	return a != nil && r.isObject(r.effectiveType(a)) &&
		parent != nil && r.isObject(r.effectiveType(parent))
}

// fieldFor returns the mutable shadow record for an AST AttributeExpr.
func (r *payloadDefaultResolver) fieldFor(a *AttributeExpr) *shadowField {
	if sf, ok := r.fields[a]; ok {
		return sf
	}
	sf := &shadowField{
		typ:          a.Type,
		defaultValue: a.DefaultValue,
		meta:         a.Meta,
	}
	r.fields[a] = sf
	return sf
}

// effectiveType returns the simulated DataType of a.
func (r *payloadDefaultResolver) effectiveType(a *AttributeExpr) DataType {
	if a == nil {
		return nil
	}
	return r.fieldFor(a).typ
}

// asObject unwraps UserType chains in the shadow model to their underlying Object.
func (r *payloadDefaultResolver) asObject(dt DataType) *Object {
	seen := make(map[UserType]struct{})
	for {
		switch t := dt.(type) {
		case UserType:
			if _, ok := seen[t]; ok || t.Attribute() == nil {
				return nil
			}
			seen[t] = struct{}{}
			dt = r.effectiveType(t.Attribute())
		case *Object:
			return t
		default:
			return nil
		}
	}
}

// isObject reports whether dt resolves to an Object in the shadow model.
func (r *payloadDefaultResolver) isObject(dt DataType) bool {
	return r.asObject(dt) != nil
}

// objectEntries returns the shadow named-attribute slice for a's Object.
func (r *payloadDefaultResolver) objectEntries(a *AttributeExpr) []*NamedAttributeExpr {
	obj := r.asObject(r.effectiveType(a))
	if obj == nil {
		return nil
	}
	if entries, ok := r.objects[obj]; ok {
		return entries
	}
	entries := make([]*NamedAttributeExpr, len(*obj))
	copy(entries, *obj)
	r.objects[obj] = entries
	return entries
}

// setObjectEntry replaces or appends name in a's shadow Object slice.
func (r *payloadDefaultResolver) setObjectEntry(a *AttributeExpr, name string, attr *AttributeExpr) {
	obj := r.asObject(r.effectiveType(a))
	entries := r.objectEntries(a)
	for i, nat := range entries {
		if nat.Name == name {
			entries[i] = &NamedAttributeExpr{Name: name, Attribute: attr}
			return
		}
	}
	r.objects[obj] = append(entries, &NamedAttributeExpr{Name: name, Attribute: attr})
}
