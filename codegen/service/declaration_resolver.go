// This file writes service type definitions and references using the Go names
// chosen for each generated package. A type with an explicit package location
// uses that import path; a child type without one stays in its enclosing type's
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
	// declarationResolver writes service fields and type references from the
	// package declarations recorded by Plan.
	declarationResolver struct {
		generation  *codegen.Generation
		aliases     *importAliases
		serviceName string
		currentPath string
		outputPath  string
		derived     map[expr.UserType]codegen.DerivedTypeID
		validators  map[validatorKey]*codegen.NameDeclaration
		view        bool
	}
)

// newServiceResolver starts in the assigned service package and qualifies type
// references for the package that will contain the generated file.
func newServiceResolver(generation *codegen.Generation, aliases *importAliases, serviceName, servicePath, outputPath string) *declarationResolver {
	return &declarationResolver{
		generation:  generation,
		aliases:     aliases,
		serviceName: serviceName,
		currentPath: servicePath,
		outputPath:  outputPath,
	}
}

// newViewResolver starts in the assigned views package and associates each
// generated view type with the service result type from which it was built.
func newViewResolver(generation *codegen.Generation, aliases *importAliases, serviceName, viewsPath string, derived map[expr.UserType]codegen.DerivedTypeID) *declarationResolver {
	return &declarationResolver{
		generation:  generation,
		aliases:     aliases,
		serviceName: serviceName,
		currentPath: viewsPath,
		outputPath:  viewsPath,
		derived:     derived,
		view:        true,
	}
}

// Name returns the generated Go type name for att. The resolver's current
// import path chooses the package containing the declaration; the textual pkg
// argument does not.
func (r *declarationResolver) Name(att *expr.AttributeExpr, _ string, ptr, useDefault bool) string {
	switch actual := att.Type.(type) {
	case expr.Primitive:
		return r.generation.Package(r.outputPath).Scope().GoTypeName(att)
	case *expr.Array:
		return "[]" + r.Ref(actual.ElemType, "")
	case *expr.Map:
		return fmt.Sprintf("map[%s]%s", r.Ref(actual.KeyType, ""), r.Ref(actual.ElemType, ""))
	case *expr.Object:
		return r.Def(att, ptr, useDefault)
	case expr.UserType:
		if actual == expr.Empty {
			return "struct {}"
		}
		if expr.IsErrorResult(actual) {
			return "goa.ServiceError"
		}
		owner := r.owner(att)
		declaration := r.userType(owner, actual)
		return r.qualify(owner, declaration.Name())
	case *expr.Union:
		owner := r.owner(att)
		declaration, err := r.generation.Package(owner).Union(att)
		if err != nil {
			panic(fmt.Sprintf("resolve union %q for service %q in package %q: %v", actual.Name(), r.serviceName, owner, err))
		}
		return r.qualify(owner, declaration.Name())
	case expr.CompositeExpr:
		return r.Name(actual.Attribute(), "", ptr, useDefault)
	default:
		panic(fmt.Sprintf("resolve service type %T for service %q", actual, r.serviceName))
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
		panic(fmt.Sprintf("define service type %T for service %q", actual, r.serviceName))
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

// UnionConstructor returns the planned constructor for one service OneOf
// branch, qualified for the generated file that will call it.
func (r *declarationResolver) UnionConstructor(attribute *expr.AttributeExpr, branch string) (string, error) {
	owner := r.owner(attribute)
	declaration, err := r.generation.Package(owner).UnionBranch(attribute, branch)
	if err != nil {
		return "", err
	}
	return r.qualify(owner, declaration.Constructor()), nil
}

// GoTypeLayout returns the retained Go fields and references selected by the
// service package for attribute.
func (r *declarationResolver) GoTypeLayout(attribute *expr.AttributeExpr, policy codegen.GoLayoutPolicy) (codegen.LinkedGoType, error) {
	owner := r.owner(attribute)
	layout, err := codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
		Owner:            owner,
		Policy:           policy,
		RetainNamedValue: true,
		Bind: func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
			bindingOwner := request.InheritedOwner
			if location := codegen.UserTypeLocation(request.Attribute.Type); location != nil {
				bindingOwner = path.Join(r.generation.GenPkg(), location.RelImportPath)
			}
			switch request.Kind {
			case codegen.GoNamed:
				userType := request.Attribute.Type.(expr.UserType)
				return codegen.GoTypeBinding{
					Owner: bindingOwner,
					Type:  r.userType(bindingOwner, userType),
				}, nil
			case codegen.GoUnion:
				declaration, lookupErr := r.generation.Package(bindingOwner).Union(request.Attribute)
				if lookupErr != nil {
					return codegen.GoTypeBinding{}, lookupErr
				}
				return codegen.GoTypeBinding{Owner: bindingOwner, Union: declaration}, nil
			default:
				return codegen.GoTypeBinding{}, fmt.Errorf("resolve unsupported service Go type kind %s", request.Kind)
			}
		},
	})
	if err != nil {
		return codegen.LinkedGoType{}, err
	}
	return layout.Link(r.outputPath, func(importPath string) string {
		return r.aliases.name(r.outputPath, importPath)
	}), nil
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
	return r.aliases.name(r.outputPath, owner)
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

