package codegen

import (
	"fmt"

	"goa.design/goa/v3/expr"
)

type (
	// Attributor defines the behavior of an attribute expression during code
	// generation.
	Attributor interface {
		Scoper
		// Name generates a valid name for the given attribute type. ptr and
		// useDefault are used to generate inline struct type definitions.
		Name(att *expr.AttributeExpr, pkg string, ptr, useDefault bool) string
		// Ref generates a valid reference to the given attribute type.
		Ref(att *expr.AttributeExpr, pkg string) string
		// Field generates a valid data structure field identifier for the given
		// attribute and field name. If firstUpper is true then the field name
		// first letter is capitalized.
		Field(att *expr.AttributeExpr, name string, firstUpper bool) string
	}

	// AttributeContext contains properties which impacts the code generating
	// behavior of an attribute.
	AttributeContext struct {
		// Pointer if true indicates that the attribute uses pointers to hold
		// primitive types even if they are required or has a default value.
		// It ignores UseDefault and IgnoreRequired properties.
		Pointer bool
		// IgnoreRequired if true indicates that the transport object
		// (proto) uses non-pointers to hold required attributes and
		// therefore do not need to be validated.
		IgnoreRequired bool
		// UseDefault if true indicates that the attribute uses non-pointers for
		// primitive types if they have default value. If false, the attribute with
		// primitive types are non-pointers if they are required, otherwise they
		// are pointers.
		UseDefault bool
		// Scope is the attribute scope.
		Scope Attributor
		// DefaultPkg is the default package name where the attribute
		// type is found. it can be overridden via struct:pkg:path meta.
		DefaultPkg string
		// IsInterface is true if the attribute is an interface (union type).
		// In this case assigning child attributes requires a type assertion.
		IsInterface bool
	}

	// AttributeScope contains the scope of an attribute. It implements the
	// Attributor interface.
	AttributeScope struct {
		// scope is the name scope for the attribute.
		scope *NameScope
	}

	// TransformAttrs are the attributes that help in the transformation.
	TransformAttrs struct {
		// SourceCtx and TargetCtx are the source and target attribute context.
		SourceCtx, TargetCtx *AttributeContext
		// Prefix is the transform function helper prefix.
		Prefix string
	}

	// TransformFunctionData describes a helper function used to transform
	// user types. These are necessary to prevent potential infinite
	// recursion when a type attribute is defined recursively. For example:
	//
	//     var Recursive = Type("Recursive", func() {
	//         Attribute("r", "Recursive")
	//     }
	//
	// Transforming this type requires generating an intermediary function:
	//
	//     func recursiveToRecursive(r *Recursive) *service.Recursive {
	//         var t service.Recursive
	//         if r.R != nil {
	//             t.R = recursiveToRecursive(r.R)
	//         }
	//    }
	//
	TransformFunctionData struct {
		Name          string
		ParamTypeRef  string
		ResultTypeRef string
		Code          string
	}
)

// NewAttributeContext initializes an attribute context.
func NewAttributeContext(pointer, reqIgnore, useDefault bool, pkg string, scope *NameScope) *AttributeContext {
	return &AttributeContext{
		Pointer:        pointer,
		IgnoreRequired: reqIgnore,
		UseDefault:     useDefault,
		Scope:          NewAttributeScope(scope),
		DefaultPkg:     pkg,
	}
}

// NewAttributeScope initializes an attribute scope.
func NewAttributeScope(scope *NameScope) *AttributeScope {
	return &AttributeScope{scope: scope}
}

