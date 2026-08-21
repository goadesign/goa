// This file defines the declaration catalog owned by one generated Go
// package. Planning reserves every public type name here before generators use
// the frozen records to render declarations and references.
package codegen

import (
	"fmt"
	"sort"

	"goa.design/goa/v3/expr"
)

type (
	// GeneratedPackage owns type declarations and their shared naming scope for
	// one generated Go package.
	GeneratedPackage struct {
		path          string
		scope         *NameScope
		userTypes     map[expr.UserType]*TypeDeclaration
		unions        map[string]*unionDeclaration
		userTypeNames map[string]string
		frozen        bool
	}

	// TypeDeclaration records the canonical name and package path of one
	// generated type declaration.
	TypeDeclaration struct {
		// Name is the unqualified Go declaration name. Union declarations keep
		// Name empty until the owning generation is frozen.
		Name string
		// PackagePath is the import path of the package that owns the declaration.
		PackagePath string
	}

	// unionDeclaration retains the expression needed to allocate the public
	// union name deterministically when the generation freezes.
	unionDeclaration struct {
		union       *expr.Union
		declaration *TypeDeclaration
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
	if declaredName, ok := p.userTypeNames[name]; ok {
		return nil, fmt.Errorf(
			"generated package %q cannot declare user type %q as %q: already declared by user type %q",
			p.path,
			userType.Name(),
			name,
			declaredName,
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
	p.userTypeNames[name] = userType.Name()
	return declaration, nil
}

// DeclareUnion records union's emitted definition and returns the same
// declaration for unions with the same emitted identity. The declaration name
// remains empty until the owning generation freezes its package catalogs.
func (p *GeneratedPackage) DeclareUnion(union *expr.Union) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	identity := UnionTypeHash(union)
	if planned, ok := p.unions[identity]; ok {
		return planned.declaration, nil
	}

	declaration := &TypeDeclaration{PackagePath: p.path}
	p.unions[identity] = &unionDeclaration{
		union:       union,
		declaration: declaration,
	}
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
	if planned, ok := p.unions[UnionTypeHash(union)]; ok {
		return planned.declaration, nil
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
		path:          path,
		scope:         NewNameScope(),
		userTypes:     make(map[expr.UserType]*TypeDeclaration),
		unions:        make(map[string]*unionDeclaration),
		userTypeNames: make(map[string]string),
	}
}

// freeze assigns pending union names in structural-identity order, then ends
// declaration and scope mutation while preserving read-only lookups.
func (p *GeneratedPackage) freeze() {
	identities := make([]string, 0, len(p.unions))
	for identity := range p.unions {
		identities = append(identities, identity)
	}
	sort.Strings(identities)
	for _, identity := range identities {
		planned := p.unions[identity]
		name := p.scope.HashedUnique(planned.union, Goify(planned.union.Name(), true), "")
		planned.declaration.Name = name
	}
	p.scope.Freeze()
	p.frozen = true
}
