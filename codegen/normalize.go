package codegen

import (
	"strings"

	"goa.design/goa/v3/expr"
)

// NormalizeRoot applies the only sanctioned design mutation that may happen
// after the DSL has been evaluated and finalized: it wraps the raw object
// payload, result and streaming types of every service method into
// synthesized user types named after the method. Every code generation layer
// (service, transports, OpenAPI, example, CLI and type conversion) relies on
// method payload and result types being named, so the wrapping must happen
// before any generator reads the design.
//
// NormalizeRoot is idempotent: already wrapped methods are left untouched. It
// must run after the prepare plugins so that plugin contributed endpoints are
// normalized too, and before any generator runs. Past this point the design
// expression tree is read-only for all generators; the purity test in
// codegen/generator enforces that contract.
//
// The synthesized type names are resolved against a name scope seeded with
// the exact same registrations the service generator performs when it
// collects the service user types (see codegen/service analyze) so that
// wrapping up front produces the very same type names the service generator
// produced when it owned the wrapping.
func NormalizeRoot(r *expr.RootExpr) {
	for _, svc := range r.Services {
		normalizeService(svc)
	}
}

// normalizeService wraps the raw object method types of svc into synthesized
// user types. The name scope fed to PeekUnique is seeded by replaying the
// scope side effects of the service generator analysis in the same order:
// reserved identifiers and package name first, then the user types reachable
// from the service errors and from each method payload, streaming payload,
// result, streaming result, projected result types and method errors.
func normalizeService(svc *expr.ServiceExpr) {
	scope := NewNameScope()
	scope.Unique("Use")       // Reserve "Use" for Endpoints struct Use method.
	scope.Unique("websocket") // Reserve "websocket" to avoid collision with gorilla/websocket
	scope.HashedUnique(svc, strings.ToLower(Goify(svc.Name, false)), "svc")
	seen := make(map[string]struct{})
	for _, er := range svc.Errors {
		seedTypeNames(er.AttributeExpr, scope, seen)
	}
	seedMethodAtt := func(att *expr.AttributeExpr) {
		if att == nil {
			return
		}
		if ut, ok := att.Type.(expr.UserType); ok {
			att = ut.Attribute()
		}
		seedTypeNames(att, scope, seen)
	}
	seenProjected := make(map[string]struct{})
	for _, m := range svc.Methods {
		seedMethodAtt(m.Payload)
		seedMethodAtt(m.StreamingPayload)
		seedMethodAtt(m.Result)
		if m.HasMixedResults() {
			seedMethodAtt(m.StreamingResult)
		}
		if hasResultTypeExpr(m.Result, make(map[string]struct{})) {
			seedProjectedNames(expr.DupAtt(m.Result), m.Result, scope, seenProjected)
		}
		for _, er := range m.Errors {
			seedTypeNames(er.AttributeExpr, scope, seen)
		}
	}
	wrap := func(att *expr.AttributeExpr, name, id string) {
		if att == nil {
			return
		}
		if _, ok := att.Type.(*expr.Object); !ok {
			return
		}
		att.Type = &expr.UserTypeExpr{
			AttributeExpr: expr.DupAtt(att),
			TypeName:      scope.PeekUnique(name),
			UID:           id,
		}
	}
	for _, m := range svc.Methods {
		name := Goify(m.Name, true)
		wrap(m.Payload, name+"Payload", svc.Name+"#"+name+"Payload")
		wrap(m.StreamingPayload, name+"StreamingPayload", svc.Name+"#"+name+"StreamingPayload")
		wrap(m.Result, name+"Result", svc.Name+"#"+name+"Result")
		if m.HasMixedResults() {
			wrap(m.StreamingResult, name+"StreamingResult", svc.Name+"#"+name+"StreamingResult")
		}
	}
}

