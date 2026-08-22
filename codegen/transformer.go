// This file defines the naming and pointer contracts shared by Go
// transformation and validation generators.
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
		// Package returns the qualifier used to reference att from the current
		// generated file, or the empty string for a same-package declaration.
		Package(att *expr.AttributeExpr) string
		// Enter returns the resolver that owns att and declarations nested in it.
		Enter(att *expr.AttributeExpr) Attributor
		// IsSumType reports whether unions use Goa's generated sum-type layout.
		IsSumType() bool
		// ValidatorName returns the package-level validation function for att and
		// the selected result-type view.
		ValidatorName(att *expr.AttributeExpr, view string) string
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
		// UnionPointer if true indicates that optional sum-type union fields use
		// pointers to preserve transport-level presence. Required union fields also
		// use pointers when Pointer is true. Service types leave this false because
		// the empty union discriminator represents omission after decoding.
		UnionPointer bool
	}

	// AttributeScope contains the scope of an attribute. It implements the
	// Attributor interface.
	AttributeScope struct {
		// scope is the name scope for the attribute.
		scope *NameScope
		// pkg is the default generated Go package qualifier.
		pkg string
	}

	// TransformAttrs are the attributes that help in the transformation.
	TransformAttrs struct {
		// SourceCtx and TargetCtx are the source and target attribute context.
		SourceCtx, TargetCtx *AttributeContext
		// Prefix is the transform function helper prefix.
		Prefix string
		// Hooks are optional generator specific extension points
		// consulted by the transform engine. Nil selects the engine
		// defaults.
		Hooks   *TransformHooks
		helpers map[TransformHelperID]TransformHelper
		calls   *transformCallCursor
	}

	// TransformHelperID identifies one recursive helper selected by a transform
	// plan. Its representation is deliberately private: generators may compare
	// IDs or use them as map keys but cannot reconstruct them from generated
	// names.
	TransformHelperID struct {
		plan  *TransformPlan
		index int
	}

	// TransformHelper describes one recursive source-to-target operation
	// retained by a transform plan. Render uses its ID and declaration for both
	// calls and definitions.
	TransformHelper struct {
		// ID is the opaque identity owned by the transform plan.
		ID TransformHelperID
		// Source is the exact source attribute selected during planning.
		Source *expr.AttributeExpr
		// Target is the exact target attribute selected during planning.
		Target *expr.AttributeExpr
		// Required reports whether nil is rejected by the helper operation.
		Required bool
		// Occurrence is the one-based position of this helper operation in the
		// transform plan's stable traversal.
		Occurrence int
		// Declaration is the canonical package-level function bound before render.
		// Render rejects an unbound helper.
		Declaration *NameDeclaration
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
		// ID is the retained helper identity used to render this definition.
		ID TransformHelperID
		// Declaration is the canonical package declaration for retained transforms.
		// It is nil for the separate one-pass transform API.
		Declaration *NameDeclaration
		// Name is the generated helper name for staged legacy callers. It is empty
		// when Declaration owns the final name.
		Name string
		// ParamTypeRef is the generated Go reference to the helper parameter type.
		ParamTypeRef string
		// ResultTypeRef is the generated Go reference to the helper result type.
		ResultTypeRef string
		// Code is the helper body.
		Code string
	}

	// TransformPlan retains the exact source-target operations and recursive
	// helpers selected for one Go transformation. Generators build the plan
	// before package names freeze and render it afterward with contexts that
	// resolve the final declarations.
	TransformPlan struct {
		source     *expr.AttributeExpr
		target     *expr.AttributeExpr
		sourceCtx  *AttributeContext
		targetCtx  *AttributeContext
		helpers    []TransformHelper
		operations []*transformOperation
	}

	// transformPair identifies one recursive source-target operation by its
	// expression declarations rather than a provisional helper spelling.
	transformPair struct {
		source expr.DataType
		target expr.DataType
	}

	// transformOperation retains the ordered helper calls made while rendering
	// the top-level transform or one helper body.
	transformOperation struct {
		calls []transformCall
	}

	// transformCall binds one ordered call edge to the helper that renders its
	// conversion.
	transformCall struct {
		helper TransformHelperID
	}

	// transformCallCursor tracks the retained calls consumed by one render.
	transformCallCursor struct {
		calls []transformCall
		next  int
	}
)

// NewAttributeContext initializes an attribute context.
func NewAttributeContext(pointer, reqIgnore, useDefault bool, pkg string, scope *NameScope) *AttributeContext {
	return &AttributeContext{
		Pointer:        pointer,
		IgnoreRequired: reqIgnore,
		UseDefault:     useDefault,
		Scope:          newAttributeScope(scope, pkg),
	}
}

