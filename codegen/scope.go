package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"goa.design/goa.v2/design"
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
		Hash() string
	}
)

// NewNameScope creates an empty name scope.
func NewNameScope() *NameScope {
	return &NameScope{
		names:  make(map[string]string),
		counts: make(map[string]int),
	}
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
func (s *NameScope) GoTypeDef(att *design.AttributeExpr) string {
	switch actual := att.Type.(type) {
	case design.Primitive:
		return GoNativeTypeName(actual)
	case *design.Array:
		d := s.GoTypeDef(actual.ElemType)
		if design.IsObject(actual.ElemType.Type) {
			d = "*" + d
		}
		return "[]" + d
	case *design.Map:
		keyDef := s.GoTypeDef(actual.KeyType)
		if design.IsObject(actual.KeyType.Type) {
			keyDef = "*" + keyDef
		}
		elemDef := s.GoTypeDef(actual.ElemType)
		if design.IsObject(actual.ElemType.Type) {
			elemDef = "*" + elemDef
		}
		return fmt.Sprintf("map[%s]%s", keyDef, elemDef)
	case design.Object:
		var ss []string
		ss = append(ss, "struct {")
		WalkAttributes(actual, func(name string, at *design.AttributeExpr) error {
			var (
				fn   string
				tdef string
				desc string
				tags string
			)
			{
				fn = GoifyAtt(at, name, true)
				tdef = s.GoTypeDef(at)
				if design.IsObject(at.Type) || att.IsPrimitivePointer(name) {
					tdef = "*" + tdef
				}
				if at.Description != "" {
					desc = Comment(at.Description) + "\n\t"
				}
				tags = AttributeTags(att, at)
			}
			ss = append(ss, fmt.Sprintf("\t%s%s %s%s", desc, fn, tdef, tags))
			return nil
		})
		ss = append(ss, "}")
		return strings.Join(ss, "\n")
	case design.UserType:
		return s.GoTypeName(actual)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}

// GoTypeRef returns the Go code that refers to the Go type which matches the
// given data type.
func (s *NameScope) GoTypeRef(dt design.DataType) string {
	tname := s.GoTypeName(dt)
	if _, ok := dt.(design.Object); ok {
		return tname
	}
	if design.IsObject(dt) {
		return "*" + tname
	}
	return tname
}

// GoTypeName returns the Go type name of the given data type.
func (s *NameScope) GoTypeName(dt design.DataType) string {
	switch actual := dt.(type) {
	case design.Primitive:
		return GoNativeTypeName(dt)
	case *design.Array:
		return "[]" + s.GoTypeRef(actual.ElemType.Type)
	case *design.Map:
		return fmt.Sprintf("map[%s]%s", s.GoTypeRef(actual.KeyType.Type), s.GoTypeRef(actual.ElemType.Type))
	case design.Object:
		return "map[string]interface{}"
	case design.UserType:
		return s.Unique(dt, Goify(actual.Name(), true), "")
	case design.CompositeExpr:
		return s.GoTypeName(actual.Attribute().Type)
	default:
		panic(fmt.Sprintf("unknown data type %T", actual)) // bug
	}
}
