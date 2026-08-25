// This file defines the naming and pointer contracts shared by Go
// transformation and validation generators.
package codegen

import (
	"fmt"
	"strings"

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
		// Enter returns the resolver for the package containing att and declarations
		// nested in it.
		Enter(att *expr.AttributeExpr) Attributor
		// IsSumType reports whether unions use Goa's generated sum-type layout.
		IsSumType() bool
		// ValidatorCall returns the complete call that validates target as att.
		// path is the generated expression used as the start of nested error paths.
		ValidatorCall(att *expr.AttributeExpr, view, target, path string) string
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
		// ArrayElementPointer keeps primitive array elements as pointers when
		// generated validation must distinguish null from the primitive zero value.
		ArrayElementPointer bool
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
		Hooks           *TransformHooks
		helpers         map[TransformHelperID]TransformHelper
		calls           *transformCallCursor
		collectionDepth int
		unionDepth      int
	}

	// TransformHelperID selects one recursive function in a TransformPlan. Its
	// fields are private so callers cannot rebuild it from a generated name.
	TransformHelperID struct {
		plan  *TransformPlan
		index int
	}

	// TransformHelperDefinitionID selects one function body in a TransformPlan.
	// Its fields are private so callers cannot rebuild it from type names.
	TransformHelperDefinitionID struct {
		plan  *TransformPlan
		index int
	}

	// TransformHelperDefinitionLocation orders function bodies by the authored
	// field or collection position where they are used. Its representation is
	// private and is never used as a generated Go name.
	TransformHelperDefinitionLocation struct {
		encoded string
	}

	// TransformHelperDefinition describes one function body shared by equivalent
	// helper calls in a TransformPlan. Requiredness belongs to each call and does
	// not create a different definition.
	TransformHelperDefinition struct {
		// ID selects this function body in its TransformPlan.
		ID TransformHelperDefinitionID
		// Source describes the source attribute converted by the function.
		// HelperDefinitions returns a detached copy, so changing it does not affect
		// Render.
		Source *expr.AttributeExpr
		// Target describes the target attribute produced by the function.
		// HelperDefinitions returns a detached copy, so changing it does not affect
		// Render.
		Target *expr.AttributeExpr
		// Location provides a stable order for this function body.
		Location TransformHelperDefinitionLocation
		// Declaration is populated when BindHelperDefinition assigns the
		// package-level function name. It remains nil when callers bind only the
		// individual helper occurrences.
		Declaration *NameDeclaration
		helpers     []int
	}

	// TransformHelper describes one generated function that converts a nested or
	// recursive value. The same chosen function name is used at every call and at
	// its definition.
	TransformHelper struct {
		// ID selects this function in its TransformPlan.
		ID TransformHelperID
		// Source describes the source attribute selected for this function.
		// Helpers returns a detached copy, so changing it does not affect Render.
		Source *expr.AttributeExpr
		// Target describes the target attribute selected for this function.
		// Helpers returns a detached copy, so changing it does not affect Render.
		Target *expr.AttributeExpr
		// Required reports whether the caller reaches this function through a
		// required value. Optional callers check for nil before calling it.
		Required bool
		// Occurrence is the one-based position of this helper operation in the
		// transform plan's stable traversal.
		Occurrence int
		// Declaration holds the package-level function name chosen before source
		// is written. Render returns an error when it is missing.
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
		// ID selects the recursive function rendered by this value.
		ID TransformHelperID
		// Declaration is the package-level function chosen before writing code.
		// It is nil when GoTransformWithAttrs created this value while writing
		// code.
		Declaration *NameDeclaration
		// Name is the final helper name kept for existing plugins.
		//
		// Deprecated: Use Declaration.Name() after planning.
		Name string
		// ParamTypeRef is the generated Go reference to the helper parameter type.
		ParamTypeRef string
		// ResultTypeRef is the generated Go reference to the helper result type.
		ResultTypeRef string
		// Code is the helper body.
		Code string
	}

	// TransformPlan owns copied source and target expressions plus every
	// recursive function needed to convert between them. Create a plan, inspect
	// the detached helper descriptions from Helpers and declare their names, bind
	// those declarations and the completed type resolvers, then call Render.
	// A helper ID remains bound to this plan, but changing a description returned
	// by Helpers cannot change the private expressions Render uses. Render caches
	// each exact argument set, so repeated calls return the first generated code.
	TransformPlan struct {
		source         *expr.AttributeExpr
		target         *expr.AttributeExpr
		sourceBaseline *expr.AttributeExpr
		targetBaseline *expr.AttributeExpr
		sourceCopier   *expr.AttributeGraphCopier
		targetCopier   *expr.AttributeGraphCopier
		prefix         string
		hooks          *TransformHooks
		sourceCtx      *AttributeContext
		targetCtx      *AttributeContext
		helpers        []TransformHelper
		definitions    []TransformHelperDefinition
		operations     []*transformOperation
		renders        map[transformRenderRequest]transformRenderResult
	}

	// transformRenderRequest identifies one Render invocation. Repeating the
	// same invocation returns its first result instead of invoking hooks again.
	transformRenderRequest struct {
		sourceVar string
		targetVar string
		newVar    bool
	}

	// transformRenderResult is the private immutable cache for one Render call.
	transformRenderResult struct {
		code    string
		helpers []*TransformFunctionData
		err     error
	}

	// transformPair holds the exact copied source and target types whose fields
	// are currently being visited.
	transformPair struct {
		source expr.DataType
		target expr.DataType
	}

	// transformOperation stores the recursive calls made by the top-level
	// conversion or one function body, in call order.
	transformOperation struct {
		calls []transformCall
	}

	// transformCall selects the recursive function used by one call.
	transformCall struct {
		helper TransformHelperID
	}

	// transformCallCursor counts the planned calls used by one render.
	transformCallCursor struct {
		calls []transformCall
		next  int
	}
)

