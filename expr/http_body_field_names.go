// This file applies explicit field names to HTTP-owned copies after body
// construction. Shared declarations remain unchanged, and generated conversion
// code can still match each wire property to its original service field.
package expr

import (
	"sort"
	"strings"

	"goa.design/goa/v3/eval"
)

// validateBodyFieldNames checks completed named types so inherited fields are
// available before a service's mapping can be used by HTTP generation.
func (svc *HTTPServiceExpr) validateBodyFieldNames(verr *eval.ValidationErrors) {
	types := make([]UserType, 0, len(svc.BodyFieldNames))
	for typ := range svc.BodyFieldNames {
		types = append(types, typ)
	}
	sort.Slice(types, func(i, j int) bool { return types[i].Name() < types[j].Name() })
	for _, typ := range types {
		fields := svc.BodyFieldNames[typ]
		shape := DupAtt(typ.Attribute())
		shape.Finalize()
		for _, name := range sortedBodyFieldNames(fields) {
			if shape.Find(name) == nil {
				verr.Add(svc, "BodyFieldName field %q does not exist in %s", name, typ.Name())
			}
		}
		seen := make(map[string]string)
		for _, field := range *AsObject(shape.Type) {
			wire := field.Name
			if tags := field.Attribute.Meta["struct:tag:json"]; len(tags) > 0 {
				name, _, _ := strings.Cut(strings.Join(tags, ","), ",")
				if name != "" {
					wire = name
				}
			}
			if mapped, ok := fields[field.Name]; ok {
				wire = mapped
			} else if wire == "-" {
				continue
			}
			if prior, ok := seen[wire]; ok {
				verr.Add(svc, "BodyFieldName collision in %s: %q and %q both use %q", typ.Name(), prior, field.Name, wire)
			}
			seen[wire] = field.Name
		}
	}
	if len(svc.BodyFieldNames) > 0 && len(verr.Errors) == 0 {
		svc.validateMappedBodyUnions(verr)
	}
}

// mapBodyFieldNames copies one completed HTTP body and changes only the names
// selected by this service. original identifies the service declaration before
// HTTP construction replaced the root with its request or response wrapper.
func (svc *HTTPServiceExpr) mapBodyFieldNames(body *AttributeExpr, original UserType) *AttributeExpr {
	if len(svc.BodyFieldNames) == 0 || body.Type == Empty || !svc.bodyTypeHasMapping(body.Type, original, make(map[UserType]bool)) {
		return body
	}
	copy := DupAtt(body)
	seen := make(map[UserType]bool)
	affected := make(map[UserType]bool)
	svc.collectMappedBodyTypes(copy.Type, original, affected, seen)
	clear(seen)
	svc.mapBodyType(copy.Type, original, affected, seen)
	return copy
}

// bodyFieldNameSource records the exact declaration before HTTP construction
// replaces it with a wrapper. An explicit body type owns its own field names.
func bodyFieldNameSource(body, service *AttributeExpr) UserType {
	source := service
	if body != nil {
		source = body
		if selected := body.Meta["origin:attribute"]; len(selected) > 0 {
			source = service.Find(selected[0])
		}
	}
	if named, ok := source.Type.(UserType); ok {
		return named.Origin()
	}
	return nil
}

