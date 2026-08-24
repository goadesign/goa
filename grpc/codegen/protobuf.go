// This file defines protobuf-specific attribute naming used by gRPC type,
// validation, and transformation generation.
package codegen

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"

	"goa.design/goa/v3/codegen"
)

type (
	// protoBufScope supplies the Go names used for protobuf fields and types in
	// one generated service package.
	protoBufScope struct {
		service *ServiceData
		pkg     string
	}
)

const (
	// wrappedField is the name of the single field of the wrapper messages
	// synthesized by wrapAttr for non-object protobuf message types
	// (primitives, arrays and maps).
	wrappedField = "field"
	// wrappedAttrMeta is the meta key set by wrapAttr on the attribute of
	// synthesized wrapper message user types. Its presence is the contract
	// used by isWrappedAttr and unwrapAttr to identify wrapper messages
	// deterministically.
	wrappedAttrMeta = "grpc:wrapped"
)

// Name returns the protocol buffer type name.
func (p *protoBufScope) Name(att *expr.AttributeExpr, pkg string, _, _ bool) string {
	return protoBufGoFullTypeName(att, pkg, p.service)
}

// Ref returns the protocol buffer type reference.
func (p *protoBufScope) Ref(att *expr.AttributeExpr, pkg string) string {
	return protoBufGoFullTypeRef(att, pkg, p.service)
}

// Package returns the protocol buffer package qualifier for att.
func (p *protoBufScope) Package(*expr.AttributeExpr) string {
	return p.pkg
}

// Enter keeps nested protobuf messages in the same generated package.
func (p *protoBufScope) Enter(*expr.AttributeExpr) codegen.Attributor {
	return p
}

// IsSumType reports that protobuf unions use generated oneof messages rather
// than Goa service values that hold one selected branch.
func (*protoBufScope) IsSumType() bool {
	return false
}

// ValidatorCall returns a call using the first choice for a protobuf validation
// function name. The service may add a suffix when another function uses it.
func (p *protoBufScope) ValidatorCall(att *expr.AttributeExpr, view, target, _ string) string {
	name := "Validate" + p.Name(att, "", false, true) + codegen.Goify(view, true)
	return fmt.Sprintf("%s(%s)", name, target)
}

// Field returns the exact Go field name produced by the supported protobuf
// tools.
func (p *protoBufScope) Field(att *expr.AttributeExpr, name string, firstUpper bool) string {
	planned, ok := p.service.protobuf.plan.fieldName(att)
	if !ok {
		panic(fmt.Sprintf("protobuf field %q was not planned", name))
	}
	return planned
}

// OneofWrapper returns the generated wrapper type for one branch in one
// parent message.
func (p *protoBufScope) OneofWrapper(attribute *expr.AttributeExpr) string {
	name, ok := p.service.protobuf.plan.wrapperName(attribute)
	if !ok {
		panic("protobuf oneof branch was not planned")
	}
	if p.pkg == "" {
		return name
	}
	return p.pkg + "." + name
}

// Scope returns the object that assigns unique Go names in this package.
func (p *protoBufScope) Scope() *codegen.NameScope {
	return p.service.Scope
}

// protoBufTypeContext describes Go fields generated from protobuf messages.
// Singular booleans, numbers, strings, enums, and their aliases use pointers
// so validation can tell an omitted field from an explicit zero value. Bytes
// remain slices. Message fields, including Any, already use pointers.
func protoBufTypeContext(pkg string, service *ServiceData) *codegen.AttributeContext {
	ctx := codegen.NewAttributeContext(true, false, false, pkg, service.Scope)
	ctx.Scope = &protoBufScope{service: service, pkg: pkg}
	return ctx
}

