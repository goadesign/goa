package codegen

import (
	"strings"

	"goa.design/goa/v3/expr"
)

// Goify makes a valid Go identifier out of any string. It does that by removing
// any non letter and non digit character and by making sure the first character
// is a letter or "_". Goify produces a "CamelCase" version of the string, if
// firstUpper is true the first character of the identifier is uppercase
// otherwise it's lowercase.
func Goify(str string, firstUpper bool) string {
	// Optimize trivial case
	if str == "" {
		return ""
	}

	// Remove optional suffix that defines corresponding transport specific
	// name.
	idx := strings.Index(str, ":")
	if idx > 0 {
		str = str[:idx]
	}

	str = CamelCase(str, firstUpper, true)
	if str == "" {
		// All characters are invalid. Produce a default value.
		if firstUpper {
			return "Val"
		}
		return "val"
	}
	return fixReservedGo(str)
}

// GoifyAtt honors any struct:field:name meta set on the attribute and calls
// Goify with the tag value if present or the given name otherwise.
func GoifyAtt(att *expr.AttributeExpr, name string, upper bool) string {
	if tname, ok := att.Meta["struct:field:name"]; ok {
		if len(tname) > 0 {
			name = tname[0]
		}
	}
	return Goify(name, upper)
}

// UnionValTypeName returns the Go type name of the interface and method used to
// type the union.
func UnionValTypeName(unionName string) string {
	return Goify(unionName+"Val", false)
}

// fixReservedGo appends an underscore on to Go reserved keywords.
func fixReservedGo(w string) string {
	if reservedGo[w] {
		w += "_"
	}
	return w
}

var (
	// reserved golang keywords and package names
	reservedGo = map[string]bool{
		// predeclared types
		"bool":       true,
		"byte":       true,
		"complex128": true,
		"complex64":  true,
		"float32":    true,
		"float64":    true,
		"int":        true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"int8":       true,
		"rune":       true,
		"string":     true,
		"uint":       true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uint8":      true,
		"uintptr":    true,

		// reserved keywords
		"break":       true,
		"case":        true,
		"chan":        true,
		"const":       true,
		"continue":    true,
		"default":     true,
		"defer":       true,
		"delete":      true,
		"else":        true,
		"fallthrough": true,
		"for":         true,
		"func":        true,
		"go":          true,
		"goto":        true,
		"if":          true,
		"import":      true,
		"interface":   true,
		"map":         true,
		"package":     true,
		"range":       true,
		"return":      true,
		"select":      true,
		"struct":      true,
		"switch":      true,
		"type":        true,
		"var":         true,

		// predeclared constants
		"true":  true,
		"false": true,
		"iota":  true,
		"nil":   true,

		// stdlib and goa packages used by generated code
		"fmt":  true,
		"http": true,
		"json": true,
		"os":   true,
		"url":  true,
		"time": true,
	}
)
