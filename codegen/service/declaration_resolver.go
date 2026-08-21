// This file resolves service type definitions and references through the
// frozen generated-package catalog. It follows explicit type locations by
// import path and keeps unlocated nested declarations in their enclosing
// package.
package service

import (
	"fmt"
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// declarationResolver renders service-side attributes from the package
	// records selected during Plan.
	declarationResolver struct {
		generation  *codegen.Generation
		aliases     *importAliases
		service     *expr.ServiceExpr
		currentPath string
		outputPath  string
		derived     map[expr.UserType]codegen.DerivedTypeID
		view        bool
	}

	// methodDeclarationAttributor binds one normalized method wrapper to its
	// frozen declaration while leaving all other transport naming unchanged.
	methodDeclarationAttributor struct {
		origin      expr.UserType
		declaration *codegen.TypeDeclaration
		delegate    codegen.Attributor
	}
)

// NewMethodTypeContext returns the service-side transport context for a named
// method type. The exact wrapper uses its frozen declaration; nested and wire
// attributes retain the transport's existing naming scope.
func NewMethodTypeContext(attribute *expr.AttributeExpr, declaration *codegen.TypeDeclaration, pkg string, scope *codegen.NameScope) *codegen.AttributeContext {
	userType, ok := attribute.Type.(expr.UserType)
	if !ok || declaration == nil {
		panic("method type context requires a named generated declaration")
	}
	delegate := codegen.NewAttributeContext(false, false, true, pkg, scope).Scope
	return &codegen.AttributeContext{
		UseDefault: true,
		Scope: &methodDeclarationAttributor{
			origin:      userType.Origin(),
			declaration: declaration,
			delegate:    delegate,
		},
	}
}

// newServiceResolver resolves declarations starting in service's generated
// package and qualifies names relative to outputPath.
func newServiceResolver(generation *codegen.Generation, aliases *importAliases, service *expr.ServiceExpr, outputPath string) *declarationResolver {
	return &declarationResolver{
		generation:  generation,
		aliases:     aliases,
		service:     service,
		currentPath: servicePackagePath(generation.GenPkg(), service),
		outputPath:  outputPath,
	}
}

// newViewResolver resolves every declaration in service's views package.
// derived binds rebuilt projected expression origins to their typed catalog
// identities.
func newViewResolver(generation *codegen.Generation, aliases *importAliases, service *expr.ServiceExpr, derived map[expr.UserType]codegen.DerivedTypeID) *declarationResolver {
	viewsPath := servicePackagePath(generation.GenPkg(), service) + "/views"
	return &declarationResolver{
		generation:  generation,
		aliases:     aliases,
		service:     service,
		currentPath: viewsPath,
		outputPath:  viewsPath,
		derived:     derived,
		view:        true,
	}
}

// Name returns the generated Go type name for att. Package ownership comes
// from the resolver's current import path, not from the textual pkg argument.
func (r *declarationResolver) Name(att *expr.AttributeExpr, _ string, ptr, useDefault bool) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		if custom, spec := codegen.GetMetaType(att); custom != "" {
			if spec == nil {
				return custom
			}
			_, typeName, qualified := strings.Cut(custom, ".")
			if !qualified {
				return custom
			}
			return r.aliases.name(spec.Path) + "." + typeName
		}
		return codegen.GoNativeTypeName(actual)
	case *expr.Array:
		return "[]" + r.Ref(actual.ElemType, "")
	case *expr.Map:
		return fmt.Sprintf("map[%s]%s", r.Ref(actual.KeyType, ""), r.Ref(actual.ElemType, ""))
	case *expr.Object:
		return r.Def(att, ptr, useDefault)
	case expr.UserType:
		if actual == expr.ErrorResult {
			return "goa.ServiceError"
		}
		owner := r.owner(att)
		declaration := r.userType(owner, actual)
		return r.qualify(owner, declaration.Name())
	case *expr.Union:
		owner := r.owner(att)
		declaration, err := r.generation.GeneratedPackage(owner).Union(actual)
		if err != nil {
			panic(fmt.Sprintf("resolve union %q for service %q in package %q: %v", actual.Name(), r.service.Name, owner, err))
		}
		return r.qualify(owner, declaration.Name())
	case expr.CompositeExpr:
		return r.Name(actual.Attribute(), "", ptr, useDefault)
	default:
		panic(fmt.Sprintf("resolve service type %T for service %q", actual, r.service.Name))
	}
}