// makeProtoBufMessage ensures the resulting attribute is an object user type so
// that it can be directly mapped to a protobuf type (protobuf messages must
// always be objects). If the given attribute type is a primitive, array, or a
// map, it wraps the given attribute with an object with a single "field"
// attribute. For nested arrays/maps, the inner array/map is wrapped into a
// user type.
func makeProtoBufMessage(att *expr.AttributeExpr, tname string, owner expr.ExampleIdentity) *expr.AttributeExpr {
	att = expr.DupAtt(att)
	expr.RemovePkgPath(att)
	ut, isut := att.Type.(expr.UserType)
	switch {
	case att.Type == expr.Empty:
		att.Type = expr.NewGeneratedUserType(tname, &expr.AttributeExpr{Type: &expr.Object{}}, owner)
		return att
	case expr.IsPrimitive(att.Type):
		wrapAttr(att, tname, true, owner)
		return att
	case isut:
		if expr.IsArray(ut) || expr.IsMap(ut) {
			wrapAttr(att, tname, false, owner)
		}
	case expr.IsArray(att.Type) || expr.IsMap(att.Type):
		wrapAttr(att, tname, false, owner)
	case expr.IsObject(att.Type) || expr.IsUnion(att.Type):
		att.Type = expr.NewGeneratedUserType(tname, expr.DupAtt(att), owner)
	}
	n := ""
	makeProtoBufMessageR(att, &n, owner, make(map[expr.UserType]struct{}))
	return att
}

// makeProtoBufMessageR is the recursive implementation of makeProtoBufMessage.
func makeProtoBufMessageR(att *expr.AttributeExpr, tname *string, owner expr.ExampleIdentity, seen map[expr.UserType]struct{}) {
	ut, isut := att.Type.(expr.UserType)

	// handle infinite recursions
	if isut {
		origin := ut.Origin()
		if _, ok := seen[origin]; ok {
			return
		}
		seen[origin] = struct{}{}
	}

	wrap := func(att *expr.AttributeExpr, tname string) {
		switch {
		case expr.IsArray(att.Type):
			wrapAttr(att, "ArrayOf"+tname+
				codegen.ProtobufName(protoBufShapeTypeName(expr.AsArray(att.Type).ElemType)), true, owner)
		case expr.IsMap(att.Type):
			m := expr.AsMap(att.Type)
			wrapAttr(att, tname+"MapOf"+
				codegen.ProtobufName(protoBufShapeTypeName(m.KeyType))+
				codegen.ProtobufName(protoBufShapeTypeName(m.ElemType)), true, owner)
		}
	}

	switch {
	case expr.IsPrimitive(att.Type):
		return
	case isut:
		switch {
		case expr.IsArray(ut):
			wrapAttr(ut.Attribute(), ut.Name(), false, expr.GRPCArrayWrapperExampleIdentity(ut))
		case expr.IsMap(ut):
			wrapAttr(ut.Attribute(), ut.Name(), false, expr.GRPCMapWrapperExampleIdentity(ut))
		}
		makeProtoBufMessageR(ut.Attribute(), tname, owner, seen)
	case expr.IsArray(att.Type):
		ar := expr.AsArray(att.Type)
		elementOwner := owner.ArrayElement(0)
		makeProtoBufMessageR(ar.ElemType, tname, elementOwner, seen)
		wrap(ar.ElemType, *tname)
	case expr.IsMap(att.Type):
		m := expr.AsMap(att.Type)
		valueOwner := owner.MapValue(0)
		makeProtoBufMessageR(m.ElemType, tname, valueOwner, seen)
		wrap(m.ElemType, *tname)
	case expr.IsUnion(att.Type):
		for _, nat := range expr.AsUnion(att.Type).Values {
			makeProtoBufMessageR(nat.Attribute, tname, owner.UnionMember(nat.Name), seen)
		}
	case expr.IsObject(att.Type):
		for _, nat := range *(expr.AsObject(att.Type)) {
			makeProtoBufMessageR(nat.Attribute, tname, owner.Member(nat.Name), seen)
		}
	}
}

