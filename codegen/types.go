package codegen

import (
	"fmt"
	"sort"
	"strings"

	"goa.design/goa/v3/expr"
)

// GoNativeTypeName returns the Go built-in type corresponding to the given
// primitive type. GoNativeType panics if t is not a primitive type.
func GoNativeTypeName(t expr.DataType) string {
	switch t.Kind() {
	case expr.BooleanKind:
		return "bool"
	case expr.IntKind:
		return "int"
	case expr.Int32Kind:
		return "int32"
	case expr.Int64Kind:
		return "int64"
	case expr.UIntKind:
		return "uint"
	case expr.UInt32Kind:
		return "uint32"
	case expr.UInt64Kind:
		return "uint64"
	case expr.Float32Kind:
		return "float32"
	case expr.Float64Kind:
		return "float64"
	case expr.StringKind:
		return "string"
	case expr.BytesKind:
		return "[]byte"
	case expr.AnyKind:
		return "any"
	default:
		panic(fmt.Sprintf("cannot compute native Go type for %T", t)) // bug
	}
}

// AttributeTags computes the struct field tags from its metadata if any.
func AttributeTags(_, att *expr.AttributeExpr) string {
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
			if key == "struct:tag:json:name" {
				continue
			}
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

// AttributeTagsWithName computes the struct field tags from its metadata,
// interpreting the "struct:tag:json:name" key when present.
//
// The "struct:tag:json" meta key always takes precedence and is treated as a
// complete tag override value. When only "struct:tag:json:name" is set, Goa
// computes the json tag and appends ",omitempty" when the field is not
// required by its parent object.
func AttributeTagsWithName(parent *expr.AttributeExpr, fieldName string, att *expr.AttributeExpr) string {
	if att == nil || len(att.Meta) == 0 {
		return ""
	}
	tags := make(map[string]string)
	var jsonName string
	keys := make([]string, 0, len(att.Meta))
	for k := range att.Meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := att.Meta[key]
		if !strings.HasPrefix(key, "struct:tag:") {
			continue
		}
		switch key {
		case "struct:tag:json:name":
			if jsonName == "" && len(val) > 0 {
				jsonName = strings.Join(val, ",")
			}
			continue
		default:
		}
		name := key[11:]
		value := strings.Join(val, ",")
		tags[name] = value
		if name == "json" {
			// Full override: do not attempt to merge with json:name.
			jsonName = ""
		}
	}
	if _, ok := tags["json"]; !ok && jsonName != "" {
		if parent != nil && fieldName != "" && !parent.IsRequired(fieldName) {
			jsonName += ",omitempty"
		}
		tags["json"] = jsonName
	}
	if len(tags) == 0 {
		return ""
	}
	names := make([]string, 0, len(tags))
	for n := range tags {
		names = append(names, n)
	}
	sort.Strings(names)
	elems := make([]string, 0, len(names))
	for _, n := range names {
		elems = append(elems, fmt.Sprintf("%s:\"%s\"", n, tags[n]))
	}
	return " `" + strings.Join(elems, " ") + "`"
}
