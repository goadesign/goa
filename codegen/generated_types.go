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
		typeBindings  map[expr.UserType]*TypeDeclaration
		derivedTypes  map[DerivedTypeID]*derivedTypeDeclaration
		unions        map[UnionTypeID]*unionDeclaration
		userTypeNames map[string]string
		derivedKeys   map[derivedTypeOrder]DerivedTypeID
		frozen        bool
	}

	// DerivedTypeID identifies a generated view declaration by the exact source
	// declaration and the closed transformation that produces it.
	DerivedTypeID struct {
		origin expr.UserType
		kind   derivedTypeKind
	}

	// TypeDeclaration records the canonical name and package path of one
	// generated type declaration.
	TypeDeclaration struct {
		// Name is the unqualified Go declaration name. Derived view types and
		// generated union branch aliases keep Name empty until the owning
		// generation is frozen.
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

	// UnionBranchDeclaration records the package-level declarations emitted for
	// one union branch.
	UnionBranchDeclaration struct {
		// KindConst is the unqualified discriminator constant name.
		KindConst string
		// Constructor is the unqualified constructor function name.
		Constructor string
		// Type is the optional generated alias declaration for the branch.
		Type *TypeDeclaration

		typeName string
	}

	// unionDeclaration retains the expression needed to allocate the public
	// union name deterministically when the generation freezes.
	unionDeclaration struct {
		union       *expr.Union
		declaration *UnionDeclaration
		branches    map[unionBranchID]*UnionBranchDeclaration
	}

	// unionBranchID identifies one generated branch alias within its union
	// declaration family.
	unionBranchID struct {
		name string
	}

	// derivedTypeKind distinguishes the only two view declaration families
	// rebuilt independently during planning and rendering.
	derivedTypeKind uint

	// derivedTypeDeclaration retains the preferred name until package freeze.
	derivedTypeDeclaration struct {
		declaration *TypeDeclaration
		name        string
		order       derivedTypeOrder
	}

	// derivedTypeOrder contains only stable semantic values so view declaration
	// suffixes never depend on expression pointer addresses or traversal order.
	derivedTypeOrder struct {
		kind        derivedTypeKind
		name        string
		sourceName  string
		sourceID    string
		sourceShape string
	}
)

const (
	projectedTypeKind derivedTypeKind = iota + 1
	viewedResultTypeKind
)

// NewProjectedTypeID returns the generated declaration identity for the
// pointer-backed projection of source emitted in a service views package.
func NewProjectedTypeID(source expr.UserType) DerivedTypeID {
	return DerivedTypeID{origin: source.Origin(), kind: projectedTypeKind}
}

// NewViewedResultTypeID returns the generated declaration identity for the
// viewed-result wrapper of source emitted in a service views package.
func NewViewedResultTypeID(source expr.UserType) DerivedTypeID {
	return DerivedTypeID{origin: source.Origin(), kind: viewedResultTypeKind}
}

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
	origin := userType.Origin()
	if declaration, ok := p.userTypes[origin]; ok {
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
	p.userTypes[origin] = declaration
	p.typeBindings[origin] = declaration
	p.userTypeNames[name] = userType.Name()
	return declaration, nil
}

// DeclareDerivedType records one generated view declaration. Rebuilding the
// projected expression from a copy of the same source origin returns the same
// declaration record.
func (p *GeneratedPackage) DeclareDerivedType(identity DerivedTypeID, name string) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	if planned, ok := p.derivedTypes[identity]; ok {
		if planned.name != name {
			return nil, fmt.Errorf(
				"derived type from %q cannot declare both %q and %q in generated package %q",
				identity.origin.Name(),
				planned.name,
				name,
				p.path,
			)
		}
		return planned.declaration, nil
	}
	order := newDerivedTypeOrder(identity, name)
	if existing, ok := p.derivedKeys[order]; ok && existing != identity {
		return nil, fmt.Errorf(
			"generated package %q cannot deterministically order derived type %q from %q",
			p.path,
			name,
			identity.origin.Name(),
		)
	}
	declaration := &TypeDeclaration{PackagePath: p.path}
	p.derivedTypes[identity] = &derivedTypeDeclaration{
		declaration: declaration,
		name:        name,
		order:       order,
	}
	p.derivedKeys[order] = identity
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
	branches := make(map[unionBranchID]*UnionBranchDeclaration, len(union.Values))
	for _, branch := range union.Values {
		identity := unionBranchID{name: branch.Name}
		if _, ok := branches[identity]; ok {
			return nil, fmt.Errorf("union %q declares branch %q more than once", union.Name(), branch.Name)
		}
		branches[identity] = &UnionBranchDeclaration{}
	}
	p.unions[identity] = &unionDeclaration{
		union:       union,
		declaration: declaration,
		branches:    branches,
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
	branch, ok := planned.branches[identity]
	if !ok {
		return nil, fmt.Errorf("branch %q of union %q is not declared in generated package %q", branchName, union.Name(), p.path)
	}
	if branch.Type != nil {
		name := Goify(userType.Name(), true)
		if branch.typeName != name {
			return nil, fmt.Errorf(
				"branch %q of union %q cannot declare both %q and %q",
				branchName,
				union.Name(),
				branch.typeName,
				name,
			)
		}
		origin := userType.Origin()
		if existing, ok := p.typeBindings[origin]; ok && existing != branch.Type {
			return nil, fmt.Errorf("user type %q is already bound to another declaration in generated package %q", userType.Name(), p.path)
		}
		p.typeBindings[origin] = branch.Type
		return branch.Type, nil
	}
	declaration := &TypeDeclaration{PackagePath: p.path}
	origin := userType.Origin()
	if existing, ok := p.typeBindings[origin]; ok && existing != declaration {
		return nil, fmt.Errorf("user type %q is already bound to another declaration in generated package %q", userType.Name(), p.path)
	}
	branch.Type = declaration
	branch.typeName = Goify(userType.Name(), true)
	p.typeBindings[origin] = declaration
	return declaration, nil
}