// seedTypeNames mirrors the name scope side effects of the service generator
// user type collection (collectTypes in codegen/service): every user type
// reachable from at reserves its Go type name, the names referenced by its
// type definition and its type reference, in the same order. The returned
// strings are discarded, only the scope registrations matter.
func seedTypeNames(at *expr.AttributeExpr, scope *NameScope, seen map[string]struct{}) {
	if at == nil || at.Type == expr.Empty {
		return
	}
	switch dt := at.Type.(type) {
	case expr.UserType:
		if _, ok := seen[dt.ID()]; ok {
			return
		}
		scope.GoTypeName(at)
		scope.GoTypeDef(dt.Attribute(), false, true)
		scope.GoTypeRef(at)
		seen[dt.ID()] = struct{}{}
		seedTypeNames(dt.Attribute(), scope, seen)
	case *expr.Object:
		for _, nat := range *dt {
			seedTypeNames(nat.Attribute, scope, seen)
		}
	case *expr.Array:
		seedTypeNames(dt.ElemType, scope, seen)
	case *expr.Map:
		seedTypeNames(dt.KeyType, scope, seen)
		seedTypeNames(dt.ElemType, scope, seen)
	case *expr.Union:
		for _, nat := range dt.Values {
			seedTypeNames(nat.Attribute, scope, seen)
		}
	}
}

// seedProjectedNames mirrors the name scope side effects of the projected
// type collection (collectProjectedTypes and the view conversion builders in
// codegen/service): projected is a detached copy of the method result
// attribute whose user types are renamed with the "View" suffix while
// traversing, and every result type with views reserves the projected and
// original type names in the service scope, children first.
func seedProjectedNames(projected, att *expr.AttributeExpr, scope *NameScope, seen map[string]struct{}) {
	switch pt := projected.Type.(type) {
	case expr.UserType:
		dt := att.Type.(expr.UserType)
		if _, ok := seen[dt.ID()]; ok {
			return
		}
		seen[dt.ID()] = struct{}{}
		pt.Rename(pt.Name() + "View")
		seedProjectedNames(pt.Attribute(), dt.Attribute(), scope, seen)
		if rt, ok := pt.(*expr.ResultTypeExpr); ok && len(rt.Views) > 0 {
			if parr := expr.AsArray(pt); parr != nil {
				scope.GoTypeName(parr.ElemType)
			}
			scope.GoTypeName(projected)
			scope.GoTypeName(att)
		}
	case *expr.Array:
		seedProjectedNames(pt.ElemType, att.Type.(*expr.Array).ElemType, scope, seen)
	case *expr.Map:
		dt := att.Type.(*expr.Map)
		seedProjectedNames(pt.KeyType, dt.KeyType, scope, seen)
		seedProjectedNames(pt.ElemType, dt.ElemType, scope, seen)
	case *expr.Object:
		dt := att.Type.(*expr.Object)
		for _, n := range *pt {
			seedProjectedNames(n.Attribute, dt.Attribute(n.Name), scope, seen)
		}
	case *expr.Union:
		dt := att.Type.(*expr.Union)
		for i, n := range pt.Values {
			seedProjectedNames(n.Attribute, dt.Values[i].Attribute, scope, seen)
		}
	}
}

// hasResultTypeExpr reports whether att transitively references a result type
// expression. It mirrors hasResultType in codegen/service which decides
// whether the service generator collects projected types for a method result.
func hasResultTypeExpr(att *expr.AttributeExpr, seen map[string]struct{}) bool {
	if _, ok := att.Type.(*expr.ResultTypeExpr); ok {
		return true
	}
	switch a := att.Type.(type) {
	case expr.UserType:
		if _, ok := seen[a.ID()]; ok {
			return false
		}
		seen[a.ID()] = struct{}{}
		return hasResultTypeExpr(a.Attribute(), seen)
	case *expr.Array:
		return hasResultTypeExpr(a.ElemType, seen)
	case *expr.Map:
		return hasResultTypeExpr(a.KeyType, seen) || hasResultTypeExpr(a.ElemType, seen)
	case *expr.Object:
		for _, nat := range *a {
			if hasResultTypeExpr(nat.Attribute, seen) {
				return true
			}
		}
	case *expr.Union:
		for _, nat := range a.Values {
			if hasResultTypeExpr(nat.Attribute, seen) {
				return true
			}
		}
	}
	return false
}
