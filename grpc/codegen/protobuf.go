package codegen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"

	"goa.design/goa/v3/codegen"
)

type (
	// protoBufScope is the scope for protocol buffer attribute types.
	protoBufScope struct {
		scope *codegen.NameScope
	}
)

// Name returns the protocol buffer type name.
func (p *protoBufScope) Name(att *expr.AttributeExpr, pkg string, ptr, useDefault bool) string {
	return protoBufGoFullTypeName(att, pkg, p.scope)
}

// Ref returns the protocol buffer type reference.
func (p *protoBufScope) Ref(att *expr.AttributeExpr, pkg string) string {
	return protoBufGoFullTypeRef(att, pkg, p.scope)
}

// Field returns the field name as generated by protocol buffer compiler.
// NOTE: protoc does not care about common initialisms like api -> API so we
// first transform the name into snake case to end up with Api.
func (p *protoBufScope) Field(att *expr.AttributeExpr, name string, firstUpper bool) string {
	return protoBufifyAtt(att, codegen.SnakeCase(name), firstUpper)
}

// Scope returns the name scope.
func (p *protoBufScope) Scope() *codegen.NameScope {
	return p.scope
}

// protoBufTypeContext returns a contextual attribute for the protocol buffer type.
func protoBufTypeContext(pkg string, scope *codegen.NameScope) *codegen.AttributeContext {
	ctx := codegen.NewAttributeContext(false, true, true, pkg, scope)
	ctx.Scope = &protoBufScope{scope: scope}
	return ctx
}

// makeProtoBufMessage ensures the resulting attribute is an object user type so
// that it can be directly mapped to a protobuf type (protobuf messages must
// always be objects). If the given attribute type is a primitive, array, or a
// map, it wraps the given attribute with an object with a single "field"
// attribute. For nested arrays/maps, the inner array/map is wrapped into a
// user type.
func makeProtoBufMessage(att *expr.AttributeExpr, tname string, sd *ServiceData) *expr.AttributeExpr {
	att = expr.DupAtt(att)
	ut, isut := att.Type.(expr.UserType)
	switch {
	case att.Type == expr.Empty:
		att.Type = &expr.UserTypeExpr{
			TypeName:      tname,
			AttributeExpr: &expr.AttributeExpr{Type: &expr.Object{}},
			UID:           sd.Name + "#" + tname,
		}
		return att
	case expr.IsPrimitive(att.Type):
		wrapAttr(att, tname, sd)
		return att
	case isut:
		if expr.IsArray(ut) {
			wrapAttr(att, tname, sd)
		}
	case expr.IsArray(att.Type) || expr.IsMap(att.Type):
		wrapAttr(att, tname, sd)
	case expr.IsObject(att.Type) || expr.IsUnion(att.Type):
		att.Type = &expr.UserTypeExpr{
			TypeName:      tname,
			AttributeExpr: expr.DupAtt(att),
			UID:           sd.Name + "#" + tname,
		}
	}
	n := ""
	makeProtoBufMessageR(att, &n, sd, make(map[string]struct{}))
	return att
}

// makeProtoBufMessageR is the recursive implementation of makeProtoBufMessage.
func makeProtoBufMessageR(att *expr.AttributeExpr, tname *string, sd *ServiceData, seen map[string]struct{}) {
	ut, isut := att.Type.(expr.UserType)

	// handle infinite recursions
	if isut {
		if _, ok := seen[ut.ID()]; ok {
			return
		}
		seen[ut.ID()] = struct{}{}
	}

	wrap := func(att *expr.AttributeExpr, tname string) {
		switch {
		case expr.IsArray(att.Type):
			wrapAttr(att, "ArrayOf"+tname+
				protoBufify(protoBufMessageDef(expr.AsArray(att.Type).ElemType, sd), true, true), sd)
		case expr.IsMap(att.Type):
			m := expr.AsMap(att.Type)
			wrapAttr(att, tname+"MapOf"+
				protoBufify(protoBufMessageDef(m.KeyType, sd), true, true)+
				protoBufify(protoBufMessageDef(m.ElemType, sd), true, true), sd)
		}
	}

	switch {
	case expr.IsPrimitive(att.Type):
		return
	case isut:
		if expr.IsArray(ut) {
			wrapAttr(ut.Attribute(), ut.Name(), sd)
		}
		makeProtoBufMessageR(ut.Attribute(), tname, sd, seen)
	case expr.IsArray(att.Type):
		ar := expr.AsArray(att.Type)
		makeProtoBufMessageR(ar.ElemType, tname, sd, seen)
		wrap(ar.ElemType, *tname)
	case expr.IsMap(att.Type):
		m := expr.AsMap(att.Type)
		makeProtoBufMessageR(m.ElemType, tname, sd, seen)
		wrap(m.ElemType, *tname)
	case expr.IsUnion(att.Type):
		for _, nat := range expr.AsUnion(att.Type).Values {
			makeProtoBufMessageR(nat.Attribute, tname, sd, seen)
		}
	case expr.IsObject(att.Type):
		for _, nat := range *(expr.AsObject(att.Type)) {
			makeProtoBufMessageR(nat.Attribute, tname, sd, seen)
		}
	}
}