// UserType returns userType's existing package declaration without allocating
// a name or declaration record.
func (p *GeneratedPackage) UserType(userType expr.UserType) (*TypeDeclaration, error) {
	if declaration, ok := p.userTypes[userType.Origin()]; ok {
		return declaration, nil
	}
	return nil, fmt.Errorf("user type %q is not declared in generated package %q", userType.Name(), p.path)
}

// Type returns the frozen exact or generated branch declaration bound to
// userType's origin in this package.
func (p *GeneratedPackage) Type(userType expr.UserType) (*TypeDeclaration, error) {
	if declaration, ok := p.typeBindings[userType.Origin()]; ok {
		return declaration, nil
	}
	return nil, fmt.Errorf("user type %q has no declaration in generated package %q", userType.Name(), p.path)
}

// DerivedType returns a previously planned generated view declaration.
func (p *GeneratedPackage) DerivedType(identity DerivedTypeID) (*TypeDeclaration, error) {
	if planned, ok := p.derivedTypes[identity]; ok {
		return planned.declaration, nil
	}
	return nil, fmt.Errorf(
		"derived type from %q is not declared in generated package %q",
		identity.origin.Name(),
		p.path,
	)
}

// UnionBranch returns the existing declaration family for one union branch
// without allocating package names.
func (p *GeneratedPackage) UnionBranch(union *expr.Union, branchName string) (*UnionBranchDeclaration, error) {
	planned, ok := p.unions[NewUnionTypeID(union)]
	if !ok {
		return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
	}
	branch, ok := planned.branches[unionBranchID{name: branchName}]
	if !ok {
		return nil, fmt.Errorf("branch %q of union %q is not declared in generated package %q", branchName, union.Name(), p.path)
	}
	return branch, nil
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
func (p *GeneratedPackage) UnionBranchType(union *expr.Union, branchName string) (*TypeDeclaration, error) {
	branch, err := p.UnionBranch(union, branchName)
	if err != nil {
		return nil, err
	}
	if branch.Type == nil {
		return nil, fmt.Errorf("branch %q of union %q has no generated type in package %q", branchName, union.Name(), p.path)
	}
	return branch.Type, nil
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
		typeBindings:  make(map[expr.UserType]*TypeDeclaration),
		derivedTypes:  make(map[DerivedTypeID]*derivedTypeDeclaration),
		unions:        make(map[UnionTypeID]*unionDeclaration),
		userTypeNames: make(map[string]string),
		derivedKeys:   make(map[derivedTypeOrder]DerivedTypeID),
	}
}

// freeze assigns derived declarations in stable source order and union
// families in structural-identity order, then ends declaration and scope
// mutation while preserving read-only lookups.
func (p *GeneratedPackage) freeze() {
	derived := make([]*derivedTypeDeclaration, 0, len(p.derivedTypes))
	for _, planned := range p.derivedTypes {
		derived = append(derived, planned)
	}
	slices.SortFunc(derived, func(a, b *derivedTypeDeclaration) int {
		return compareDerivedTypeOrder(a.order, b.order)
	})
	for _, planned := range derived {
		planned.declaration.Name = p.scope.Unique(planned.name)
	}

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
			if branch.Type != nil {
				branch.Type.Name = p.scope.Unique(branch.typeName)
			}
			branch.KindConst = p.scope.Unique(planned.declaration.KindName + Goify(identity.name, true))
			branch.Constructor = p.scope.Unique("New" + planned.declaration.Name + Goify(identity.name, true))
		}
	}
	p.scope.Freeze()
	p.frozen = true
}

// newDerivedTypeOrder builds deterministic ordering data independent of
// expression pointer addresses.
func newDerivedTypeOrder(identity DerivedTypeID, name string) derivedTypeOrder {
	return derivedTypeOrder{
		kind:        identity.kind,
		name:        name,
		sourceName:  identity.origin.Name(),
		sourceID:    identity.origin.ID(),
		sourceShape: expr.Hash(identity.origin, false, false, false),
	}
}

// compareDerivedTypeOrder orders view declarations by stable typed fields.
func compareDerivedTypeOrder(left, right derivedTypeOrder) int {
	if left.kind != right.kind {
		return int(left.kind) - int(right.kind)
	}
	for _, values := range [][2]string{
		{left.name, right.name},
		{left.sourceName, right.sourceName},
		{left.sourceID, right.sourceID},
		{left.sourceShape, right.sourceShape},
	} {
		if compared := strings.Compare(values[0], values[1]); compared != 0 {
			return compared
		}
	}
	return 0
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
