package codegen

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// NameScope defines a naming scope.
	NameScope struct {
		names  map[string]string // type hash to unique name
		counts map[string]int    // raw type name to occurrence count
	}

	// Hasher is the interface implemented by the objects that must be
	// scoped.
	Hasher interface {
		// Hash computes a unique instance hash suitable for indexing
		// in a map.
		Hash() string
	}

	// Scoper provides a scope for generating unique names.
	Scoper interface {
		Scope() *NameScope
	}
)

// NewNameScope creates an empty name scope.
func NewNameScope() *NameScope {
	return &NameScope{
		names:  make(map[string]string),
		counts: make(map[string]int),
	}
}

// HashedUnique builds the unique name for key using name and - if not unique -
// appending suffix and - if still not unique - a counter value. It returns
// the same value when called multiple times for a key returning the same hash.
func (s *NameScope) HashedUnique(key Hasher, name string, suffix ...string) string {
	if n, ok := s.names[key.Hash()]; ok {
		return n
	}
	name = s.Unique(name, suffix...)
	s.names[key.Hash()] = name
	return name
}

// Unique returns a unique name for the given name. A suffix is appended to the
// name if given name is not unique. If suffixed name is still not unique, a
// counter value is added to the suffixed name until unique.
func (s *NameScope) Unique(name string, suffix ...string) string {
	c, ok := s.counts[name]
	if !ok {
		s.counts[name]++
		return name
	}
	if len(suffix) > 0 {
		name += suffix[0]
		c, ok = s.counts[name]
		if !ok {
			s.counts[name]++
			return name
		}
	}
	for i := c; ; i++ {
		ret := name + strconv.Itoa(i+1)
		if _, ok := s.counts[ret]; !ok {
			s.counts[ret]++
			return ret
		}
	}
}

// Name returns a unique name for the given name by adding a counter value to
// the name until unique. It returns the same value when called multiple times
// for the same given name.
func (s *NameScope) Name(name string) string {
	i, ok := s.counts[name]
	if !ok {
		return name
	}
	name += strconv.Itoa(i + 1)
	return name
}

// GoTypeDef returns the Go code that defines a Go type which matches the data
// structure definition (the part that comes after `type foo`).
//
// ptr if true indicates that the attribute must be stored in a pointer
// (except array and map types which are always non-pointers)
//
// useDefault if true indicates that the attribute must not be a pointer
// if it has a default value.
func (s *NameScope) GoTypeDef(att *expr.AttributeExpr, ptr, useDefault bool) string {
	pkg := ""
	if loc := UserTypeLocation(att.Type); loc != nil {
		pkg = loc.PackageName()
	} else if p, ok := att.Meta.Last("struct:pkg:path"); ok && p != "" {
		pkg = Goify(filepath.Base(p), false)
	}
	return s.goTypeDef(att, ptr, useDefault, pkg)
}

// GoTypeDefWithTargetPkg returns the Go code that defines a Go type which matches the data
// structure definition, with UserType references qualified for the target package.
// This is used when generating inline struct definitions that may reference UserTypes
// from other packages.
//
// targetPkg is the package where the generated code will be placed.
func (s *NameScope) GoTypeDefWithTargetPkg(att *expr.AttributeExpr, ptr, useDefault bool, targetPkg string) string {
	// For inline structs, we need to ensure UserTypes are qualified appropriately
	// based on the target package where the code is being generated.
	return s.goTypeDefForTargetPkg(att, ptr, useDefault, targetPkg)
}

// goTypeDefForTargetPkg is like goTypeDef but uses the target package for qualifying UserTypes.
func (s *NameScope) goTypeDefForTargetPkg(att *expr.AttributeExpr, ptr, useDefault bool, targetPkg string) string {
	namer := func(a *expr.AttributeExpr) string {
		return s.GoFullTypeName(a, targetPkg)
	}
	return s.goTypeDefWithUserTypeNamer(att, ptr, useDefault, namer)
}

func (s *NameScope) goTypeDef(att *expr.AttributeExpr, ptr, useDefault bool, pkg string) string {
	namer := func(a *expr.AttributeExpr) string {
		// Qualify with declaring package only when different from the struct's package
		if loc := UserTypeLocation(a.Type); loc != nil && loc.PackageName() != pkg {
			return s.GoFullTypeName(a, loc.PackageName())
		}
		return s.GoFullTypeName(a, "")
	}
	return s.goTypeDefWithUserTypeNamer(att, ptr, useDefault, namer)
}

