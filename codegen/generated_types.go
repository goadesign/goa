// This file defines the declaration catalog owned by one generated Go
// package. Planning reserves every package-level name here before generators
// use the frozen records to render declarations and references.
package codegen

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// GeneratedPackage owns declarations and their shared naming scope for one
	// generated Go package.
	GeneratedPackage struct {
		claim        string
		path         string
		outputDir    string
		scope        *NameScope
		names        []*NameDeclaration
		exactNames   map[string]*NameDeclaration
		userTypes    map[expr.UserType]*TypeDeclaration
		typeBindings map[expr.UserType]*TypeDeclaration
		derivedTypes map[DerivedTypeID]*TypeDeclaration
		unions       map[UnionTypeID]*unionDeclaration
		frozen       bool
	}

	// DerivedTypeID identifies a generated declaration by the exact source
	// declaration and the closed transformation that produces it.
	DerivedTypeID struct {
		origin expr.UserType
		kind   derivedTypeKind
	}

	// MethodTypeIdentity identifies one closed normalized service method role.
	// It supplies both the semantic expression UID and the compiler declaration
	// kind used for the wrapper created from a raw object.
	MethodTypeIdentity struct {
		name            string
		kind            derivedTypeKind
		exampleIdentity expr.ExampleIdentity
		origin          expr.UserType
	}

	// TypeDeclaration records the canonical name and package path of one
	// generated type declaration.
	TypeDeclaration struct {
		declaration *NameDeclaration
	}

	// UnionDeclaration records the canonical union and discriminator names in
	// the package that emits them.
	UnionDeclaration struct {
		declaration     *NameDeclaration
		kindDeclaration *NameDeclaration
	}

	// UnionBranchDeclaration records the package-level declarations emitted for
	// one union branch.
	UnionBranchDeclaration struct {
		kindDeclaration        *NameDeclaration
		constructorDeclaration *NameDeclaration
		branchType             *TypeDeclaration
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

	// derivedTypeOrder contains only stable semantic values so view declaration
	// suffixes never depend on expression pointer addresses or traversal order.
	derivedTypeOrder struct {
		kind       derivedTypeKind
		name       string
		sourceName string
		sourceID   string
	}

	// unionNameOrder identifies one declaration in a generated union family.
	unionNameOrder struct {
		union  UnionTypeID
		role   unionNameRole
		branch string
	}

	// unionNameRole orders the closed package-level symbols emitted for unions.
	unionNameRole uint8
)

const (
	projectedTypeKind derivedTypeKind = iota + 1
	viewedResultTypeKind
	methodPayloadTypeKind
	methodStreamingPayloadTypeKind
	methodResultTypeKind
	methodStreamingResultTypeKind
)

