// This file chooses every package-level Go name before source files are
// written. Generators record the names they need, then use the chosen names
// when writing code.
package codegen

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// GeneratedPackage stores declarations, final names, and the output location
	// for one generated Go package.
	GeneratedPackage struct {
		claim        string
		path         string
		outputDir    string
		scope        *NameScope
		names        []*NameDeclaration
		exactNames   map[string]*NameDeclaration
		nameBindings map[string]*NameDeclaration
		importPlan   *importAliasPlan
		imports      map[string]importAliasBinding
		userTypes    map[expr.UserType]*TypeDeclaration
		typeBindings map[expr.UserType]*TypeDeclaration
		derivedTypes map[DerivedTypeID]*TypeDeclaration
		unions       map[UnionDeclarationID]*unionDeclaration
		frozen       bool
	}

	// DerivedTypeID identifies a type Goa generated from a source type, such as a
	// view, method payload, or method result.
	DerivedTypeID struct {
		origin expr.UserType
		kind   derivedTypeKind
	}

	// MethodTypeIdentity records the API, Go wrapper name, whether it holds a
	// payload or result, the key used to repeat its examples, and the source type
	// for one service method.
	MethodTypeIdentity struct {
		api             string
		name            string
		kind            derivedTypeKind
		exampleIdentity expr.ExampleIdentity
		origin          expr.UserType
	}

	// TypeDeclaration stores the final name and package path of one generated Go
	// type.
	TypeDeclaration struct {
		declaration *NameDeclaration
	}

	// UnionDeclaration stores the generated union type name and its kind type
	// name.
	UnionDeclaration struct {
		declaration     *NameDeclaration
		kindDeclaration *NameDeclaration
	}

	// UnionBranchDeclaration stores the constant, constructor, and optional type
	// generated for one union branch.
	UnionBranchDeclaration struct {
		kindDeclaration        *NameDeclaration
		constructorDeclaration *NameDeclaration
		branchType             *TypeDeclaration
	}

	// unionDeclaration stores a union expression and the names generated for the
	// union and each branch.
	unionDeclaration struct {
		union       *expr.Union
		declaration *UnionDeclaration
		branches    map[unionBranchID]*UnionBranchDeclaration
	}

	// unionBranchID selects a generated union branch by its design name.
	unionBranchID struct {
		name string
	}

	// derivedTypeKind records whether Goa is generating a view, viewed result,
	// method payload, or method result.
	derivedTypeKind uint

	// derivedTypeOrder sorts generated view types from copied strings and numbers
	// so pointer addresses and visit order cannot change their suffixes.
	derivedTypeOrder struct {
		kind       derivedTypeKind
		name       string
		sourceName string
		sourceID   string
		api        string
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

// NewProjectedTypeID returns the key used to find the view-specific copy of
// source whose fields use pointers in the generated service views package.
func NewProjectedTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, projectedTypeKind)
}

// NewViewedResultTypeID returns the key used to find the viewed-result wrapper
// generated from source in a service views package.
func NewViewedResultTypeID(source expr.UserType) DerivedTypeID {
	return newDerivedTypeID(source, viewedResultTypeKind)
}

// Name returns the Go wrapper name assigned while preparing the method.
func (i MethodTypeIdentity) Name() string {
	return i.name
}

// UID returns the stable example key stored when the method value was prepared.
func (i MethodTypeIdentity) UID() string {
	return "generated:" + i.exampleIdentity.Seed()
}

// Name returns the Go type name without a package qualifier. It panics until
// Generation.Freeze chooses every declaration name because another
// declaration may still change this one.
func (d *TypeDeclaration) Name() string {
	return d.declaration.Name()
}

// PackagePath returns the import path of the package that declares the type.
func (d *TypeDeclaration) PackagePath() string {
	return d.declaration.packagePath()
}

// Declaration returns the NameDeclaration used for this generated type.
func (d *TypeDeclaration) Declaration() *NameDeclaration {
	return d.declaration
}

