// This file lets an HTTP service name fields of existing types on the wire
// without changing the shared service declarations or their stored JSON.
package dsl

import (
	"strings"
	"unicode"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// BodyFieldName maps one field of an existing named object type to an HTTP JSON
// property in the current service. The mapping must have the form
// "attributeName:wireName". All request, response and error bodies in that HTTP
// service use the mapping, including nested occurrences. Service values, other
// services and protobuf declarations retain their original names.
//
// BodyFieldName must appear in a service HTTP expression. It matches the exact
// named declaration: a type that extends another type needs its own mapping.
// Repeated mappings, unknown fields and colliding wire names are rejected.
// An explicit mapping replaces the name in a complete JSON tag on the HTTP
// copy, preserving its encoding options. The original type's tag is unchanged.
func BodyFieldName(typ expr.UserType, mapping string) {
	svc, ok := eval.Current().(*expr.HTTPServiceExpr)
	if !ok {
		eval.IncompatibleDSL()
		return
	}
	field, wire, found := strings.Cut(mapping, ":")
	if !found || field == "" || wire == "" || strings.Contains(wire, ":") {
		eval.ReportError("BodyFieldName requires field:wireName with two nonempty names")
		return
	}
	if wire == "-" {
		eval.ReportError("BodyFieldName wire name %q cannot be represented by a JSON struct tag", wire)
		return
	}
	for _, character := range wire {
		// Go's JSON encoder accepts letters, digits and this punctuation in
		// tag names. Other characters make it ignore the selected name.
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", character) {
			eval.ReportError("BodyFieldName wire name %q cannot be represented by a JSON struct tag", wire)
			return
		}
	}
	if typ == nil || !expr.IsObject(typ) {
		eval.ReportError("BodyFieldName requires an existing named object type")
		return
	}
	if svc.BodyFieldNames == nil {
		svc.BodyFieldNames = make(map[expr.UserType]map[string]string)
	}
	origin := typ.Origin()
	fields := svc.BodyFieldNames[origin]
	if fields == nil {
		fields = make(map[string]string)
		svc.BodyFieldNames[origin] = fields
	}
	if _, exists := fields[field]; exists {
		eval.ReportError("BodyFieldName already maps field %q of %s", field, typ.Name())
		return
	}
	fields[field] = wire
}