const (
	unionTypeNameRole unionNameRole = iota + 1
	unionKindNameRole
	unionBranchTypeNameRole
	unionBranchKindNameRole
	unionBranchConstructorNameRole
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

// Name returns the semantic wrapper name assigned during normalization.
func (i MethodTypeIdentity) Name() string {
	return i.name
}

// UID returns the stable semantic expression identifier assigned during
// normalization. It reuses the wrapper's typed example owner so declaration
// identity and example identity cannot disagree about the method role.
func (i MethodTypeIdentity) UID() string {
	return "generated:" + i.exampleIdentity.Seed()
}

// Name returns the unqualified Go declaration name. It panics until the
// generation freezes declarations whose names depend on package collisions.
func (d *TypeDeclaration) Name() string {
	return d.declaration.Name()
}

// PackagePath returns the import path of the package that owns the declaration.
func (d *TypeDeclaration) PackagePath() string {
	return d.declaration.packagePath()
}

// Declaration returns the canonical package-owned name record.
func (d *TypeDeclaration) Declaration() *NameDeclaration {
	return d.declaration
}

// Name returns the unqualified Go union declaration name. It panics until the
// generation freezes the owning package.
func (d *UnionDeclaration) Name() string {
	return d.declaration.Name()
}

// KindName returns the unqualified Go discriminator type name. It panics until
// the generation freezes the owning package.
func (d *UnionDeclaration) KindName() string {
	return d.kindDeclaration.Name()
}

// PackagePath returns the import path of the package that owns the union.
func (d *UnionDeclaration) PackagePath() string {
	return d.declaration.packagePath()
}

// Declaration returns the canonical package-owned union type name.
func (d *UnionDeclaration) Declaration() *NameDeclaration {
	return d.declaration
}

// KindDeclaration returns the canonical package-owned discriminator type name.
func (d *UnionDeclaration) KindDeclaration() *NameDeclaration {
	return d.kindDeclaration
}

// KindConst returns the unqualified discriminator constant for the branch.
func (d *UnionBranchDeclaration) KindConst() string {
	return d.kindDeclaration.Name()
}

// Constructor returns the unqualified constructor function for the branch.
func (d *UnionBranchDeclaration) Constructor() string {
	return d.constructorDeclaration.Name()
}

// KindDeclaration returns the canonical discriminator constant name.
func (d *UnionBranchDeclaration) KindDeclaration() *NameDeclaration {
	return d.kindDeclaration
}

// ConstructorDeclaration returns the canonical branch constructor name.
func (d *UnionBranchDeclaration) ConstructorDeclaration() *NameDeclaration {
	return d.constructorDeclaration
}

// Type returns the generated branch alias declaration and whether the branch
// emits one.
func (d *UnionBranchDeclaration) Type() (*TypeDeclaration, bool) {
	return d.branchType, d.branchType != nil
}

// Ref returns the Go reference spelling for declaration's data type, including
// Goa's pointer/value semantics for named objects, unions, and aliases.
func (d *TypeDeclaration) Ref(dataType expr.DataType) string {
	return goTypeRef(d.Name(), dataType)
}

// DeclareName registers one canonical package-level declaration. Registering
// the same record again is idempotent; another owner or ambiguous order fails.
func (p *GeneratedPackage) DeclareName(declaration *NameDeclaration) error {
	if p.frozen {
		return fmt.Errorf("generated package %q is frozen", p.path)
	}
	if err := validateNameDeclaration(declaration); err != nil {
		return err
	}
	if declaration.owner != nil {
		if declaration.owner == p {
			return nil
		}
		return fmt.Errorf(
			"package name %q already belongs to generated package %q",
			declaration.preferredName(),
			declaration.owner.path,
		)
	}
	if declaration.base != nil {
		switch {
		case declaration.base.owner == nil:
			return fmt.Errorf(
				"generated package %q cannot declare preferred %s %q: base declaration is not owned",
				p.path,
				declaration.kind,
				declaration.preferredName(),
			)
		case declaration.base.owner != p:
			return fmt.Errorf(
				"generated package %q cannot declare preferred %s %q: base declaration belongs to generated package %q",
				p.path,
				declaration.kind,
				declaration.preferredName(),
				declaration.base.owner.path,
			)
		}
	}
	if declaration.exact {
		if existing, ok := p.exactNames[declaration.preferred]; ok {
			return fmt.Errorf(
				"generated package %q cannot declare exact %s %q: already declared by exact %s",
				p.path,
				declaration.kind,
				declaration.preferred,
				existing.kind,
			)
		}
		p.exactNames[declaration.preferred] = declaration
	} else {
		if err := validatePackageNameOrder(declaration.order); err != nil {
			return fmt.Errorf(
				"generated package %q cannot declare preferred %s %q: %w",
				p.path,
				declaration.kind,
				declaration.preferredName(),
				err,
			)
		}
		for _, existing := range p.names {
			if existing.exact || reflect.TypeOf(existing.order) != reflect.TypeOf(declaration.order) {
				continue
			}
			if existing.order.ComparePackageName(declaration.order) == 0 {
				return fmt.Errorf(
					"generated package %q cannot deterministically order preferred %s %q",
					p.path,
					declaration.kind,
					declaration.preferredName(),
				)
			}
		}
	}
	declaration.owner = p
	p.names = append(p.names, declaration)
	return nil
}

// DeclareUserType reserves userType's exact exported Go name and returns its
// canonical package declaration. Repeated calls return the same declaration.
func (p *GeneratedPackage) DeclareUserType(userType expr.UserType) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	origin := userType.Origin()
	name := Goify(userType.Name(), true)
	if declaration, ok := p.userTypes[origin]; ok {
		if declaration.declaration.preferred != name {
			return nil, fmt.Errorf(
				"user type origin %q cannot declare both %q and %q in generated package %q",
				origin.Name(),
				declaration.declaration.preferred,
				name,
				p.path,
			)
		}
		return declaration, nil
	}

	nameDeclaration := NewExactName(NameType, name)
	if err := p.DeclareName(nameDeclaration); err != nil {
		return nil, fmt.Errorf("declare user type %q: %w", userType.Name(), err)
	}
	declaration := &TypeDeclaration{declaration: nameDeclaration}
	if err := p.bindType(origin, declaration); err != nil {
		return nil, err
	}
	p.bindName(nameDeclaration, userType)
	p.userTypes[origin] = declaration
	return declaration, nil
}

