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

	// DerivedTypeID identifies a generated declaration by the exact source
	// declaration and the closed transformation that produces it.
	DerivedTypeID struct {
		origin expr.UserType
		kind   derivedTypeKind
	}

	// TypeDeclaration records the canonical name and package path of one
	// generated type declaration.
	TypeDeclaration struct {
		name        string
		packagePath string
	}

	// UnionDeclaration records the canonical union and discriminator names in
	// the package that emits them.
	UnionDeclaration struct {
		name        string
		kindName    string
		packagePath string
	}

	// UnionBranchDeclaration records the package-level declarations emitted for
	// one union branch.
	UnionBranchDeclaration struct {
		kindConst   string
		constructor string
		branchType  *TypeDeclaration
		typeName    string
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

	// derivedTypeKind distinguishes the closed declaration families rebuilt
	// independently during planning and rendering.
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
		kind       derivedTypeKind
		name       string
		sourceName string
		sourceID   string
	}
)

const (
	projectedTypeKind derivedTypeKind = iota + 1
	viewedResultTypeKind
	methodPayloadTypeKind
	methodStreamingPayloadTypeKind
	methodResultTypeKind
	methodStreamingResultTypeKind
)

// NewProjectedTypeID returns the generated declaration identity for the
// pointer-backed projection of source emitted in a service views package.
func NewProjectedTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, projectedTypeKind)
}

// NewViewedResultTypeID returns the generated declaration identity for the
// viewed-result wrapper of source emitted in a service views package.
func NewViewedResultTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, viewedResultTypeKind)
}

// NewMethodPayloadTypeID returns the generated declaration identity for a raw
// object wrapped as a service method payload.
func NewMethodPayloadTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, methodPayloadTypeKind)
}

// NewMethodStreamingPayloadTypeID returns the generated declaration identity
// for a raw object wrapped as a service method streaming payload.
func NewMethodStreamingPayloadTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, methodStreamingPayloadTypeKind)
}

// NewMethodResultTypeID returns the generated declaration identity for a raw
// object wrapped as a service method result.
func NewMethodResultTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, methodResultTypeKind)
}

// NewMethodStreamingResultTypeID returns the generated declaration identity
// for a raw object wrapped as a service method streaming result.
func NewMethodStreamingResultTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, methodStreamingResultTypeKind)
}

// Name returns the unqualified Go declaration name. It is empty until the
// generation freezes declarations whose names depend on package collisions.
func (d *TypeDeclaration) Name() string {
	return d.name
}

// PackagePath returns the import path of the package that owns the declaration.
func (d *TypeDeclaration) PackagePath() string {
	return d.packagePath
}

// Name returns the unqualified Go union declaration name. It is empty until
// the generation freezes the owning package.
func (d *UnionDeclaration) Name() string {
	return d.name
}

// KindName returns the unqualified Go discriminator type name. It is empty
// until the generation freezes the owning package.
func (d *UnionDeclaration) KindName() string {
	return d.kindName
}

// PackagePath returns the import path of the package that owns the union.
func (d *UnionDeclaration) PackagePath() string {
	return d.packagePath
}

// KindConst returns the unqualified discriminator constant for the branch.
func (d *UnionBranchDeclaration) KindConst() string {
	return d.kindConst
}

// Constructor returns the unqualified constructor function for the branch.
func (d *UnionBranchDeclaration) Constructor() string {
	return d.constructor
}

// Type returns the generated branch alias declaration and whether the branch
// emits one.
func (d *UnionBranchDeclaration) Type() (*TypeDeclaration, bool) {
	return d.branchType, d.branchType != nil
}

// Ref returns the Go reference spelling for declaration's data type, including
// Goa's pointer/value semantics for named objects, unions, and aliases.
func (d *TypeDeclaration) Ref(dataType expr.DataType) string {
	return goTypeRef(d.name, dataType)
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
	declaration := &TypeDeclaration{name: name, packagePath: p.path}
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
	declaration := &TypeDeclaration{packagePath: p.path}
	if identity.kind.isMethodType() {
		if _, ok := p.typeBindings[identity.origin]; ok {
			return nil, fmt.Errorf(
				"user type %q is already bound to another declaration in generated package %q",
				identity.origin.Name(),
				p.path,
			)
		}
	}
	p.derivedTypes[identity] = &derivedTypeDeclaration{
		declaration: declaration,
		name:        name,
		order:       order,
	}
	p.derivedKeys[order] = identity
	if identity.kind.isMethodType() {
		p.typeBindings[identity.origin] = declaration
	}
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

	declaration := &UnionDeclaration{packagePath: p.path}
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
	if branch.branchType != nil {
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
		if existing, ok := p.typeBindings[origin]; ok && existing != branch.branchType {
			return nil, fmt.Errorf("user type %q is already bound to another declaration in generated package %q", userType.Name(), p.path)
		}
		p.typeBindings[origin] = branch.branchType
		return branch.branchType, nil
	}
	declaration := &TypeDeclaration{packagePath: p.path}
	origin := userType.Origin()
	if existing, ok := p.typeBindings[origin]; ok && existing != declaration {
		return nil, fmt.Errorf("user type %q is already bound to another declaration in generated package %q", userType.Name(), p.path)
	}
	branch.branchType = declaration
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
	if branch.branchType == nil {
		return nil, fmt.Errorf("branch %q of union %q has no generated type in package %q", branchName, union.Name(), p.path)
	}
	return branch.branchType, nil
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
		planned.declaration.name = p.scope.Unique(planned.name)
	}

	identities := make([]UnionTypeID, 0, len(p.unions))
	for identity := range p.unions {
		identities = append(identities, identity)
	}
	slices.Sort(identities)
	for _, identity := range identities {
		planned := p.unions[identity]
		name := p.scope.HashedUnique(identity, Goify(planned.union.Name(), true), "")
		planned.declaration.name = name
		planned.declaration.kindName = p.scope.Unique(name + "Kind")

		branches := make([]unionBranchID, 0, len(planned.branches))
		for branch := range planned.branches {
			branches = append(branches, branch)
		}
		slices.SortFunc(branches, func(a, b unionBranchID) int {
			return strings.Compare(a.name, b.name)
		})
		for _, identity := range branches {
			branch := planned.branches[identity]
			if branch.branchType != nil {
				branch.branchType.name = p.scope.Unique(branch.typeName)
			}
			branch.kindConst = p.scope.Unique(planned.declaration.kindName + Goify(identity.name, true))
			branch.constructor = p.scope.Unique("New" + planned.declaration.name + Goify(identity.name, true))
		}
	}
	p.scope.Freeze()
	p.frozen = true
}

// newDerivedTypeID validates and records the exact declaration origin used by
// independently rebuilt planning and rendering graphs.
func newDerivedTypeID(source expr.UserType, kind derivedTypeKind) DerivedTypeID {
	if source == nil || source.Origin() == nil {
		panic("derived type source has no declaration origin")
	}
	return DerivedTypeID{origin: source.Origin(), kind: kind}
}

// isMethodType reports whether the derived declaration names a raw method
// object wrapper in the service package.
func (k derivedTypeKind) isMethodType() bool {
	return k >= methodPayloadTypeKind && k <= methodStreamingResultTypeKind
}

// newDerivedTypeOrder builds deterministic ordering data independent of
// expression pointer addresses.
func newDerivedTypeOrder(identity DerivedTypeID, name string) derivedTypeOrder {
	return derivedTypeOrder{
		kind:       identity.kind,
		name:       name,
		sourceName: identity.origin.Name(),
		sourceID:   identity.origin.ID(),
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