// wrapAttr makes the attribute type a user type by wrapping the given
// attribute into an attribute named "field".
func wrapAttr(att *expr.AttributeExpr, tname string, sd *ServiceData) {
	wrap := func(attr *expr.AttributeExpr) *expr.AttributeExpr {
		return &expr.AttributeExpr{
			Type: &expr.Object{
				&expr.NamedAttributeExpr{
					Name: "field",
					Attribute: &expr.AttributeExpr{
						Type:       attr.Type,
						Meta:       expr.MetaExpr{"rpc:tag": []string{"1"}},
						Validation: attr.Validation,
					},
				},
			},
		}
	}
	switch dt := att.Type.(type) {
	case expr.UserType:
		// Don't change the original user type. Create a copy and wrap that.
		ut := expr.Dup(dt).(expr.UserType)
		ut.SetAttribute(wrap(ut.Attribute()))
		att.Type = ut
	default:
		att.Type = &expr.UserTypeExpr{
			TypeName:      tname,
			AttributeExpr: wrap(att),
			UID:           sd.Name + "#" + tname,
		}
	}
	// Validation is moved to wrapped attribute.
	att.Validation = nil
}

// unwrapAttr returns the attribute under the attribute name "field".
// If "field" does not exist, it returns the given attribute.
func unwrapAttr(att *expr.AttributeExpr) *expr.AttributeExpr {
	if a := att.Find("field"); a != nil {
		return a
	}
	return att
}

// protoBufMessageName returns the protocol buffer message name of the given
// attribute type.
func protoBufMessageName(att *expr.AttributeExpr, s *codegen.NameScope) string {
	return protoBufFullMessageName(att, "", s)
}

// protoBufFullMessageName returns the protocol buffer message name of the
// given user type qualified with the given package name if applicable.
func protoBufFullMessageName(att *expr.AttributeExpr, pkg string, s *codegen.NameScope) string {
	switch actual := att.Type.(type) {
	case expr.UserType, *expr.Union:
		n := s.HashedUnique(actual, protoBufify(actual.Name(), true, true), "")
		if pkg == "" {
			return n
		}
		return pkg + "." + n
	case expr.CompositeExpr:
		return protoBufFullMessageName(actual.Attribute(), pkg, s)
	default:
		panic(fmt.Sprintf("data type is not a user type or union: received type %T", actual)) // bug
	}
}

// protoBufGoFullTypeName returns the protocol buffer type name for the given
// attribute generated after compiling the proto file (in *.pb.go).
func protoBufGoTypeName(att *expr.AttributeExpr, s *codegen.NameScope) string {
	return protoBufGoFullTypeName(att, "", s)
}