// DeclareDerivedType records one declaration produced by a closed compiler
// transformation. Rebuilding it from the same source origin returns the same
// canonical declaration record.
func (p *GeneratedPackage) DeclareDerivedType(identity DerivedTypeID, name string) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	canonicalName := Goify(name, true)
	if declaration, ok := p.derivedTypes[identity]; ok {
		if declaration.declaration.preferred != canonicalName {
			return nil, fmt.Errorf(
				"derived type from %q cannot declare both %q and %q in generated package %q",
				identity.origin.Name(),
				declaration.declaration.preferred,
				canonicalName,
				p.path,
			)
		}
		return declaration, nil
	}
	order := newDerivedTypeOrder(identity, canonicalName)
	nameDeclaration := NewPreferredName(NameType, canonicalName, order)
	if err := p.DeclareName(nameDeclaration); err != nil {
		return nil, err
	}
	declaration := &TypeDeclaration{declaration: nameDeclaration}
	if identity.kind.isMethodType() {
		if err := p.bindType(identity.origin, declaration); err != nil {
			return nil, err
		}
	}
	p.derivedTypes[identity] = declaration
	return declaration, nil
}

// DeclareMethodType records the declaration created for identity from source
// and returns the derived identity used for later lookup.
func (p *GeneratedPackage) DeclareMethodType(identity MethodTypeIdentity, source expr.UserType) (*TypeDeclaration, DerivedTypeID, error) {
	if identity.origin == nil || identity.origin != source.Origin() {
		return nil, DerivedTypeID{}, fmt.Errorf(
			"user type %q is not the compiler-owned method wrapper %q",
			source.Name(),
			identity.UID(),
		)
	}
	derived := newDerivedTypeID(source, identity.kind)
	declaration, err := p.DeclareDerivedType(derived, identity.Name())
	return declaration, derived, err
}

