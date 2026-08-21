// This file defines the declaration catalog owned by one generated Go
// package. Planning reserves every public type name here before generators use
// the frozen records to render declarations and references.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/expr"
)

type (
	// GeneratedPackage owns type declarations and their shared naming scope for
	// one generated Go package.
	GeneratedPackage struct {
		path         string
		scope        *NameScope
		userTypes    map[expr.UserType]*TypeDeclaration
		unions       map[string]*TypeDeclaration
		declarations map[string]declarationOwner
		frozen       bool
	}

	// TypeDeclaration records the canonical name and package path of one
	// generated type declaration.
	TypeDeclaration struct {
		// Name is the unqualified Go declaration name.
		Name string
		// PackagePath is the import path of the package that owns the declaration.
		PackagePath string
	}

	// declarationOwner describes the design expression that reserved a public
	// name so collision errors can identify both declarations.
	declarationOwner struct {
		kind string
		name string
	}
)

// DeclareUserType reserves userType's exact exported Go name and returns its
// canonical package declaration. Repeated calls return the same declaration.
func (p *GeneratedPackage) DeclareUserType(userType expr.UserType) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	if declaration, ok := p.userTypes[userType]; ok {
		return declaration, nil
	}

	name := Goify(userType.Name(), true)
	if owner, ok := p.declarations[name]; ok {
		return nil, fmt.Errorf(
			"generated package %q cannot declare user type %q as %q: already declared by %s %q",
			p.path,
			userType.Name(),
			name,
			owner.kind,
			owner.name,
		)
	}
	if p.scope.PeekUnique(name) != name {
		return nil, fmt.Errorf(
			"generated package %q cannot declare user type %q as %q: name is already reserved",
			p.path,
			userType.Name(),
			name,
		)
	}
	p.scope.HashedUnique(userType, name, "")
	declaration := &TypeDeclaration{Name: name, PackagePath: p.path}
	p.userTypes[userType] = declaration
	p.declarations[name] = declarationOwner{kind: "user type", name: userType.Name()}
	return declaration, nil
}

// DeclareUnion reserves a canonical name for union's emitted definition and
// returns the same declaration for unions with the same emitted identity.
func (p *GeneratedPackage) DeclareUnion(union *expr.Union) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	identity := UnionTypeHash(union)
	if declaration, ok := p.unions[identity]; ok {
		return declaration, nil
	}

	name := p.scope.HashedUnique(union, Goify(union.Name(), true), "")
	declaration := &TypeDeclaration{Name: name, PackagePath: p.path}
	p.unions[identity] = declaration
	p.declarations[name] = declarationOwner{kind: "union", name: union.Name()}
	return declaration, nil
}

// UserType returns userType's existing package declaration without allocating
// a name or declaration record.
func (p *GeneratedPackage) UserType(userType expr.UserType) (*TypeDeclaration, error) {
	if declaration, ok := p.userTypes[userType]; ok {
		return declaration, nil
	}
	return nil, fmt.Errorf("user type %q is not declared in generated package %q", userType.Name(), p.path)
}

// Union returns union's existing package declaration without allocating a
// name or declaration record.
func (p *GeneratedPackage) Union(union *expr.Union) (*TypeDeclaration, error) {
	if declaration, ok := p.unions[UnionTypeHash(union)]; ok {
		return declaration, nil
	}
	return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
}

// Scope returns the package-owned name scope shared by declaration planning
// and generated references.
func (p *GeneratedPackage) Scope() *NameScope {
	return p.scope
}

// newGeneratedPackage creates an empty mutable declaration catalog for path.
func newGeneratedPackage(path string) *GeneratedPackage {
	return &GeneratedPackage{
		path:         path,
		scope:        NewNameScope(),
		userTypes:    make(map[expr.UserType]*TypeDeclaration),
		unions:       make(map[string]*TypeDeclaration),
		declarations: make(map[string]declarationOwner),
	}
}

// freeze ends declaration planning while preserving read-only lookups.
func (p *GeneratedPackage) freeze() {
	p.frozen = true
}