// Def returns the Go definition for att while resolving every nested named
// declaration through its actual generated package.
func (r *declarationResolver) Def(att *expr.AttributeExpr, ptr, useDefault bool) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		return r.Name(att, "", ptr, useDefault)
	case *expr.Array:
		definition := r.Enter(actual.ElemType).(*declarationResolver).Def(actual.ElemType, ptr, useDefault)
		if expr.IsObject(actual.ElemType.Type) {
			definition = "*" + definition
		}
		return "[]" + definition
	case *expr.Map:
		key := r.Enter(actual.KeyType).(*declarationResolver).Def(actual.KeyType, ptr, useDefault)
		if expr.IsObject(actual.KeyType.Type) {
			key = "*" + key
		}
		value := r.Enter(actual.ElemType).(*declarationResolver).Def(actual.ElemType, ptr, useDefault)
		if expr.IsObject(actual.ElemType.Type) {
			value = "*" + value
		}
		return fmt.Sprintf("map[%s]%s", key, value)
	case *expr.Object:
		lines := []string{"struct {"}
		for _, field := range *actual {
			fieldResolver := r.Enter(field.Attribute).(*declarationResolver)
			definition := fieldResolver.Def(field.Attribute, ptr, useDefault)
			if serviceFieldIsPointer(att, field.Name, ptr, useDefault) {
				definition = "*" + definition
			}
			var description string
			if field.Attribute.Description != "" {
				description = codegen.Comment(field.Attribute.Description) + "\n\t"
			}
			lines = append(lines, fmt.Sprintf(
				"\t%s%s %s%s",
				description,
				codegen.GoifyAtt(field.Attribute, field.Name, true),
				definition,
				codegen.AttributeTagsWithName(att, field.Name, field.Attribute),
			))
		}
		return strings.Join(append(lines, "}"), "\n")
	case expr.UserType, *expr.Union:
		return r.Name(att, "", ptr, useDefault)
	case expr.CompositeExpr:
		return r.Def(actual.Attribute(), ptr, useDefault)
	default:
		panic(fmt.Sprintf("define service type %T for service %q", actual, r.service.Name))
	}
}

// Ref returns the generated Go reference for att.
func (r *declarationResolver) Ref(att *expr.AttributeExpr, pkg string) string {
	name := r.Name(att, pkg, false, false)
	if _, ok := att.Type.(*expr.Object); ok {
		return name
	}
	if expr.IsObject(att.Type) || expr.IsUnion(att.Type) {
		return "*" + name
	}
	return name
}

// Field returns the generated Go field name for one service attribute.
func (*declarationResolver) Field(att *expr.AttributeExpr, name string, firstUpper bool) string {
	return codegen.GoifyAtt(att, name, firstUpper)
}

// Package returns the qualifier for att relative to the file being rendered.
func (r *declarationResolver) Package(att *expr.AttributeExpr) string {
	owner := r.currentPath
	if att != nil {
		owner = r.owner(att)
	}
	if owner == r.outputPath {
		return ""
	}
	return r.aliases.name(owner)
}

// Enter returns a resolver whose current package owns att and its unlocated
// nested declarations.
func (r *declarationResolver) Enter(att *expr.AttributeExpr) codegen.Attributor {
	owner := r.owner(att)
	if owner == r.currentPath {
		return r
	}
	entered := *r
	entered.currentPath = owner
	return &entered
}

// inOutputPackage returns a resolver for a file emitted in packagePath.
func (r *declarationResolver) inOutputPackage(packagePath string) *declarationResolver {
	if packagePath == r.currentPath && packagePath == r.outputPath {
		return r
	}
	output := *r
	output.currentPath = packagePath
	output.outputPath = packagePath
	return &output
}

// withOutputPackage returns a resolver that keeps its current declaration
// owner but qualifies references for a file emitted in packagePath.
func (r *declarationResolver) withOutputPackage(packagePath string) *declarationResolver {
	if packagePath == r.outputPath {
		return r
	}
	output := *r
	output.outputPath = packagePath
	return &output
}

// bindDerived returns a resolver that associates a render-only expression
// origin with one declaration planned from its exact source type.
func (r *declarationResolver) bindDerived(origin expr.UserType, identity codegen.DerivedTypeID) *declarationResolver {
	bound := *r
	bound.derived = make(map[expr.UserType]codegen.DerivedTypeID, len(r.derived)+1)
	for existing, existingIdentity := range r.derived {
		bound.derived[existing] = existingIdentity
	}
	bound.derived[origin.Origin()] = identity
	return &bound
}

// IsSumType reports that service unions use Goa's generated sum-type structs.
func (*declarationResolver) IsSumType() bool {
	return true
}

// Scope returns the frozen name scope owned by the resolver's current package.
func (r *declarationResolver) Scope() *codegen.NameScope {
	return r.generation.GeneratedPackage(r.currentPath).Scope()
}

// owner returns the import path that owns att. View projections stay in the
// views package after their original struct:pkg:path metadata is removed.
func (r *declarationResolver) owner(att *expr.AttributeExpr) string {
	if r.view {
		return r.currentPath
	}
	if location := codegen.UserTypeLocation(att.Type); location != nil {
		return path.Join(r.generation.GenPkg(), location.RelImportPath)
	}
	return r.currentPath
}