// DeclareUnion records union's emitted definition and returns the same
// declaration for unions with the same emitted identity. Reading its name
// panics until the owning generation freezes its package catalogs.
func (p *GeneratedPackage) DeclareUnion(union *expr.Union) (*UnionDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	identity := NewUnionTypeID(union)
	if planned, ok := p.unions[identity]; ok {
		return planned.declaration, nil
	}

	nameDeclaration := NewPreferredName(NameType, union.Name(), unionNameOrder{
		union: identity,
		role:  unionTypeNameRole,
	})
	kindDeclaration := newDependentName(NameType, nameDeclaration, "", "Kind", unionNameOrder{
		union: identity,
		role:  unionKindNameRole,
	})
	if err := p.DeclareName(nameDeclaration); err != nil {
		return nil, err
	}
	if err := p.DeclareName(kindDeclaration); err != nil {
		return nil, err
	}
	declaration := &UnionDeclaration{
		declaration:     nameDeclaration,
		kindDeclaration: kindDeclaration,
	}
	p.bindName(nameDeclaration, identity)
	branches := make(map[unionBranchID]*UnionBranchDeclaration, len(union.Values))
	for _, branch := range union.Values {
		branchIdentity := unionBranchID{name: branch.Name}
		if _, ok := branches[branchIdentity]; ok {
			return nil, fmt.Errorf("union %q declares branch %q more than once", union.Name(), branch.Name)
		}
		kindDeclaration := newDependentName(
			NameConstant,
			declaration.kindDeclaration,
			"",
			Goify(branch.Name, true),
			unionNameOrder{union: identity, role: unionBranchKindNameRole, branch: branch.Name},
		)
		constructorDeclaration := newDependentName(
			NameFunction,
			declaration.declaration,
			"New",
			Goify(branch.Name, true),
			unionNameOrder{union: identity, role: unionBranchConstructorNameRole, branch: branch.Name},
		)
		if err := p.DeclareName(kindDeclaration); err != nil {
			return nil, err
		}
		if err := p.DeclareName(constructorDeclaration); err != nil {
			return nil, err
		}
		branches[branchIdentity] = &UnionBranchDeclaration{
			kindDeclaration:        kindDeclaration,
			constructorDeclaration: constructorDeclaration,
		}
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
		if branch.branchType.declaration.preferred != name {
			return nil, fmt.Errorf(
				"branch %q of union %q cannot declare both %q and %q",
				branchName,
				union.Name(),
				branch.branchType.declaration.preferred,
				name,
			)
		}
		if err := p.bindType(userType.Origin(), branch.branchType); err != nil {
			return nil, err
		}
		return branch.branchType, nil
	}
	typeName := Goify(userType.Name(), true)
	nameDeclaration := NewPreferredName(NameType, typeName, unionNameOrder{
		union:  NewUnionTypeID(union),
		role:   unionBranchTypeNameRole,
		branch: branchName,
	})
	if err := p.DeclareName(nameDeclaration); err != nil {
		return nil, err
	}
	declaration := &TypeDeclaration{declaration: nameDeclaration}
	origin := userType.Origin()
	if err := p.bindType(origin, declaration); err != nil {
		return nil, err
	}
	branch.branchType = declaration
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
	if declaration, ok := p.derivedTypes[identity]; ok {
		return declaration, nil
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

// ComparePackageName orders two derived service declaration identities.
func (o derivedTypeOrder) ComparePackageName(other PackageNameOrder) int {
	return compareDerivedTypeOrder(o, other.(derivedTypeOrder))
}

// ComparePackageName orders union declarations by emitted identity and role.
func (o unionNameOrder) ComparePackageName(other PackageNameOrder) int {
	right := other.(unionNameOrder)
	if compared := strings.Compare(string(o.union), string(right.union)); compared != 0 {
		return compared
	}
	if o.role != right.role {
		return int(o.role) - int(right.role)
	}
	return strings.Compare(o.branch, right.branch)
}

// newGeneratedPackage creates an empty mutable declaration catalog for path.
func newGeneratedPackage(claim, path, outputDir string) *GeneratedPackage {
	return &GeneratedPackage{
		claim:        claim,
		path:         path,
		outputDir:    outputDir,
		scope:        NewNameScope(),
		exactNames:   make(map[string]*NameDeclaration),
		userTypes:    make(map[expr.UserType]*TypeDeclaration),
		typeBindings: make(map[expr.UserType]*TypeDeclaration),
		derivedTypes: make(map[DerivedTypeID]*TypeDeclaration),
		unions:       make(map[UnionTypeID]*unionDeclaration),
	}
}

// freeze allocates exact names first, then independent preferred names in
// stable typed order, followed by names derived from an already frozen base.
func (p *GeneratedPackage) freeze() error {
	exact := make([]*NameDeclaration, 0, len(p.names))
	preferred := make([]*NameDeclaration, 0, len(p.names))
	dependent := make([]*NameDeclaration, 0, len(p.names))
	for _, declaration := range p.names {
		switch {
		case declaration.exact:
			exact = append(exact, declaration)
		case declaration.base == nil:
			preferred = append(preferred, declaration)
		default:
			dependent = append(dependent, declaration)
		}
	}
	slices.SortFunc(exact, func(left, right *NameDeclaration) int {
		return strings.Compare(left.preferred, right.preferred)
	})
	for _, declaration := range exact {
		declaration.final = p.scope.Unique(declaration.preferred)
		if declaration.final != declaration.preferred {
			return fmt.Errorf(
				"generated package %q cannot preserve exact %s name %q",
				p.path,
				declaration.kind,
				declaration.preferred,
			)
		}
		declaration.frozen = true
	}
	slices.SortFunc(preferred, comparePackageNames)
	for _, declaration := range preferred {
		declaration.final = p.scope.Unique(declaration.preferred)
		declaration.frozen = true
	}
	for len(dependent) > 0 {
		ready := dependent[:0]
		waiting := make([]*NameDeclaration, 0, len(dependent))
		for _, declaration := range dependent {
			if declaration.base.frozen {
				ready = append(ready, declaration)
			} else {
				waiting = append(waiting, declaration)
			}
		}
		if len(ready) == 0 {
			return fmt.Errorf("generated package %q contains a package-name dependency cycle", p.path)
		}
		slices.SortFunc(ready, comparePackageNames)
		for _, declaration := range ready {
			declaration.final = p.scope.Unique(declaration.preferredName())
			declaration.frozen = true
		}
		dependent = waiting
	}
	for _, declaration := range p.names {
		for _, hash := range declaration.hashes {
			p.scope.bind(hash, declaration.final)
		}
	}
	p.scope.Freeze()
	p.frozen = true
	return nil
}

// bindName associates a type identity with a canonical declaration after all
// package names have been allocated.
func (p *GeneratedPackage) bindName(declaration *NameDeclaration, hash Hasher) {
	declaration.hashes = append(declaration.hashes, hash)
}

// bindType gives one exact expression origin one canonical package
// declaration. Repeating the same binding is harmless; claiming the origin for
// another record is a planning error.
func (p *GeneratedPackage) bindType(origin expr.UserType, declaration *TypeDeclaration) error {
	if existing, ok := p.typeBindings[origin]; ok {
		if existing == declaration {
			return nil
		}
		return fmt.Errorf(
			"user type %q is already bound to another declaration in generated package %q",
			origin.Name(),
			p.path,
		)
	}
	p.typeBindings[origin] = declaration
	return nil
}

// newDerivedTypeID validates and records the exact declaration origin used by
// independently rebuilt planning and rendering graphs.
func newDerivedTypeID(source expr.UserType, kind derivedTypeKind) DerivedTypeID {
	if source == nil || source.Origin() == nil {
		panic("derived type source has no declaration origin")
	}
	return DerivedTypeID{origin: source.Origin(), kind: kind}
}

// newMethodTypeIdentity records the declaration role and exact example owner
// of one wrapper created from a raw method object.
func newMethodTypeIdentity(methodName string, kind derivedTypeKind, exampleIdentity expr.ExampleIdentity) MethodTypeIdentity {
	if !kind.isMethodType() {
		panic("method type identity requires a method role")
	}
	return MethodTypeIdentity{
		name:            Goify(methodName, true) + kind.methodSuffix(),
		kind:            kind,
		exampleIdentity: exampleIdentity,
	}
}

// bind records the exact wrapper created during normalization. The pointer is
// provenance for this run and does not participate in stable names or IDs.
func (i MethodTypeIdentity) bind(source expr.UserType) MethodTypeIdentity {
	i.origin = source.Origin()
	return i
}

// methodSuffix returns the semantic suffix for one closed method wrapper kind.
func (k derivedTypeKind) methodSuffix() string {
	switch k {
	case methodPayloadTypeKind:
		return "Payload"
	case methodStreamingPayloadTypeKind:
		return "StreamingPayload"
	case methodResultTypeKind:
		return "Result"
	case methodStreamingResultTypeKind:
		return "StreamingResult"
	default:
		panic("derived type kind is not a method role")
	}
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
