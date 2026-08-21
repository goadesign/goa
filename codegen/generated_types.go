// This file defines the declaration catalog owned by one generated Go
// package. Planning reserves every public type name here before generators use
// the frozen records to render declarations and references.
package codegen

import (
	"fmt"
	"slices"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// GeneratedPackage owns type declarations and their shared naming scope for
	// one generated Go package.
	GeneratedPackage struct {
		path          string
		scope         *NameScope
		userTypes     map[expr.UserType]*TypeDeclaration
		unions        map[UnionTypeID]*unionDeclaration
		userTypeNames map[string]string
		frozen        bool
	}

	// TypeDeclaration records the canonical name and package path of one
	// generated type declaration.
	TypeDeclaration struct {
		// Name is the unqualified Go declaration name. Generated union branch
		// aliases keep Name empty until the owning generation is frozen.
		Name string
		// PackagePath is the import path of the package that owns the declaration.
		PackagePath string
	}

	// UnionDeclaration records the canonical union and discriminator names in
	// the package that emits them.
	UnionDeclaration struct {
		// Name is the unqualified Go union name. It remains empty until the
		// owning generation is frozen.
		Name string
		// KindName is the unqualified Go discriminator type name. It remains
		// empty until the owning generation is frozen.
		KindName string
		// PackagePath is the import path of the package that owns both names.
		PackagePath string
	}

	// unionDeclaration retains the expression needed to allocate the public
	// union name deterministically when the generation freezes.
	unionDeclaration struct {
		union       *expr.Union
		declaration *UnionDeclaration
		branches    map[unionBranchID]*unionBranchDeclaration
	}

	// unionBranchID identifies one generated branch alias within its union
	// declaration family.
	unionBranchID struct {
		name string
	}

	// unionBranchDeclaration keeps every expression copy that refers to one
	// branch alias while owning a single emitted declaration.
	unionBranchDeclaration struct {
		userTypes   map[expr.UserType]struct{}
		declaration *TypeDeclaration
		name        string
	}
)

// Ref returns the Go reference spelling for declaration's data type, including
// Goa's pointer/value semantics for named objects, unions, and aliases.
func (d *TypeDeclaration) Ref(dataType expr.DataType) string {
	return goTypeRef(d.Name, dataType)
}

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
func (p *GeneratedPackage) DeclareUnion(union *expr.Union) (*UnionDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	identity := NewUnionTypeID(union)
	if planned, ok := p.unions[identity]; ok {
		return planned.declaration, nil
	}

	declaration := &UnionDeclaration{PackagePath: p.path}
	p.unions[identity] = &unionDeclaration{
		union:       union,
		declaration: declaration,
		branches:    make(map[unionBranchID]*unionBranchDeclaration),
	}
	return declaration, nil
}

// DeclareUnionBranchType records a generated user type that names one branch
// of union. Equivalent union expressions share the same branch declaration;
// ordinary DSL user types must instead be declared with DeclareUserType.
func (p *GeneratedPackage) DeclareUnionBranchType(union *expr.Union, branchName string, userType expr.UserType) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	planned, ok := p.unions[NewUnionTypeID(union)]
	if !ok {
		return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
	}
	if !unionHasBranchType(union, branchName, userType) {
		return nil, fmt.Errorf("user type %q is not branch %q of union %q", userType.Name(), branchName, union.Name())
	}

	identity := unionBranchID{name: branchName}
	if branch, ok := planned.branches[identity]; ok {
		name := Goify(userType.Name(), true)
		if branch.name != name {
			return nil, fmt.Errorf(
				"branch %q of union %q cannot declare both %q and %q",
				branchName,
				union.Name(),
				branch.name,
				name,
			)
		}
		branch.userTypes[userType] = struct{}{}
		return branch.declaration, nil
	}
	declaration := &TypeDeclaration{PackagePath: p.path}
	planned.branches[identity] = &unionBranchDeclaration{
		userTypes:   map[expr.UserType]struct{}{userType: {}},
		declaration: declaration,
		name:        Goify(userType.Name(), true),
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
func (p *GeneratedPackage) Union(union *expr.Union) (*UnionDeclaration, error) {
	if planned, ok := p.unions[NewUnionTypeID(union)]; ok {
		return planned.declaration, nil
	}
	return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
}

// UnionBranchType returns the existing declaration for one generated branch
// alias without allocating a name or declaration record.
func (p *GeneratedPackage) UnionBranchType(union *expr.Union, branchName string, userType expr.UserType) (*TypeDeclaration, error) {
	planned, ok := p.unions[NewUnionTypeID(union)]
	if !ok {
		return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
	}
	branch, ok := planned.branches[unionBranchID{name: branchName}]
	if !ok {
		return nil, fmt.Errorf("branch %q of union %q is not declared in generated package %q", branchName, union.Name(), p.path)
	}
	if _, ok := branch.userTypes[userType]; !ok {
		return nil, fmt.Errorf("user type %q is not declared as branch %q of union %q", userType.Name(), branchName, union.Name())
	}
	return branch.declaration, nil
}

// Scope returns the frozen package-owned name scope used to render generated
// references. It panics before declaration planning has been frozen.
func (p *GeneratedPackage) Scope() *NameScope {
	if !p.frozen {
		panic(fmt.Sprintf("generated package %q scope requested before freeze", p.path))
	}
	return p.scope
}

// newGeneratedPackage creates an empty mutable declaration catalog for path.
func newGeneratedPackage(path string) *GeneratedPackage {
	return &GeneratedPackage{
		path:          path,
		scope:         NewNameScope(),
		userTypes:     make(map[expr.UserType]*TypeDeclaration),
		unions:        make(map[UnionTypeID]*unionDeclaration),
		userTypeNames: make(map[string]string),
	}
}

// freeze assigns pending union names in structural-identity order, then ends
// declaration and scope mutation while preserving read-only lookups.
func (p *GeneratedPackage) freeze() {
	identities := make([]UnionTypeID, 0, len(p.unions))
	for identity := range p.unions {
		identities = append(identities, identity)
	}
	slices.Sort(identities)
	for _, identity := range identities {
		planned := p.unions[identity]
		name := p.scope.HashedUnique(identity, Goify(planned.union.Name(), true), "")
		planned.declaration.Name = name
		planned.declaration.KindName = p.scope.Unique(name + "Kind")

		branches := make([]unionBranchID, 0, len(planned.branches))
		for branch := range planned.branches {
			branches = append(branches, branch)
		}
		slices.SortFunc(branches, func(a, b unionBranchID) int {
			return strings.Compare(a.name, b.name)
		})
		for _, identity := range branches {
			branch := planned.branches[identity]
			branch.declaration.Name = p.scope.Unique(branch.name)
		}
	}
	p.scope.Freeze()
	p.frozen = true
}

// unionHasBranchType verifies that userType is the branch expression supplied
// for this concrete union copy. Structural reuse is established separately by
// UnionTypeID when the owning union declaration is looked up.
func unionHasBranchType(union *expr.Union, branchName string, userType expr.UserType) bool {
	for _, branch := range union.Values {
		if branch.Name == branchName && branch.Attribute.Type == userType {
			return true
		}
	}
	return false
}