// mapBodyType follows typed fields, collection values and union branches. It
// leaves scalar values and map keys untouched and visits recursive types once.
func (svc *HTTPServiceExpr) mapBodyType(typ DataType, root UserType, affected, seen map[UserType]bool) {
	switch actual := typ.(type) {
	case UserType:
		if seen[actual] || !affected[actual] {
			return
		}
		seen[actual] = true
		origin := actual.Origin()
		if root != nil {
			origin = root
		}
		attribute := actual.Attribute()
		if object := AsObject(actual); object != nil {
			for _, field := range *object {
				if wire, ok := svc.BodyFieldNames[origin][field.Name]; ok {
					field.Name += ":" + wire
					if tags := field.Attribute.Meta["struct:tag:json"]; len(tags) > 0 {
						_, options, hasOptions := strings.Cut(strings.Join(tags, ","), ",")
						tag := wire
						if hasOptions {
							tag += "," + options
						}
						field.Attribute.Meta["struct:tag:json"] = []string{tag}
					}
				}
			}
		}
		// A type containing mapped fields needs a separate HTTP identity so
		// another service cannot reuse its changed schema. Unrelated types
		// retain their existing identities and generated examples.
		identity := newExampleIdentity(httpMappedBodyTypeExampleKind, []byte(svc.Name()), []byte(actual.ID()))
		switch named := actual.(type) {
		case *UserTypeExpr:
			named.origin = nil
			named.UID = "generated:" + identity.Seed()
			named.exampleIdentity = identity
		case *ResultTypeExpr:
			named.origin = nil
			named.UID = "generated:" + identity.Seed()
			named.exampleIdentity = identity
		}
		svc.mapBodyType(attribute.Type, nil, affected, seen)
	case *Object:
		for _, field := range *actual {
			svc.mapBodyType(field.Attribute.Type, nil, affected, seen)
		}
	case *Array:
		svc.mapBodyType(actual.ElemType.Type, nil, affected, seen)
	case *Map:
		svc.mapBodyType(actual.ElemType.Type, nil, affected, seen)
	case *Union:
		for _, branch := range actual.Values {
			svc.mapBodyType(branch.Attribute.Type, nil, affected, seen)
		}
	}
}

// collectMappedBodyTypes records affected declarations before any copied type
// changes its identity. Shared children therefore produce the same result
// regardless of the order in which their parent fields are visited.
func (svc *HTTPServiceExpr) collectMappedBodyTypes(typ DataType, root UserType, affected, seen map[UserType]bool) {
	switch actual := typ.(type) {
	case UserType:
		if seen[actual] {
			return
		}
		seen[actual] = true
		affected[actual] = svc.bodyTypeHasMapping(actual, root, make(map[UserType]bool))
		svc.collectMappedBodyTypes(actual.Attribute().Type, nil, affected, seen)
	case *Object:
		for _, field := range *actual {
			svc.collectMappedBodyTypes(field.Attribute.Type, nil, affected, seen)
		}
	case *Array:
		svc.collectMappedBodyTypes(actual.ElemType.Type, nil, affected, seen)
	case *Map:
		svc.collectMappedBodyTypes(actual.ElemType.Type, nil, affected, seen)
	case *Union:
		for _, branch := range actual.Values {
			svc.collectMappedBodyTypes(branch.Attribute.Type, nil, affected, seen)
		}
	}
}

// bodyTypeHasMapping checks whether a type contains a declaration selected by
// this service. It follows collection values and union branches, stopping at
// already visited declarations so recursive fields cannot cause a loop.
func (svc *HTTPServiceExpr) bodyTypeHasMapping(typ DataType, root UserType, seen map[UserType]bool) bool {
	switch actual := typ.(type) {
	case UserType:
		if seen[actual] {
			return false
		}
		seen[actual] = true
		origin := actual.Origin()
		if root != nil {
			origin = root
		}
		return len(svc.BodyFieldNames[origin]) > 0 || svc.bodyTypeHasMapping(actual.Attribute().Type, nil, seen)
	case *Object:
		for _, field := range *actual {
			if svc.bodyTypeHasMapping(field.Attribute.Type, nil, seen) {
				return true
			}
		}
	case *Array:
		return svc.bodyTypeHasMapping(actual.ElemType.Type, nil, seen)
	case *Map:
		return svc.bodyTypeHasMapping(actual.ElemType.Type, nil, seen)
	case *Union:
		for _, branch := range actual.Values {
			if svc.bodyTypeHasMapping(branch.Attribute.Type, nil, seen) {
				return true
			}
		}
	}
	return false
}

func sortedBodyFieldNames(fields map[string]string) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