// wrapAttr makes the attribute type a user type by wrapping the given
// attribute into an attribute named "field".
func wrapAttr(att *expr.AttributeExpr, tname string, req bool, owner expr.ExampleIdentity) {
	wrap := func(attr *expr.AttributeExpr) *expr.AttributeExpr {
		res := &expr.AttributeExpr{
			Type: &expr.Object{
				&expr.NamedAttributeExpr{
					Name: wrappedField,
					Attribute: &expr.AttributeExpr{
						Type:       attr.Type,
						Meta:       expr.MetaExpr{"rpc:tag": []string{"1"}},
						Validation: attr.Validation,
					},
				},
			},
			Meta: expr.MetaExpr{wrappedAttrMeta: []string{wrappedField}},
		}
		if req && !expr.IsArray(attr.Type) && !expr.IsMap(attr.Type) {
			res.Validation = &expr.ValidationExpr{
				Required: []string{wrappedField},
			}
		}
		return res
	}
	switch dt := att.Type.(type) {
	case expr.UserType:
		// Don't change the original user type. Create a copy and wrap that.
		att.Type = expr.NewGeneratedUserType(dt.Name(), wrap(expr.DupAtt(dt.Attribute())), owner)
	default:
		att.Type = expr.NewGeneratedUserType(tname, wrap(att), owner)
	}
	// Validation is moved to wrapped attribute.
	att.Validation = nil
}

// isWrappedAttr reports whether att references a wrapper message synthesized
// by wrapAttr: att.Type is a user type whose attribute carries the
// wrappedAttrMeta marker. User type chains are followed because wrapAttr
// nests the wrapper user type inside user defined array types (see
// makeProtoBufMessageR).
func isWrappedAttr(att *expr.AttributeExpr) bool {
	ut, ok := att.Type.(expr.UserType)
	if !ok {
		return false
	}
	wrapper := ut.Attribute()
	if len(wrapper.Meta[wrappedAttrMeta]) > 0 {
		return true
	}
	return isWrappedAttr(wrapper)
}

// unwrapAttr returns the attribute wrapped by the wrapper message referenced
// by att as synthesized by wrapAttr. att must either carry the
// wrappedAttrMeta marker directly (it is the wrapper user type attribute
// itself) or reference a (possibly nested) wrapper user type. unwrapAttr
// panics when att is not a wrapper or when the wrapper field is missing:
// both denote a violation of the wrapping contract established at message
// creation.
func unwrapAttr(att *expr.AttributeExpr) *expr.AttributeExpr {
	wrapper := att
	for len(wrapper.Meta[wrappedAttrMeta]) == 0 {
		ut, ok := wrapper.Type.(expr.UserType)
		if !ok {
			panic(fmt.Sprintf("attribute of type %q is not a protobuf wrapper message", att.Type.Name())) // bug
		}
		wrapper = ut.Attribute()
	}
	field := expr.AsObject(wrapper.Type).Attribute(wrappedField)
	if field == nil {
		panic(fmt.Sprintf("protobuf wrapper message of type %q has no %q attribute", att.Type.Name(), wrappedField)) // bug
	}
	return field
}

// protoBufMessageName returns the protocol buffer message name of the given
// attribute type.
func protoBufMessageName(att *expr.AttributeExpr, service *ServiceData) string {
	return protoBufFullMessageName(att, "", service)
}

// protoBufSourceMessageName returns the message name written to the .proto
// file.
func protoBufSourceMessageName(att *expr.AttributeExpr, service *ServiceData) string {
	userType, ok := att.Type.(expr.UserType)
	if !ok {
		if composite, ok := att.Type.(expr.CompositeExpr); ok {
			return protoBufSourceMessageName(composite.Attribute(), service)
		}
		panic(fmt.Sprintf("data type is not a protobuf message: received type %T", att.Type)) // bug
	}
	record := service.protobuf.message(att)
	if record == nil {
		panic(fmt.Sprintf("protobuf message %q has no planned name", userType.Name()))
	}
	return record.protoName
}

