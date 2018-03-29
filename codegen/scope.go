package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"goa.design/goa/design"
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
)

// NewNameScope creates an empty name scope.
func NewNameScope() *NameScope {
	ns := &NameScope{
		names:  make(map[string]string),
		counts: make(map[string]int),
	}
	if design.Root.API != nil {
		ns.Unique(design.Root.API, design.Root.API.Name)
	}
	return ns
}

// Unique builds the unique name for key using name and - if not unique -
// appending suffix and - if still not unique - a counter value.
func (s *NameScope) Unique(key Hasher, name string, suffix ...string) string {
	if n, ok := s.names[key.Hash()]; ok {
		return n
	}
	var (
		i   int
		suf string
	)
	_, ok := s.counts[name]
	if !ok {
		goto done
	}
	if len(suffix) > 0 {
		suf = suffix[0]
	}
	name += suf
	i, ok = s.counts[name]
	if !ok {
		goto done
	}
	name += strconv.Itoa(i + 1)
done:
	s.counts[name] = i + 1
	s.names[key.Hash()] = name
	return name
}

// GoTypeDef returns the Go code that defines a Go type which matches the data
// structure definition (the part that comes after `type foo`).
func (s *NameScope) GoTypeDef(att *design.AttributeExpr, useDefault bool) string {
	switch actual := att.Type.(type) {
	case design.Primitive:
		return GoNativeTypeName(actual)
	case *design.Array:
		d := s.GoTypeDef(actual.ElemType, useDefault)
		if design.IsObject(actual.ElemType.Type) {
			d = "*" + d
		}
		return "[]" + d
	case *design.Map:
		keyDef := s.GoTypeDef(actual.KeyType, useDefault)
		if design.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := s.GoTypeDef(actual.ElemType, useDefault)
		if design.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case *design.Object:
		var ss []string
		ss = append(ss, "struct {")
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
				tdef = s.GoTypeDef(at, useDefault)
				if design.IsObject(at.Type) || att.IsPrimitivePointer(name, useDefault) {
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
	case design.UserType:
		return s.GoTypeName(att)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// GoTypeRef returns the Go code that refers to the Go type which matches the
// given attribute type.
func (s *NameScope) GoTypeRef(att *design.AttributeExpr) string {
	name := s.GoTypeName(att)
	return goTypeRef(name, att.Type)
}

// GoFullTypeRef returns the Go code that refers to the Go type which matches
// the given attribute type defined in the given package if a user type.
func (s *NameScope) GoFullTypeRef(att *design.AttributeExpr, pkg string) string {
	name := s.GoFullTypeName(att, pkg)
	return goTypeRef(name, att.Type)
}

// GoTypeName returns the Go type name of the given attribute type.
func (s *NameScope) GoTypeName(att *design.AttributeExpr) string {
	return s.GoFullTypeName(att, "")
}

// GoFullTypeName returns the Go type name of the given data type qualified with
// the given package name if applicable and if not the empty string.
func (s *NameScope) GoFullTypeName(att *design.AttributeExpr, pkg string) string {
	switch actual := att.Type.(type) {
	case design.Primitive:
		return GoNativeTypeName(actual)
	case *design.Array:
		return "[]" + s.GoFullTypeRef(actual.ElemType, pkg)
	case *design.Map:
		return fmt.Sprintf("map[%s]%s",
			s.GoFullTypeRef(actual.KeyType, pkg),
			s.GoFullTypeRef(actual.ElemType, pkg))
	case *design.Object:
		return s.GoTypeDef(att, false)

	case design.UserType:
		n := s.Unique(actual, Goify(actual.Name(), true), "")
		if pkg == "" {
			return n
		}
		return pkg + "." + n
	case design.CompositeExpr:
		return s.GoFullTypeName(actual.Attribute(), pkg)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

func goTypeRef(name string, dt design.DataType) string {
	// For a raw struct, no need to dereference
	if _, ok := dt.(*design.Object); ok {
		return name
	}
	if design.IsObject(dt) {
		return "*" + name
	}
	return name
}