// IsCompatible returns an error if a and b are not both objects, both arrays,
// both maps, both unions or one union and one object.  actx and bctx are used
// to build the error message if any.
func IsCompatible(a, b expr.DataType, actx, bctx string) error {
	switch {
	case expr.IsObject(a):
		if !expr.IsObject(b) && !expr.IsUnion(b) {
			return fmt.Errorf("%s is an object but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsArray(a):
		if !expr.IsArray(b) {
			return fmt.Errorf("%s is an array but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsMap(a):
		if !expr.IsMap(b) {
			return fmt.Errorf("%s is a hash but %s type is %s", actx, bctx, b.Name())
		}
	case expr.IsUnion(a):
		if !expr.IsUnion(b) && !expr.IsObject(b) {
			return fmt.Errorf("%s is a union but %s type is %s", actx, bctx, b.Name())
		}
	default:
		aUT, isAUT := a.(expr.UserType)
		bUT, isBUT := b.(expr.UserType)
		switch {
		case isAUT && isBUT:
			return IsCompatible(aUT.Attribute().Type, bUT.Attribute().Type, actx, bctx)
		case isAUT:
			return IsCompatible(aUT.Attribute().Type, b, actx, bctx)
		case isBUT:
			return IsCompatible(a, bUT.Attribute().Type, actx, bctx)
		case a.Kind() != b.Kind():
			return fmt.Errorf("%s is a %s but %s type is %s", actx, a.Name(), bctx, b.Name())
		}
	}
	return nil
}

// AppendHelpers takes care of only appending helper functions from newH that
// are not already in oldH.
func AppendHelpers(oldH, newH []*TransformFunctionData) []*TransformFunctionData {
	for _, h := range newH {
		found := false
		for _, h2 := range oldH {
			if h.Name == h2.Name {
				found = true
				break
			}
		}
		if !found {
			oldH = append(oldH, h)
		}
	}
	return oldH
}

// MapDepth returns the level of nested maps. For unnested maps, it returns 0.
func MapDepth(m *expr.Map) int {
	return mapDepth(m.ElemType.Type, 0)
}

func mapDepth(dt expr.DataType, depth int, seen ...map[string]struct{}) int {
	if mp := expr.AsMap(dt); mp != nil {
		depth++
		depth = mapDepth(mp.ElemType.Type, depth, seen...)
	} else if ar := expr.AsArray(dt); ar != nil {
		depth = mapDepth(ar.ElemType.Type, depth, seen...)
	} else if mo := expr.AsObject(dt); mo != nil {
		var s map[string]struct{}
		if len(seen) > 0 {
			s = seen[0]
		} else {
			s = make(map[string]struct{})
			seen = append(seen, s)
		}
		key := dt.Name()
		if u, ok := dt.(expr.UserType); ok {
			key = u.ID()
		}
		if _, ok := s[key]; ok {
			return depth
		}
		s[key] = struct{}{}
		var level int
		for _, nat := range *mo {
			// if object type has attributes of type map then find out the attribute that has
			// the deepest level of nested maps
			lvl := 0
			lvl = mapDepth(nat.Attribute.Type, lvl, seen...)
			if lvl > level {
				level = lvl
			}
		}
		depth += level
	}
	return depth
}

// IsPrimitivePointer returns true if the attribute with the given name is a
// primitive pointer in the given parent attribute.
func (a *AttributeContext) IsPrimitivePointer(name string, att *expr.AttributeExpr) bool {
	if at := att.Find(name); at != nil && (at.Type == expr.Any || at.Type == expr.Bytes) {
		return false
	}
	if a.Pointer {
		return true
	}
	return att.IsPrimitivePointer(name, a.UseDefault)
}

// Pkg returns the package name of the given type.
func (a *AttributeContext) Pkg(att *expr.AttributeExpr) string {
	if loc := UserTypeLocation(att.Type); loc != nil {
		return loc.PackageName()
	}
	return a.DefaultPkg
}

// Dup creates a shallow copy of the AttributeContext.
func (a *AttributeContext) Dup() *AttributeContext {
	return &AttributeContext{
		Pointer:        a.Pointer,
		IgnoreRequired: a.IgnoreRequired,
		UseDefault:     a.UseDefault,
		Scope:          a.Scope,
		DefaultPkg:     a.DefaultPkg,
	}
}

// Name returns the type name for the given attribute.
func (a *AttributeScope) Name(att *expr.AttributeExpr, pkg string, ptr, useDefault bool) string {
	if _, ok := att.Type.(expr.UserType); !ok && expr.IsObject(att.Type) {
		// In the special case of anonymous / inline struct types the "name" is
		// in fact the struct typedef. In this case we need to force the
		// generation of the fields as pointers if needed as the default
		// GoTransform algorithm does not allow for an override.
		return a.scope.GoTypeDef(att, ptr, useDefault)
	}
	return a.scope.GoFullTypeName(att, pkg)
}

// Ref returns the type name for the given attribute.
func (a *AttributeScope) Ref(att *expr.AttributeExpr, pkg string) string {
	return a.scope.GoFullTypeRef(att, pkg)
}

// Field returns a valid Go struct field name.
func (a *AttributeScope) Field(att *expr.AttributeExpr, name string, firstUpper bool) string {
	return GoifyAtt(att, name, firstUpper)
}

// Scope returns the name scope.
func (a *AttributeScope) Scope() *NameScope {
	return a.scope
}