// Name returns the Go union type name without a package qualifier. It panics
// until Generation.Freeze chooses every declaration name.
func (d *UnionDeclaration) Name() string {
	return d.declaration.Name()
}

// KindName returns the Go union kind type name without a package qualifier. It
// panics until Generation.Freeze chooses every declaration name.
func (d *UnionDeclaration) KindName() string {
	return d.kindDeclaration.Name()
}

// PackagePath returns the import path of the package that declares the union.
func (d *UnionDeclaration) PackagePath() string {
	return d.declaration.packagePath()
}

// Declaration returns the NameDeclaration used for the union type.
func (d *UnionDeclaration) Declaration() *NameDeclaration {
	return d.declaration
}

// KindDeclaration returns the NameDeclaration used for the union kind type.
func (d *UnionDeclaration) KindDeclaration() *NameDeclaration {
	return d.kindDeclaration
}

// KindConst returns the branch kind constant without a package qualifier.
func (d *UnionBranchDeclaration) KindConst() string {
	return d.kindDeclaration.Name()
}

// Constructor returns the branch constructor name without a package qualifier.
func (d *UnionBranchDeclaration) Constructor() string {
	return d.constructorDeclaration.Name()
}

// KindDeclaration returns the NameDeclaration used for the branch kind
// constant.
func (d *UnionBranchDeclaration) KindDeclaration() *NameDeclaration {
	return d.kindDeclaration
}

// ConstructorDeclaration returns the NameDeclaration used for the branch
// constructor.
func (d *UnionBranchDeclaration) ConstructorDeclaration() *NameDeclaration {
	return d.constructorDeclaration
}

// Type returns the generated branch type and true when the branch has one.
func (d *UnionBranchDeclaration) Type() (*TypeDeclaration, bool) {
	return d.branchType, d.branchType != nil
}

// Ref returns the Go type reference for dataType, including the pointer chosen
// by Goa for named objects, unions, and aliases.
func (d *TypeDeclaration) Ref(dataType expr.DataType) string {
	return goTypeRef(d.Name(), dataType)
}