// protoBufFullMessageName returns the protocol buffer message name of the
// given user type qualified with the given package name if applicable.
func protoBufFullMessageName(att *expr.AttributeExpr, pkg string, service *ServiceData) string {
	switch actual := att.Type.(type) {
	case expr.UserType:
		if service.protobuf == nil {
			panic(fmt.Sprintf("protobuf message %q has no package catalog", actual.Name()))
		}
		record := service.protobuf.message(att)
		if record == nil {
			panic(fmt.Sprintf("protobuf message %q has no frozen declaration", actual.Name()))
		}
		n := record.name
		if pkg == "" {
			return n
		}
		return pkg + "." + n
	case expr.CompositeExpr:
		return protoBufFullMessageName(actual.Attribute(), pkg, service)
	default:
		panic(fmt.Sprintf("data type is not a protobuf message: received type %T", actual)) // bug
	}
}

// protoBufGoFullTypeName returns the protocol buffer type name qualified with
// the given package name for the given attribute generated after compiling
// the proto file (in *.pb.go).
func protoBufGoFullTypeName(att *expr.AttributeExpr, pkg string, service *ServiceData) string {
	if proto := att.Meta["struct:field:proto"]; len(proto) > 2 {
		typ := proto[2]
		if len(att.Meta["struct:field:proto"]) > 3 {
			elems := strings.Split(att.Meta["struct:field:proto"][3], "/")
			typ = elems[len(elems)-1] + "." + typ
		}
		return typ
	}
	if primitive := getPrimitive(att); primitive != nil {
		return protoBufGoFullTypeName(primitive, pkg, service)
	}
	switch actual := att.Type.(type) {
	case *expr.Union:
		if service.protobuf == nil {
			panic(fmt.Sprintf("protobuf oneof %q has no package catalog", actual.Name()))
		}
		name := service.protobuf.unionName(att)
		if pkg == "" {
			return name
		}
		return pkg + "." + name
	case expr.UserType, expr.CompositeExpr:
		return protoBufFullMessageName(att, pkg, service)
	case expr.Primitive:
		return protoBufNativeGoTypeName(att.Type)
	case *expr.Array:
		return "[]" + protoBufGoFullTypeRef(actual.ElemType, pkg, service)
	case *expr.Map:
		return fmt.Sprintf("map[%s]%s",
			protoBufGoFullTypeRef(actual.KeyType, pkg, service),
			protoBufGoFullTypeRef(actual.ElemType, pkg, service))
	case *expr.Object:
		return service.Scope.GoTypeDef(att, false, false)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// protoBufShapeTypeName returns the type name used when Goa wraps an array or
// map inside a protobuf message. It reads the design but does not reserve a Go
// name for the generated message.
func protoBufShapeTypeName(att *expr.AttributeExpr) string {
	if protos := att.Meta["struct:field:proto"]; len(protos) > 0 {
		return protos[0]
	}
	switch actual := att.Type.(type) {
	case expr.Primitive:
		return protoNativeType(actual)
	case expr.UserType:
		if names := att.Meta["struct:name:proto"]; len(names) > 0 {
			return names[0]
		}
		if names := actual.Attribute().Meta["struct:name:proto"]; len(names) > 0 {
			return names[0]
		}
		return codegen.ProtobufName(actual.Name())
	case expr.CompositeExpr:
		return protoBufShapeTypeName(actual.Attribute())
	case *expr.Object:
		return "Object"
	case *expr.Union:
		if actual.TypeName != "" {
			return codegen.ProtobufName(actual.TypeName)
		}
		return "Union"
	default:
		panic(fmt.Sprintf("unknown protobuf shaping type %T", actual)) // bug
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
		oneofName := codegen.ProtobufFieldName(actual.Name())
		if sd.protobuf != nil {
			oneofName = sd.protobuf.plan.sourceOneofName(att)
		}
		var fieldNames []string
		for _, nat := range actual.Values {
			fn := protobufSourceFieldName(nat.Name)
			if sd.protobuf != nil {
				fn = sd.protobuf.plan.sourceFieldName(nat.Attribute)
			}
			fieldNames = append(fieldNames, fn)
		}
		if sd.protobuf == nil {
			for slices.Contains(fieldNames, oneofName) {
				oneofName += "_oneof"
			}
		}
		def := "\toneof " + oneofName + " {"
		for i, nat := range actual.Values {
			fn := fieldNames[i]
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
			opt := protoJSONOption(nat.Attribute)
			def += fmt.Sprintf("\n\t\t%s%s %s = %d%s;", desc, typ, fn, fnum, opt)
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
		return protoBufSourceMessageName(att, sd)
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
				opt  string
				desc string
			)
			{
				fn = protobufSourceFieldName(nat.Name)
				if sd.protobuf != nil {
					fn = sd.protobuf.plan.sourceFieldName(nat.Attribute)
				}
				fnum = rpcTag(nat.Attribute)
				if prim := getPrimitive(nat.Attribute); prim != nil {
					typ = protoType(prim, sd)
				} else {
					typ = protoType(nat.Attribute, sd)
				}
				if expr.IsPrimitive(nat.Attribute.Type) && unAlias(nat.Attribute).Type.Kind() != expr.AnyKind {
					opt = "optional "
				}
				if nat.Attribute.Description != "" {
					desc = codegen.Comment(nat.Attribute.Description) + "\n\t"
				}
			}
			optJSON := protoJSONOption(nat.Attribute)
			ss = append(ss, fmt.Sprintf("\t%s%s%s %s = %d%s;", desc, opt, typ, fn, fnum, optJSON))
		}
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

func protoJSONOption(att *expr.AttributeExpr) string {
	if att == nil || att.Meta == nil {
		return ""
	}
	if names := att.Meta["proto:tag:json"]; len(names) > 0 && names[0] != "" {
		return fmt.Sprintf(" [json_name = %q]", names[0])
	}
	return ""
}

// protoBufGoFullTypeRef returns the Go code qualified with package name that
// refers to the Go type generated by compiling the protocol buffer
// (in *.pb.go) for the given attribute.
func protoBufGoFullTypeRef(att *expr.AttributeExpr, pkg string, service *ServiceData) string {
	name := protoBufGoFullTypeName(att, pkg, service)
	if expr.IsObject(att.Type) || expr.IsUnion(att.Type) {
		return "*" + name
	}
	return name
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
	case expr.AnyKind:
		return "google.protobuf.Value"
	default:
		panic(fmt.Sprintf("cannot compute native protocol buffer type for %T", t)) // bug
	}
}

// protoBufNativeGoTypeName returns the Go type corresponding to the given
// primitive type generated by the protocol buffer compiler after compiling
// the ".proto" file (in *.pb.go).
func protoBufNativeGoTypeName(t expr.DataType) string {
	switch t.Kind() {
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
	case expr.AnyKind:
		return "*structpb.Value"
	default:
		panic(fmt.Sprintf("cannot compute native protocol buffer type for %T %v", t, t)) // bug
	}
}

// rpcTag returns the unique numbered RPC tag from the given attribute. Every
// gRPC message field carries a tag by the time codegen runs: DSL validation
// rejects untagged fields and synthesized wrapper fields are tagged at
// creation, so a missing or unparseable tag is a bug.
func rpcTag(a *expr.AttributeExpr) uint64 {
	t, ok := a.FieldTag()
	if !ok {
		panic(fmt.Sprintf("attribute of type %q has no rpc:tag meta", a.Type.Name())) // bug
	}
	tag, err := strconv.ParseUint(t, 10, 64)
	if err != nil {
		panic(err) // bug
	}
	return tag
}