// withOutputPackage returns a resolver that keeps declarations in the current
// package but qualifies references for a file emitted in packagePath.
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

// withValidators returns a resolver that maps each child validation call to the
// Go function declaration submitted during service planning.
func (r *declarationResolver) withValidators(validators map[validatorKey]*codegen.NameDeclaration) *declarationResolver {
	bound := *r
	bound.validators = validators
	return &bound
}

// IsSumType reports that service unions use generated values that hold one branch.
func (*declarationResolver) IsSumType() bool {
	return true
}

// ValidatorCall returns a call to the validation function submitted for att
// and view before Generation.Freeze chose the function's Go name.
func (r *declarationResolver) ValidatorCall(att *expr.AttributeExpr, view, target, _ string) string {
	declaration := r.validatorDeclaration(att, view)
	return fmt.Sprintf("%s(%s)", r.qualify(r.owner(att), declaration.Name()), target)
}

// validatorDeclaration returns the NameDeclaration recorded for att and the
// selected view. The default view uses the same empty key as its call sites.
func (r *declarationResolver) validatorDeclaration(att *expr.AttributeExpr, view string) *codegen.NameDeclaration {
	userType, ok := att.Type.(expr.UserType)
	if !ok {
		panic(fmt.Sprintf("resolve validator for non-user type %T", att.Type))
	}
	owner := r.owner(att)
	declaration := r.userType(owner, userType)
	validator := r.validators[validatorKey{declaration: declaration, view: canonicalValidatorView(view)}]
	if validator == nil {
		panic(fmt.Sprintf(
			"validator for type %q view %q was not retained in generated package %q",
			userType.Name(), view, owner,
		))
	}
	return validator
}

// Scope returns the name set for the resolver's current generated package.
func (r *declarationResolver) Scope() *codegen.NameScope {
	return r.generation.Package(r.currentPath).Scope()
}

// owner returns the import path of the package containing att. View-specific
// result copies stay in the views package after Goa removes their original
// struct:pkg:path metadata.
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
	generatedPackage := r.generation.Package(owner)
	if identity, ok := r.derived[userType.Origin()]; ok {
		declaration, err := generatedPackage.DerivedType(identity)
		if err != nil {
			panic(fmt.Sprintf("resolve derived type %q for service %q in package %q: %v", userType.Name(), r.serviceName, owner, err))
		}
		return declaration
	}
	declaration, err := generatedPackage.Type(userType)
	if err != nil {
		panic(fmt.Sprintf("resolve user type %q for service %q in package %q: %v", userType.Name(), r.serviceName, owner, err))
	}
	return declaration
}

// qualify adds the owning package name when the current output file is in a
// different generated package.
func (r *declarationResolver) qualify(owner, name string) string {
	if owner == r.outputPath {
		return name
	}
	return r.aliases.name(r.outputPath, owner) + "." + name
}

// serviceFieldIsPointer applies Goa's service-struct pointer rules to one field
// definition.
func serviceFieldIsPointer(parent *expr.AttributeExpr, name string, pointer, useDefault bool) bool {
	field := expr.AsObject(parent.Type).Attribute(name)
	return expr.IsObject(field.Type) ||
		parent.IsPrimitivePointer(name, useDefault) ||
		pointer && expr.IsPrimitive(field.Type) && field.Type.Kind() != expr.AnyKind && field.Type.Kind() != expr.BytesKind
}