// goTypeDefWithUserTypeNamer is the shared implementation used by goTypeDef and
// goTypeDefForTargetPkg. The nameForUserType function controls how user types
// are qualified in the generated definition.
func (s *NameScope) goTypeDefWithUserTypeNamer(att *expr.AttributeExpr, ptr, useDefault bool, nameForUserType func(*expr.AttributeExpr) string) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		if t, _ := GetMetaType(att); t != "" {
			return t
		}
		return GoNativeTypeName(actual)
	case *expr.Array:
		d := s.goTypeDefWithUserTypeNamer(actual.ElemType, ptr, useDefault, nameForUserType)
		if expr.IsObject(actual.ElemType.Type) {
			d = "*" + d
		}
		return "[]" + d
	case *expr.Map:
		keyDef := s.goTypeDefWithUserTypeNamer(actual.KeyType, ptr, useDefault, nameForUserType)
		if expr.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := s.goTypeDefWithUserTypeNamer(actual.ElemType, ptr, useDefault, nameForUserType)
		if expr.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case *expr.Union:
		return fmt.Sprintf("interface{\n\t%s()\n}", UnionValTypeName(actual.TypeName))
	case *expr.Object:
		ss := []string{"struct {"}
		for _, nat := range *actual {
			var (
				fn   string
				tdef string
				desc string
				tags string

				name = nat.Name
				at   = nat.Attribute
			)
			{
				fn = GoifyAtt(at, name, true)
				tdef = s.goTypeDefWithUserTypeNamer(at, ptr, useDefault, nameForUserType)
				if expr.IsObject(at.Type) ||
					att.IsPrimitivePointer(name, useDefault) ||
					(ptr && expr.IsPrimitive(at.Type) && at.Type.Kind() != expr.AnyKind && at.Type.Kind() != expr.BytesKind) {
					tdef = "*" + tdef
				}
				if at.Description != "" {
					desc = Comment(at.Description) + "\n\t"
				}
				tags = AttributeTags(att, at)
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s%s", desc, fn, tdef, tags))
		}
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	case expr.UserType:
		if actual == expr.Empty {
			return "struct {}"
		}
		return nameForUserType(att)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// GoVar returns the Go code that returns the address of a variable of the Go type
// which matches the given attribute type.
func (*NameScope) GoVar(varName string, dt expr.DataType) string {
	// For a raw struct, no need to indirecting
	if isRawStruct(dt) {
		return varName
	}
	return "&" + varName
}

// GoTypeRef returns the Go code that refers to the Go type which matches the
// given attribute type.
func (s *NameScope) GoTypeRef(att *expr.AttributeExpr) string {
	name := s.GoTypeName(att)
	return goTypeRef(name, att.Type)
}

// GoTypeRefWithDefaults returns the Go code that refers to the Go type which
// matches the given attribute type. The result of this function differs from
// GoTypeRef when the attribute type is an object (note: not a user type) and
// the reference is thus an inline struct definition. In this case accounting
// for default values may cause child attributes to use non-pointer fields.
func (s *NameScope) GoTypeRefWithDefaults(att *expr.AttributeExpr) string {
	name := s.GoTypeNameWithDefaults(att)
	return goTypeRef(name, att.Type)
}

// GoFullTypeRef returns the Go code that refers to the Go type which matches
// the given attribute type defined in the given package if a user type.
func (s *NameScope) GoFullTypeRef(att *expr.AttributeExpr, pkg string) string {
	name := s.GoFullTypeName(att, pkg)
	return goTypeRef(name, att.Type)
}

// GoTypeName returns the Go type name of the given attribute type.
func (s *NameScope) GoTypeName(att *expr.AttributeExpr) string {
	return s.GoFullTypeName(att, "")
}

// GoTypeNameWithDefaults returns the Go type name of the given attribute type.
// The result of this function differs from GoTypeName when the attribute type
// is an object (note: not a user type) and the name is thus an inline struct
// definition. In this case accounting for default values may cause child
// attributes to use non-pointer fields.
func (s *NameScope) GoTypeNameWithDefaults(att *expr.AttributeExpr) string {
	if _, ok := att.Type.(*expr.Object); ok {
		return s.GoTypeDef(att, false, true)
	}
	return s.GoTypeName(att)
}

// GoFullTypeName returns the Go type name of the given data type qualified with
// the given package name if applicable and if not the empty string.
func (s *NameScope) GoFullTypeName(att *expr.AttributeExpr, pkg string) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		if t, _ := GetMetaType(att); t != "" {
			return t
		}
		return GoNativeTypeName(actual)
	case *expr.Array:
		return "[]" + s.GoFullTypeRef(actual.ElemType, pkgWithDefault(actual.ElemType.Type, pkg))
	case *expr.Map:
		return fmt.Sprintf("map[%s]%s",
			s.GoFullTypeRef(actual.KeyType, pkgWithDefault(actual.KeyType.Type, pkg)),
			s.GoFullTypeRef(actual.ElemType, pkgWithDefault(actual.ElemType.Type, pkg)))
	case *expr.Object:
		return s.GoTypeDef(att, false, false)
	case expr.UserType, *expr.Union:
		if actual == expr.ErrorResult {
			return "goa.ServiceError"
		}
		n := s.HashedUnique(actual, Goify(actual.Name(), true), "")
		if pkg == "" {
			return n
		}
		return pkg + "." + n
	case expr.CompositeExpr:
		return s.GoFullTypeName(actual.Attribute(), pkgWithDefault(actual.Attribute().Type, pkg))
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// pkgWithDefault returns the package defining the given type. If the types is a
// user type with "struct:pkg:path" metadata then it returns the corresponding
// value, otherwise it returns pkg.
func pkgWithDefault(dt expr.DataType, pkg string) string {
	if loc := UserTypeLocation(dt); loc != nil {
		return loc.PackageName()
	}
	return pkg
}

func goTypeRef(name string, dt expr.DataType) string {
	// For a raw struct, no need to dereference
	if isRawStruct(dt) {
		return name
	}
	return "*" + name
}

func isRawStruct(dt expr.DataType) bool {
	if _, ok := dt.(*expr.Object); ok {
		return true
	}
	if expr.IsObject(dt) {
		return false
	}
	if expr.IsUnion(dt) {
		return false
	}
	return true
}