const (
	transformObjectFieldLocation byte = iota + 1
	transformArrayElementLocation
	transformMapKeyLocation
	transformMapValueLocation
	transformUnionBranchLocation
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

// Compare returns -1, 0, or 1 when this authored location sorts before, at,
// or after other.
func (l TransformHelperDefinitionLocation) Compare(other TransformHelperDefinitionLocation) int {
	return strings.Compare(l.encoded, other.encoded)
}

// EnterCollection returns the loop variable for the current array and a copy
// used to render values nested inside that array.
func (a *TransformAttrs) EnterCollection() (string, *TransformAttrs) {
	child := *a
	child.collectionDepth++
	return string(rune('i' + a.collectionDepth)), &child
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

// AppendHelpers appends functions from newH that oldH does not already contain.
// Planned functions are the same when they use the same package declaration.
// Older functions without a declaration are the same when their names match.
// It panics when one declaration or released name has different parameter,
// result, or body text because one Go function cannot implement both values.
func AppendHelpers(oldH, newH []*TransformFunctionData) []*TransformFunctionData {
	for _, h := range newH {
		found := false
		for _, h2 := range oldH {
			if sameTransformHelper(h, h2) {
				if !transformFunctionDefinitionsEqual(h, h2) {
					panic(fmt.Sprintf("transform helper %q has different definitions", h.Name))
				}
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

// sameTransformHelper compares the chosen package declarations when both
// helpers have one. Values created while writing code have no declaration and
// keep using their generated names.
func sameTransformHelper(left, right *TransformFunctionData) bool {
	if left.Declaration != nil && right.Declaration != nil {
		return left.Declaration == right.Declaration
	}
	if left.Declaration != nil || right.Declaration != nil {
		return false
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
	if at := att.Find(name); at != nil && expr.IsPrimitive(at.Type) {
		kind := unalias(at.Type).Kind()
		if kind == expr.AnyKind || kind == expr.BytesKind {
			return false
		}
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

// IsArrayElementPointer reports whether primitive elements in array use
// pointers so generated validation can reject null before conversion.
func (a *AttributeContext) IsArrayElementPointer(array *expr.Array) bool {
	return arrayElementIsPointer(array, a.ArrayElementPointer)
}

// LayoutPolicy returns the exact pointer and default rules used by this
// generated value.
func (a *AttributeContext) LayoutPolicy() GoLayoutPolicy {
	return GoLayoutPolicy{
		Pointer:             a.Pointer,
		IgnoreRequired:      a.IgnoreRequired,
		UseDefault:          a.UseDefault,
		UnionPointer:        a.UnionPointer,
		ArrayElementPointer: a.ArrayElementPointer,
		SumType:             a.Scope.IsSumType(),
	}
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
		Pointer:             a.Pointer,
		IgnoreRequired:      a.IgnoreRequired,
		UseDefault:          a.UseDefault,
		Scope:               a.Scope,
		UnionPointer:        a.UnionPointer,
		ArrayElementPointer: a.ArrayElementPointer,
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

// GoTypeLayout records the exact names from this scope with the pointer policy
// selected by the caller.
func (a *AttributeScope) GoTypeLayout(attribute *expr.AttributeExpr, policy GoLayoutPolicy) (LinkedGoType, error) {
	return planGoTypeWithAttributor(attribute, policy, a)
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

// ValidatorCall returns a call to the validation function selected from the
// generated type and view names.
func (a *AttributeScope) ValidatorCall(att *expr.AttributeExpr, view, target, _ string) string {
	name := "Validate" + a.Name(att, "", false, true) + Goify(view, true)
	return fmt.Sprintf("%s(%s)", name, target)
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