// NewAttributeScope initializes an attribute scope.
func NewAttributeScope(scope *NameScope) *AttributeScope {
	return newAttributeScope(scope, "")
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
			if sameTransformHelper(h, h2) {
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

// sameTransformHelper compares canonical declarations for catalog-backed
// helpers and generated names for staged legacy helpers.
func sameTransformHelper(left, right *TransformFunctionData) bool {
	if left.Declaration != nil || right.Declaration != nil {
		return left.ID == right.ID
	}
	return left.Name == right.Name
}

// MapDepth returns the level of nested maps. For unnested maps, it returns 0.
func MapDepth(m *expr.Map) int {
	return mapDepth(m.ElemType.Type, 0)
}

func mapDepth(dt expr.DataType, depth int, seen ...map[expr.DataType]struct{}) int {
	if mp := expr.AsMap(dt); mp != nil {
		depth++
		depth = mapDepth(mp.ElemType.Type, depth, seen...)
	} else if ar := expr.AsArray(dt); ar != nil {
		depth = mapDepth(ar.ElemType.Type, depth, seen...)
	} else if mo := expr.AsObject(dt); mo != nil {
		var s map[expr.DataType]struct{}
		if len(seen) > 0 {
			s = seen[0]
		} else {
			s = make(map[expr.DataType]struct{})
			seen = append(seen, s)
		}
		key := dt
		if u, ok := dt.(expr.UserType); ok {
			key = u.Origin()
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

// IsFieldPointer reports whether the generated Go field is pointer-backed in
// this context.
func (a *AttributeContext) IsFieldPointer(name string, att *expr.AttributeExpr) bool {
	field := expr.AsObject(att.Type).Attribute(name)
	if expr.IsUnion(field.Type) {
		return a.IsUnionPointer(att.IsRequired(name))
	}
	if !a.Scope.IsSumType() {
		return expr.IsPrimitive(field.Type) && a.IsPrimitivePointer(name, att)
	}
	return goFieldIsPointer(att, name, a.Pointer, a.UseDefault)
}

// IsUnionPointer reports whether a sum-type union field with the given
// requiredness is pointer-backed in this context.
func (a *AttributeContext) IsUnionPointer(required bool) bool {
	return a.UnionPointer && (!required || a.Pointer)
}

// Pkg returns the package name of the given type.
func (a *AttributeContext) Pkg(att *expr.AttributeExpr) string {
	return a.Scope.Package(att)
}

// Enter returns a copy whose attributor owns att and unlocated declarations
// nested inside it.
func (a *AttributeContext) Enter(att *expr.AttributeExpr) *AttributeContext {
	entered := a.Dup()
	entered.Scope = a.Scope.Enter(att)
	return entered
}

// Dup creates a shallow copy of the AttributeContext.
func (a *AttributeContext) Dup() *AttributeContext {
	return &AttributeContext{
		Pointer:        a.Pointer,
		IgnoreRequired: a.IgnoreRequired,
		UseDefault:     a.UseDefault,
		Scope:          a.Scope,
		UnionPointer:   a.UnionPointer,
	}
}

// Name returns the type name for the given attribute.
func (a *AttributeScope) Name(att *expr.AttributeExpr, pkg string, ptr, useDefault bool) string {
	if _, ok := att.Type.(expr.UserType); !ok && expr.IsObject(att.Type) {
		// In the special case of anonymous / inline struct types the "name" is
		// in fact the struct typedef. In this case we need to force the
		// generation of the fields as pointers if needed as the default
		// GoTransform algorithm does not allow for an override.
		// Use the target package context so that user types referenced inside
		// the inline struct are qualified against the correct package (e.g.,
		// use "types.UUID" instead of incorrectly qualifying with the service
		// package alias).
		return a.scope.goTypeDefWithPkgOverride(att, ptr, useDefault, "", pkg)
	}
	if n, ok := att.Meta["struct:type:name"]; ok {
		// If the attribute has a "struct:type:name" meta then use it as the
		// type name.
		if pkg == "" {
			return n[0]
		}
		return pkg + "." + n[0]
	}
	return a.scope.GoFullTypeName(att, pkg)
}

// Ref returns the type name for the given attribute.
func (a *AttributeScope) Ref(att *expr.AttributeExpr, pkg string) string {
	return a.scope.GoFullTypeRef(att, pkg)
}

// Package returns the qualifier selected by att's explicit type location or
// the scope's default package.
func (a *AttributeScope) Package(att *expr.AttributeExpr) string {
	if att == nil {
		return a.pkg
	}
	if loc := UserTypeLocation(att.Type); loc != nil {
		return loc.PackageName()
	}
	return a.pkg
}

// ValidatorName returns the deterministic validator convention used by
// generators whose names are already isolated in a private transport scope.
func (a *AttributeScope) ValidatorName(att *expr.AttributeExpr, view string) string {
	return "Validate" + a.Name(att, "", false, true) + Goify(view, true)
}

// Enter returns a scope whose default qualifier follows att's explicit type
// location. The underlying name scope remains unchanged.
func (a *AttributeScope) Enter(att *expr.AttributeExpr) Attributor {
	if loc := UserTypeLocation(att.Type); loc != nil && loc.PackageName() != a.pkg {
		return newAttributeScope(a.scope, loc.PackageName())
	}
	return a
}

// IsSumType reports that AttributeScope renders unions using Goa's generated
// sum-type structs.
func (*AttributeScope) IsSumType() bool {
	return true
}

// Field returns a valid Go struct field name.
func (*AttributeScope) Field(att *expr.AttributeExpr, name string, firstUpper bool) string {
	return GoifyAtt(att, name, firstUpper)
}

// Scope returns the name scope.
func (a *AttributeScope) Scope() *NameScope {
	return a.scope
}

// newAttributeScope builds an attribute scope with explicit package
// qualification behavior.
func newAttributeScope(scope *NameScope, pkg string) *AttributeScope {
	return &AttributeScope{scope: scope, pkg: pkg}
}