// userType selects an exact, generated union branch, or rebuilt view record.
func (r *declarationResolver) userType(owner string, userType expr.UserType) *codegen.TypeDeclaration {
	generatedPackage := r.generation.GeneratedPackage(owner)
	if identity, ok := r.derived[userType.Origin()]; ok {
		declaration, err := generatedPackage.DerivedType(identity)
		if err != nil {
			panic(fmt.Sprintf("resolve derived type %q for service %q in package %q: %v", userType.Name(), r.service.Name, owner, err))
		}
		return declaration
	}
	declaration, err := generatedPackage.Type(userType)
	if err != nil {
		panic(fmt.Sprintf("resolve user type %q for service %q in package %q: %v", userType.Name(), r.service.Name, owner, err))
	}
	return declaration
}

// qualify adds the owning package name when the current output file is in a
// different generated package.
func (r *declarationResolver) qualify(owner, name string) string {
	if owner == r.outputPath {
		return name
	}
	return r.aliases.name(owner) + "." + name
}

// refDeclaration qualifies declaration for the resolver's output file while
// preserving the pointer or value semantics of dataType.
func (r *declarationResolver) refDeclaration(declaration *codegen.TypeDeclaration, dataType expr.DataType) string {
	qualified := r.qualify(declaration.PackagePath(), declaration.Name())
	if strings.HasPrefix(declaration.Ref(dataType), "*") {
		return "*" + qualified
	}
	return qualified
}

// declarationName returns the unqualified planned name for one named type.
func (r *declarationResolver) declarationName(attribute *expr.AttributeExpr) string {
	entered := r.Enter(attribute).(*declarationResolver)
	return entered.userType(entered.currentPath, attribute.Type.(expr.UserType)).Name()
}

// Name returns the frozen wrapper name for the bound method type and delegates
// every other attribute to the transport's existing scope.
func (a *methodDeclarationAttributor) Name(attribute *expr.AttributeExpr, pkg string, pointer, useDefault bool) string {
	if a.matches(attribute) {
		if pkg == "" {
			return a.declaration.Name()
		}
		return pkg + "." + a.declaration.Name()
	}
	return a.delegate.Name(attribute, pkg, pointer, useDefault)
}

// Ref returns the frozen wrapper reference for the bound method type and
// delegates every other attribute to the transport's existing scope.
func (a *methodDeclarationAttributor) Ref(attribute *expr.AttributeExpr, pkg string) string {
	if !a.matches(attribute) {
		return a.delegate.Ref(attribute, pkg)
	}
	name := a.Name(attribute, pkg, false, false)
	if expr.IsObject(attribute.Type) || expr.IsUnion(attribute.Type) {
		return "*" + name
	}
	return name
}

// Field delegates service field naming to the transport's existing scope.
func (a *methodDeclarationAttributor) Field(attribute *expr.AttributeExpr, name string, firstUpper bool) string {
	return a.delegate.Field(attribute, name, firstUpper)
}

// Package delegates package qualification to the transport's existing scope.
func (a *methodDeclarationAttributor) Package(attribute *expr.AttributeExpr) string {
	return a.delegate.Package(attribute)
}

// Enter keeps the frozen binding for the exact wrapper and delegates nested
// attributes to the transport's existing package rules.
func (a *methodDeclarationAttributor) Enter(attribute *expr.AttributeExpr) codegen.Attributor {
	if a.matches(attribute) {
		return a
	}
	return a.delegate.Enter(attribute)
}

// IsSumType preserves the transport scope's union representation.
func (a *methodDeclarationAttributor) IsSumType() bool {
	return a.delegate.IsSumType()
}

// Scope returns the transport naming scope used for all unbound attributes.
func (a *methodDeclarationAttributor) Scope() *codegen.NameScope {
	return a.delegate.Scope()
}

// matches reports whether attribute is the exact normalized wrapper bound to
// this rendering context.
func (a *methodDeclarationAttributor) matches(attribute *expr.AttributeExpr) bool {
	userType, ok := attribute.Type.(expr.UserType)
	return ok && userType.Origin() == a.origin
}

// serviceFieldIsPointer matches Goa service struct pointer semantics for one
// field definition.
func serviceFieldIsPointer(parent *expr.AttributeExpr, name string, pointer, useDefault bool) bool {
	field := expr.AsObject(parent.Type).Attribute(name)
	return expr.IsObject(field.Type) ||
		parent.IsPrimitivePointer(name, useDefault) ||
		pointer && expr.IsPrimitive(field.Type) && field.Type.Kind() != expr.AnyKind && field.Type.Kind() != expr.BytesKind
}
