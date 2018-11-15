package codegen

import (
	"fmt"
	"sort"
	"strings"

	"goa.design/goa/design"
)

// GoNativeTypeName returns the Go built-in type corresponding to the given
// primitive type. GoNativeType panics if t is not a primitive type.
func GoNativeTypeName(t design.DataType) string {
	switch t.Kind() {
	case design.BooleanKind:
		return "bool"
	case design.IntKind:
		return "int"
	case design.Int32Kind:
		return "int32"
	case design.Int64Kind:
		return "int64"
	case design.UIntKind:
		return "uint"
	case design.UInt32Kind:
		return "uint32"
	case design.UInt64Kind:
		return "uint64"
	case design.Float32Kind:
		return "float32"
	case design.Float64Kind:
		return "float64"
	case design.StringKind:
		return "string"
	case design.BytesKind:
		return "[]byte"
	case design.AnyKind:
		return "interface{}"
	default:
		panic(fmt.Sprintf("cannot compute native Go type for %T", t)) // bug
	}
}

// AttributeTags computes the struct field tags from its meta if any.
func AttributeTags(parent, att *design.AttributeExpr) string {
	var elems []string
	keys := make([]string, len(att.Meta))
	i := 0
	for k := range att.Meta {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := att.Meta[key]
		if strings.HasPrefix(key, "struct:tag:") {
			name := key[11:]
			value := strings.Join(val, ",")
			elems = append(elems, fmt.Sprintf("%s:\"%s\"", name, value))
		}
	}
	if len(elems) > 0 {
		return " `" + strings.Join(elems, " ") + "`"
	}
	return ""
}