// protoBufGoFullTypeName returns the protocol buffer type name qualified with
// the given package name for the given attribute generated after compiling
// the proto file (in *.pb.go).
func protoBufGoFullTypeName(att *expr.AttributeExpr, pkg string, s *codegen.NameScope) string {
	if proto := att.Meta["struct:field:proto"]; len(proto) > 2 {
		typ := proto[2]
		if len(att.Meta["struct:field:proto"]) > 3 {
			elems := strings.Split(att.Meta["struct:field:proto"][3], "/")
			typ = elems[len(elems)-1] + "." + typ
		}
		return typ
	}
	switch actual := att.Type.(type) {
	case expr.UserType, expr.CompositeExpr, *expr.Union:
		return protoBufFullMessageName(att, pkg, s)
	case expr.Primitive:
		return protoBufNativeGoTypeName(att)
	case *expr.Array:
		return "[]" + protoBufGoFullTypeRef(actual.ElemType, pkg, s)
	case *expr.Map:
		return fmt.Sprintf("map[%s]%s",
			protoBufGoFullTypeRef(actual.KeyType, pkg, s),
			protoBufGoFullTypeRef(actual.ElemType, pkg, s))
	case *expr.Object:
		return s.GoTypeDef(att, false, false)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// protoType returns the protocol buffer type name for the given attribute.
func protoType(att *expr.AttributeExpr, sd *ServiceData) string {
	if protos := att.Meta["struct:field:proto"]; len(protos) > 0 {
		return protos[0]
	}
	return protoBufMessageDef(att, sd)
}

// protoBufMessageDef returns the protocol buffer code that defines a message
// which matches the data structure definition (the part that comes after
// `message foo`). The message is defined using the proto3 syntax.
func protoBufMessageDef(att *expr.AttributeExpr, sd *ServiceData) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		return protoNativeType(att.Type)
	case *expr.Array:
		return "repeated " + protoType(actual.ElemType, sd)
	case *expr.Map:
		return fmt.Sprintf("map<%s, %s>", protoType(actual.KeyType, sd), protoType(actual.ElemType, sd))
	case *expr.Union:
		def := "\toneof " + codegen.SnakeCase(protoBufify(actual.Name(), false, false)) + " {"
		for _, nat := range actual.Values {
			fn := codegen.SnakeCase(protoBufify(nat.Name, false, false))
			fnum := rpcTag(nat.Attribute)
			var typ string
			if prim := getPrimitive(nat.Attribute); prim != nil {
				typ = protoType(prim, sd)
			} else {
				typ = protoType(nat.Attribute, sd)
			}
			var desc string
			if d := nat.Attribute.Description; d != "" {
				desc = codegen.Comment(d) + "\n\t"
			}
			def += fmt.Sprintf("\n\t\t%s%s %s = %d;", desc, typ, fn, fnum)
		}
		def += "\n\t}"
		return def
	case expr.UserType:
		if actual == expr.Empty {
			return " {}"
		}
		if prim := getPrimitive(att); prim != nil {
			return protoBufMessageDef(prim, sd)
		}
		return protoBufMessageName(att, sd.Scope)
	case *expr.Object:
		var ss []string
		ss = append(ss, " {")
		for _, nat := range *actual {
			if expr.IsUnion(nat.Attribute.Type) {
				ss = append(ss, protoBufMessageDef(nat.Attribute, sd))
				continue
			}
			var (
				fn   string
				fnum uint64
				typ  string
				desc string
			)
			{
				fn = codegen.SnakeCase(protoBufify(nat.Name, false, false))
				fnum = rpcTag(nat.Attribute)
				if prim := getPrimitive(nat.Attribute); prim != nil {
					typ = protoType(prim, sd)
				} else {
					typ = protoType(nat.Attribute, sd)
				}
				if nat.Attribute.Description != "" {
					desc = codegen.Comment(nat.Attribute.Description) + "\n\t"
				}
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s = %d;", desc, typ, fn, fnum))
		}
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// protoBufGoFullTypeRef returns the Go code qualified with package name that
// refers to the Go type generated by compiling the protocol buffer
// (in *.pb.go) for the given attribute.
func protoBufGoFullTypeRef(att *expr.AttributeExpr, pkg string, s *codegen.NameScope) string {
	name := protoBufGoFullTypeName(att, pkg, s)
	if expr.IsObject(att.Type) || expr.IsUnion(att.Type) {
		return "*" + name
	}
	return name
}

var digits = regexp.MustCompile("[0-9]+")

// protoBufify makes a valid protocol buffer identifier out of any string.
// It does that by removing any non letter and non digit character and by
// making sure the first character is a letter or "_". protoBufify produces a
// "CamelCase" version of the string.
//
// If firstUpper is true the first character of the identifier is uppercase
// otherwise it's lowercase.
func protoBufify(str string, firstUpper, acronym bool) string {
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

	// The CamelCase implementation of protoc-gen-go considers digits as words
	// but our CamelCase implementation considers them as lower case characters,
	// compensate by adding an underscore after any series of digits.
	// See https://github.com/golang/protobuf/blob/d04d7b157bb510b1e0c10132224b616ac0e26b17/protoc-gen-go/generator/generator.go#L2648-L2685
	str = string(digits.ReplaceAllFunc([]byte(str), func(match []byte) []byte {
		res := make([]byte, len(match)+1) // need to allocate new slice
		copy(res, match)
		res[len(res)-1] = '_'
		return res
	}))

	str = codegen.CamelCase(str, firstUpper, acronym)
	if str == "" {
		// All characters are invalid. Produce a default value.
		if firstUpper {
			return "Val"
		}
		return "val"
	}

	return fixReservedProtoBuf(str)
}

// protoBufifyAtt honors any struct:field:name meta set on the attribute and
// and calls protoBufify with the tag value if present or the given name
// otherwise.
func protoBufifyAtt(att *expr.AttributeExpr, name string, upper bool) string {
	if tname, ok := att.Meta["struct:field:name"]; ok {
		if len(tname) > 0 {
			name = tname[0]
		}
	}
	return protoBufify(name, upper, false)
}

// protoNativeType returns the protocol buffer built-in type
// corresponding to the given primitive type. It panics if t is not a
// primitive type.
func protoNativeType(t expr.DataType) string {
	switch t.Kind() {
	case expr.BooleanKind:
		return "bool"
	case expr.IntKind:
		return "sint32"
	case expr.Int32Kind:
		return "sint32"
	case expr.Int64Kind:
		return "sint64"
	case expr.UIntKind:
		return "uint32"
	case expr.UInt32Kind:
		return "uint32"
	case expr.UInt64Kind:
		return "uint64"
	case expr.Float32Kind:
		return "float"
	case expr.Float64Kind:
		return "double"
	case expr.StringKind:
		return "string"
	case expr.BytesKind:
		return "bytes"
	default:
		panic(fmt.Sprintf("cannot compute native protocol buffer type for %T", t)) // bug
	}
}

// protoBufNativeGoTypeName returns the Go type corresponding to the given
// primitive type generated by the protocol buffer compiler after compiling
// the ".proto" file (in *.pb.go).
func protoBufNativeGoTypeName(att *expr.AttributeExpr) string {
	typeName, _ := codegen.GetMetaType(att)
	if typeName != "" {
		return typeName
	}
	switch att.Type.Kind() {
	case expr.BooleanKind:
		return "bool"
	case expr.IntKind:
		return "int32"
	case expr.Int32Kind:
		return "int32"
	case expr.Int64Kind:
		return "int64"
	case expr.UIntKind:
		return "uint32"
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
	default:
		panic(fmt.Sprintf("cannot compute native protocol buffer type for %T", att.Type)) // bug
	}
}

// rpcTag returns the unique numbered RPC tag from the given attribute.
func rpcTag(a *expr.AttributeExpr) uint64 {
	var tag uint64
	if t, ok := a.FieldTag(); ok {
		tn, err := strconv.ParseUint(t, 10, 64)
		if err != nil {
			panic(err) // bug (should catch invalid field numbers in validation)
		}
		tag = tn
	}
	return tag
}

// fixReservedProtoBuf appends an underscore on to protocol buffer reserved
// keywords.
func fixReservedProtoBuf(w string) string {
	if _, ok := reservedProtoBuf[codegen.CamelCase(w, false, false)]; ok {
		w += "_"
	}
	return w
}

var (
	// reserved protocol buffer keywords and package names
	reservedProtoBuf = map[string]struct{}{
		// types
		"bool":     {},
		"bytes":    {},
		"double":   {},
		"fixed32":  {},
		"fixed64":  {},
		"float":    {},
		"int32":    {},
		"int64":    {},
		"sfixed32": {},
		"sfixed64": {},
		"sint32":   {},
		"sint64":   {},
		"string":   {},
		"uint32":   {},
		"uint64":   {},

		// reserved
		"enum":     {},
		"import":   {},
		"map":      {},
		"message":  {},
		"oneof":    {},
		"option":   {},
		"package":  {},
		"public":   {},
		"repeated": {},
		"reserved": {},
		"returns":  {},
		"rpc":      {},
		"service":  {},
		"syntax":   {},
	}
)