// DeclareName records one package-level Go name. Each supplied key will return
// that same name during type formatting. Repeating the same name and keys has
// no effect. It returns an error if the name belongs to another package, cannot
// be ordered, or a key already selects another name.
func (p *GeneratedPackage) DeclareName(declaration *NameDeclaration, keys ...Hasher) error {
	if p.frozen {
		return fmt.Errorf("generated package %q is frozen", p.path)
	}
	if err := validateNameDeclaration(declaration); err != nil {
		return err
	}
	if err := p.validateNameBindings(declaration, keys); err != nil {
		return err
	}
	if declaration.owner != nil {
		if declaration.owner == p {
			return p.recordNameBindings(declaration, keys)
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
	return p.recordNameBindings(declaration, keys)
}

// DeclareDependentName adds a generated declaration named by placing prefix and
// suffix around base's final name. base must already be declared in p.
func (p *GeneratedPackage) DeclareDependentName(kind PackageNameKind, base *NameDeclaration, prefix, suffix string, order PackageNameOrder) (*NameDeclaration, error) {
	declaration := newDependentName(kind, base, prefix, suffix, order)
	if err := p.DeclareName(declaration); err != nil {
		return nil, err
	}
	return declaration, nil
}

// DeclareGeneratedType adds a type produced by a generator plugin. Goa assigns
// a stable final name but does not associate the declaration with an authored
// Goa type.
func (p *GeneratedPackage) DeclareGeneratedType(preferredName string, order PackageNameOrder) (*TypeDeclaration, error) {
	declaration := NewPreferredName(NameType, Goify(preferredName, true), ExportedName, order)
	if err := p.DeclareName(declaration); err != nil {
		return nil, fmt.Errorf("declare generated type %q: %w", preferredName, err)
	}
	return &TypeDeclaration{declaration: declaration}, nil
}

// DeclareUserType adds userType with its exact exported Go name and returns the
// generated declaration. Repeated calls for the same source type return the
// same declaration.
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

// DeclareDerivedType adds one generated form of a source type. Repeated calls
// with the same DerivedTypeID return the same declaration.
func (p *GeneratedPackage) DeclareDerivedType(identity DerivedTypeID, name string) (*TypeDeclaration, error) {
	return p.declareDerivedType(identity, name, "")
}

// declareDerivedType adds one generated form and uses api only to order method
// wrappers contributed by different APIs to the same Go package.
func (p *GeneratedPackage) declareDerivedType(identity DerivedTypeID, name, api string) (*TypeDeclaration, error) {
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
	order := newDerivedTypeOrder(identity, canonicalName, api)
	nameDeclaration := NewPreferredName(NameType, canonicalName, ExportedName, order)
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

// DeclareMethodType adds the wrapper described by the MethodTypeIdentity value
// for source. It returns the declaration and DerivedTypeID used by later
// lookups.
func (p *GeneratedPackage) DeclareMethodType(identity MethodTypeIdentity, source expr.UserType) (*TypeDeclaration, DerivedTypeID, error) {
	if identity.origin == nil || identity.origin != source.Origin() {
		return nil, DerivedTypeID{}, fmt.Errorf(
			"user type %q is not the compiler-owned method wrapper %q",
			source.Name(),
			identity.UID(),
		)
	}
	derived := newDerivedTypeID(source, identity.kind)
	declaration, err := p.declareDerivedType(derived, identity.Name(), identity.api)
	return declaration, derived, err
}

// DeclareUnion adds the generated union type, kind type, branch constants, and
// branch constructors for the OneOf stored in attribute. Copies of the same
// authored OneOf and emitted definition return the same declaration. Reading
// generated names panics until Generation.Freeze chooses every declaration
// name.
func (p *GeneratedPackage) DeclareUnion(attribute *expr.AttributeExpr) (*UnionDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	union, ok := attribute.Type.(*expr.Union)
	if !ok {
		return nil, fmt.Errorf("generated package %q can only declare a OneOf attribute", p.path)
	}
	if err := validateUnionBranchGoNames(union); err != nil {
		return nil, err
	}
	identity := NewUnionDeclarationID(attribute)
	if planned, ok := p.unions[identity]; ok {
		return planned.declaration, nil
	}

	name := Goify(union.Name(), true)
	nameDeclaration := NewExactName(NameType, name)
	kindDeclaration := NewExactName(NameType, name+"Kind")
	if err := p.declareUnionName(union, nameDeclaration); err != nil {
		return nil, err
	}
	if err := p.declareUnionName(union, kindDeclaration); err != nil {
		return nil, err
	}
	declaration := &UnionDeclaration{
		declaration:     nameDeclaration,
		kindDeclaration: kindDeclaration,
	}
	branches := make(map[unionBranchID]*UnionBranchDeclaration, len(union.Values))
	for _, branch := range union.Values {
		branchIdentity := unionBranchID{name: branch.Name}
		if _, ok := branches[branchIdentity]; ok {
			return nil, fmt.Errorf("union %q declares branch %q more than once", union.Name(), branch.Name)
		}
		branchName := Goify(branch.Name, true)
		kindDeclaration := NewExactName(NameConstant, name+"Kind"+branchName)
		constructorDeclaration := NewExactName(NameFunction, "New"+name+branchName)
		if err := p.declareUnionName(union, kindDeclaration); err != nil {
			return nil, err
		}
		if err := p.declareUnionName(union, constructorDeclaration); err != nil {
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

// DeclareUnionBranchType adds the generated type used by branchName in the
// OneOf stored in attribute. Copies of that authored OneOf share the branch
// declaration. Types written directly in the DSL must use DeclareUserType
// instead.
func (p *GeneratedPackage) DeclareUnionBranchType(attribute *expr.AttributeExpr, branchName string, userType expr.UserType) (*TypeDeclaration, error) {
	if p.frozen {
		return nil, fmt.Errorf("generated package %q is frozen", p.path)
	}
	union, ok := attribute.Type.(*expr.Union)
	if !ok {
		return nil, fmt.Errorf("generated package %q can only use a OneOf attribute", p.path)
	}
	identity := NewUnionDeclarationID(attribute)
	planned, ok := p.unions[identity]
	if !ok {
		return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
	}
	if !unionHasBranchType(union, branchName, userType) {
		return nil, fmt.Errorf("user type %q is not branch %q of union %q", userType.Name(), branchName, union.Name())
	}

	branchIdentity := unionBranchID{name: branchName}
	branch, ok := planned.branches[branchIdentity]
	if !ok {
		return nil, fmt.Errorf("branch %q of union %q is not declared in generated package %q", branchName, union.Name(), p.path)
	}
	if branch.branchType != nil {
		name := unionBranchTypeName(union, branchName)
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
	typeName := unionBranchTypeName(union, branchName)
	nameDeclaration := NewExactName(NameType, typeName)
	if err := p.declareUnionName(union, nameDeclaration); err != nil {
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

// UserType returns the declaration previously added for userType. It does not
// add a name.
func (p *GeneratedPackage) UserType(userType expr.UserType) (*TypeDeclaration, error) {
	if declaration, ok := p.userTypes[userType.Origin()]; ok {
		return declaration, nil
	}
	return nil, fmt.Errorf("user type %q is not declared in generated package %q", userType.Name(), p.path)
}

// Type returns the exact user type declaration or generated union branch type
// previously associated with userType's source declaration.
func (p *GeneratedPackage) Type(userType expr.UserType) (*TypeDeclaration, error) {
	if declaration, ok := p.typeBindings[userType.Origin()]; ok {
		return declaration, nil
	}
	return nil, fmt.Errorf("user type %q has no declaration in generated package %q", userType.Name(), p.path)
}

// DerivedType returns the generated view type previously added for the supplied
// DerivedTypeID.
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

// UnionBranch returns the constant, constructor, and optional type previously
// added for branchName. It does not add names.
func (p *GeneratedPackage) UnionBranch(attribute *expr.AttributeExpr, branchName string) (*UnionBranchDeclaration, error) {
	union, ok := attribute.Type.(*expr.Union)
	if !ok {
		return nil, fmt.Errorf("generated package %q can only use a OneOf attribute", p.path)
	}
	planned, ok := p.unions[NewUnionDeclarationID(attribute)]
	if !ok {
		return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
	}
	branch, ok := planned.branches[unionBranchID{name: branchName}]
	if !ok {
		return nil, fmt.Errorf("branch %q of union %q is not declared in generated package %q", branchName, union.Name(), p.path)
	}
	return branch, nil
}

// Union returns the declaration previously added for union. It does not add a
// name.
func (p *GeneratedPackage) Union(attribute *expr.AttributeExpr) (*UnionDeclaration, error) {
	union, ok := attribute.Type.(*expr.Union)
	if !ok {
		return nil, fmt.Errorf("generated package %q can only use a OneOf attribute", p.path)
	}
	if planned, ok := p.unions[NewUnionDeclarationID(attribute)]; ok {
		return planned.declaration, nil
	}
	return nil, fmt.Errorf("union %q is not declared in generated package %q", union.Name(), p.path)
}

// UnionBranchType returns the generated type previously added for branchName.
// It returns an error when the branch does not generate a type.
func (p *GeneratedPackage) UnionBranchType(attribute *expr.AttributeExpr, branchName string) (*TypeDeclaration, error) {
	union, ok := attribute.Type.(*expr.Union)
	if !ok {
		return nil, fmt.Errorf("generated package %q can only use a OneOf attribute", p.path)
	}
	branch, err := p.UnionBranch(attribute, branchName)
	if err != nil {
		return nil, err
	}
	if branch.branchType == nil {
		return nil, fmt.Errorf("branch %q of union %q has no generated type in package %q", branchName, union.Name(), p.path)
	}
	return branch.branchType, nil
}

// Scope returns the package's NameScope after all names are final. It panics
// until Generation.Freeze chooses every declaration name.
func (p *GeneratedPackage) Scope() *NameScope {
	if !p.frozen {
		panic(fmt.Sprintf("generated package %q scope requested before freeze", p.path))
	}
	return p.scope
}

// ComparePackageName sorts two generated service type names.
func (o derivedTypeOrder) ComparePackageName(other PackageNameOrder) int {
	return compareDerivedTypeOrder(o, other.(derivedTypeOrder))
}

// newGeneratedPackage returns an empty package record for path and outputDir.
func newGeneratedPackage(claim, path, outputDir string) *GeneratedPackage {
	return &GeneratedPackage{
		claim:        claim,
		path:         path,
		outputDir:    outputDir,
		scope:        NewNameScope(),
		exactNames:   make(map[string]*NameDeclaration),
		nameBindings: make(map[string]*NameDeclaration),
		importPlan: &importAliasPlan{
			candidates: make(map[string]*importAliasCandidate),
		},
		userTypes:    make(map[expr.UserType]*TypeDeclaration),
		typeBindings: make(map[expr.UserType]*TypeDeclaration),
		derivedTypes: make(map[DerivedTypeID]*TypeDeclaration),
		unions:       make(map[UnionDeclarationID]*unionDeclaration),
	}
}

// freeze chooses exact names first, then names that may receive a number, and
// finally names built from another chosen name. It then rejects any attempt to
// add or change a name.
func (p *GeneratedPackage) freeze() error {
	if err := p.freezeImports(); err != nil {
		return err
	}
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
	for identity, union := range p.unions {
		p.scope.bindUnion(identity, union.declaration.declaration.final)
	}
	p.scope.Freeze()
	p.frozen = true
	return nil
}

// declareUnionName records one exact public name owned by union. A collision
// tells the design author how to choose another complete union name.
func (p *GeneratedPackage) declareUnionName(union *expr.Union, declaration *NameDeclaration) error {
	if err := p.DeclareName(declaration); err != nil {
		return fmt.Errorf(
			"declare OneOf %q in generated package %q: %w; set TypeName on the OneOf to a unique name",
			union.Name(),
			p.path,
			err,
		)
	}
	return nil
}

// validateUnionBranchGoNames rejects branch spellings that would create the
// same exported Go identifier and therefore the same public union functions.
func validateUnionBranchGoNames(union *expr.Union) error {
	branches := make(map[string]string, len(union.Values))
	for _, branch := range union.Values {
		goName := Goify(branch.Name, true)
		if existing, ok := branches[goName]; ok {
			return fmt.Errorf(
				"OneOf %q branches %q and %q both generate Go name %q; rename one of the branches",
				union.Name(),
				existing,
				branch.Name,
				goName,
			)
		}
		branches[goName] = branch.Name
	}
	return nil
}

// bindName makes lookups for hash return declaration's chosen Go name.
func (p *GeneratedPackage) bindName(declaration *NameDeclaration, hash Hasher) {
	if err := p.recordNameBindings(declaration, []Hasher{hash}); err != nil {
		panic(err)
	}
}

// validateNameBindings rejects nil lookup keys and keys that already return a
// different Go declaration name.
func (p *GeneratedPackage) validateNameBindings(declaration *NameDeclaration, keys []Hasher) error {
	for _, key := range keys {
		if key == nil {
			return fmt.Errorf("generated package %q cannot declare a nil lookup key", p.path)
		}
		hash := key.Hash()
		if existing := p.nameBindings[hash]; existing != nil && existing != declaration {
			return fmt.Errorf(
				"generated package %q lookup key %q already belongs to %s %q",
				p.path,
				hash,
				existing.kind,
				existing.preferredName(),
			)
		}
	}
	return nil
}

// recordNameBindings makes each lookup key return declaration's chosen Go
// name.
func (p *GeneratedPackage) recordNameBindings(declaration *NameDeclaration, keys []Hasher) error {
	if err := p.validateNameBindings(declaration, keys); err != nil {
		return err
	}
	for _, key := range keys {
		hash := key.Hash()
		if p.nameBindings[hash] == declaration {
			continue
		}
		p.nameBindings[hash] = declaration
		declaration.hashes = append(declaration.hashes, key)
	}
	return nil
}

// bindType associates one source type with one generated declaration. Repeating
// the same association has no effect; using a different declaration returns an
// error.
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

// newDerivedTypeID identifies one generated view, payload, or result type by
// the source type it came from. It panics when that source is unknown.
func newDerivedTypeID(source expr.UserType, kind derivedTypeKind) DerivedTypeID {
	if source == nil || source.Origin() == nil {
		panic("derived type source has no declaration origin")
	}
	return DerivedTypeID{origin: source.Origin(), kind: kind}
}

// newMethodTypeIdentity records the API, generated wrapper name, whether it
// holds a payload or result, and the key used to repeat examples for one
// service method value.
func newMethodTypeIdentity(apiName, methodName string, kind derivedTypeKind, exampleIdentity expr.ExampleIdentity) MethodTypeIdentity {
	if !kind.isMethodType() {
		panic("method type identity requires a method role")
	}
	return MethodTypeIdentity{
		api:             apiName,
		name:            Goify(methodName, true) + kind.methodSuffix(),
		kind:            kind,
		exampleIdentity: exampleIdentity,
	}
}

// bind records the source type wrapped for one method. Goa uses the pointer only
// while preparing this run; it does not change generated names or identifiers.
func (i MethodTypeIdentity) bind(source expr.UserType) MethodTypeIdentity {
	i.origin = source.Origin()
	return i
}

// methodSuffix returns the Go name suffix for a method payload or result
// wrapper.
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

// isMethodType reports whether this kind is a service method payload or result
// wrapper.
func (k derivedTypeKind) isMethodType() bool {
	return k >= methodPayloadTypeKind && k <= methodStreamingResultTypeKind
}

// newDerivedTypeOrder copies the values used to sort generated type names so
// expression pointer addresses cannot affect their suffixes.
func newDerivedTypeOrder(identity DerivedTypeID, name, api string) derivedTypeOrder {
	return derivedTypeOrder{
		kind:       identity.kind,
		name:       name,
		sourceName: identity.origin.Name(),
		sourceID:   identity.origin.ID(),
		api:        api,
	}
}

// compareDerivedTypeOrder sorts generated view types by kind, requested name,
// source name, source ID, and API name.
func compareDerivedTypeOrder(left, right derivedTypeOrder) int {
	if left.kind != right.kind {
		return int(left.kind) - int(right.kind)
	}
	for _, values := range [][2]string{
		{left.name, right.name},
		{left.sourceName, right.sourceName},
		{left.sourceID, right.sourceID},
		{left.api, right.api},
	} {
		if compared := strings.Compare(values[0], values[1]); compared != 0 {
			return compared
		}
	}
	return 0
}

// unionHasBranchType reports whether branchName in this OneOf occurrence uses
// userType. The caller has already found the authored OneOf declaration.
func unionHasBranchType(union *expr.Union, branchName string, userType expr.UserType) bool {
	for _, branch := range union.Values {
		if branch.Name == branchName && branch.Attribute.Type == userType {
			return true
		}
	}
	return false
}

// unionBranchTypeName separates a compiler-created branch type from the union
// kind type and branch constants in the same Go package.
func unionBranchTypeName(union *expr.Union, branchName string) string {
	return Goify(union.Name(), true) + "Branch" + Goify(branchName, true)
}
