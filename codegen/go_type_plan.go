// This file retains Go type layouts and exact generated declaration bindings
// before package names freeze. Linked formatters render only these copied facts;
// they never inspect the mutable Goa expression graph.
package codegen

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"goa.design/goa/v3/expr"
)

type (
	// GoTypeKind identifies one retained Go layout category.
	GoTypeKind uint8

	// GoTypeImport identifies one package used by a retained type spelling.
	// Name is the preferred qualifier supplied by the type contract; generated
	// declaration owners leave Name empty.
	GoTypeImport struct {
		// Name is the preferred package qualifier, when the design supplied one.
		Name string
		// Path is the canonical Go import path.
		Path string
	}

	// GoTypeBindingRequest describes one named or union occurrence whose owning
	// subsystem must bind an exact generated declaration during planning.
	GoTypeBindingRequest struct {
		// Attribute is the exact expression occurrence being planned. Binders may
		// inspect it during planning; linked formatters never do.
		Attribute *expr.AttributeExpr
		// InheritedOwner is the package inherited from the enclosing layout.
		InheritedOwner string
		// Kind distinguishes named types from union declarations.
		Kind GoTypeKind
	}

	// GoTypeBinding binds one planned occurrence to its exact generated package
	// declaration. Named occurrences require Type; union occurrences require
	// Union. The other declaration field must be nil.
	GoTypeBinding struct {
		// Owner is the canonical import path that owns the declaration.
		Owner string
		// Type is the exact generated declaration for a named user type.
		Type *TypeDeclaration
		// Union is the exact generated declaration for a sum type.
		Union *UnionDeclaration
	}

	// GoTypeBinder supplies package ownership and canonical declaration records
	// for named and union occurrences. The subsystem that owns those catalogs
	// must provide this callback; core layout planning never infers ownership.
	GoTypeBinder func(GoTypeBindingRequest) (GoTypeBinding, error)

	// GoLayoutPolicy records the complete generated Go field and validation
	// representation selected for one plan. A shared named value keeps service,
	// view, and transport planners from independently reconstructing policy.
	GoLayoutPolicy struct {
		// Pointer forces primitive object fields to use pointers.
		Pointer bool
		// IgnoreRequired suppresses required checks for primitive transport fields.
		IgnoreRequired bool
		// UseDefault keeps optional primitive fields with defaults as values.
		UseDefault bool
		// UnionPointer uses pointers for optional sum-type union fields and for
		// required union fields when Pointer is also true.
		UnionPointer bool
		// SumType reports that unions use Goa's generated struct representation.
		SumType bool
	}

	// GoTypePlanOptions configures one exact layout occurrence.
	GoTypePlanOptions struct {
		// Owner is the package inherited by the root occurrence.
		Owner string
		// FieldName is the optional design field name of the root occurrence.
		FieldName string
		// Policy is the complete Go representation selected by the caller.
		Policy GoLayoutPolicy
		// Bind resolves every named type and union to an exact declaration.
		Bind GoTypeBinder
	}

	// GoTypePlan is an immutable symbolic Go layout built while expressions are
	// available and package declarations remain mutable. It retains source
	// pointers only for occurrence identity; no method reads an expression after
	// PlanGoType returns.
	GoTypePlan struct {
		kind              GoTypeKind
		owner             string
		policy            GoLayoutPolicy
		occurrence        *expr.AttributeExpr
		fieldNameUpper    string
		fieldNameLower    string
		description       string
		comment           string
		tag               string
		fieldPointer      bool
		definitionPointer bool
		referencePointer  bool
		primitive         string
		directImport      GoTypeImport
		hasDirectImport   bool
		customQualifier   string
		typeDeclaration   *TypeDeclaration
		unionDeclaration  *UnionDeclaration
		fields            []*GoTypePlan
		branches          []*GoTypePlan
		element           *GoTypePlan
		key               *GoTypePlan
	}

	// GoTypeQualifier returns the final package qualifier for one canonical
	// import path after the generation freezes its shared import aliases.
	GoTypeQualifier func(importPath string) string

	// LinkedGoType formats one retained plan relative to an output package. It
	// resolves only frozen declaration names and retained import identities.
	LinkedGoType struct {
		plan       *GoTypePlan
		outputPath string
		qualifier  GoTypeQualifier
	}

	// goTypePlanner owns the expression-reading planning phase.
	goTypePlanner struct {
		policy GoLayoutPolicy
		bind   GoTypeBinder
	}
)

const (
	// GoPrimitive is a built-in or explicitly imported primitive spelling.
	GoPrimitive GoTypeKind = iota + 1
	// GoArray is a slice layout with one retained element occurrence.
	GoArray
	// GoMap is a map layout with retained key and element occurrences.
	GoMap
	// GoStruct is an anonymous struct with retained ordered field occurrences.
	GoStruct
	// GoNamed is a user type bound to an exact generated type declaration.
	GoNamed
	// GoUnion is a sum type bound to an exact generated union declaration.
	GoUnion
	// GoEmpty is Goa's built-in empty service type.
	GoEmpty
	// GoServiceError is Goa's built-in service error type.
	GoServiceError
)

// PlanGoType copies the complete Go layout for attribute while generated
// packages are still mutable. Callers link and format the returned plan only
// after the generation freezes its declaration and import-alias catalogs.
func PlanGoType(attribute *expr.AttributeExpr, options GoTypePlanOptions) (*GoTypePlan, error) {
	if attribute == nil {
		return nil, fmt.Errorf("plan Go type: attribute must not be nil")
	}
	if options.Owner == "" {
		return nil, fmt.Errorf("plan Go type: inherited owner must not be empty")
	}
	planner := goTypePlanner{
		policy: options.Policy,
		bind:   options.Bind,
	}
	return planner.plan(attribute, options.Owner, options.FieldName, nil, false)
}

// String returns the layout category used in planning diagnostics.
func (k GoTypeKind) String() string {
	switch k {
	case GoPrimitive:
		return "primitive"
	case GoArray:
		return "array"
	case GoMap:
		return "map"
	case GoStruct:
		return "struct"
	case GoNamed:
		return "named type"
	case GoUnion:
		return "union"
	case GoEmpty:
		return "empty type"
	case GoServiceError:
		return "service error"
	default:
		return "unknown"
	}
}

// Kind returns the retained layout category.
func (p *GoTypePlan) Kind() GoTypeKind {
	return p.kind
}

// Owner returns the canonical import path inherited or selected for this
// exact occurrence.
func (p *GoTypePlan) Owner() string {
	return p.owner
}

// Policy returns the complete generated representation selected for this
// occurrence.
func (p *GoTypePlan) Policy() GoLayoutPolicy {
	return p.policy
}

// MatchesOccurrence reports whether attribute is the exact expression pointer
// used to build this plan. It never reads the expression.
func (p *GoTypePlan) MatchesOccurrence(attribute *expr.AttributeExpr) bool {
	return p.occurrence == attribute
}

// PlansForOccurrence returns every plan in this retained layout that was built
// from attribute. Separate entries preserve distinct field, owner, or pointer
// policies when one expression pointer is reused.
func (p *GoTypePlan) PlansForOccurrence(attribute *expr.AttributeExpr) []*GoTypePlan {
	var matches []*GoTypePlan
	p.walk(func(candidate *GoTypePlan) {
		if candidate.occurrence == attribute {
			matches = append(matches, candidate)
		}
	})
	return matches
}

// TypeDeclaration returns the exact named declaration retained for this
// occurrence, or nil for layouts that are not named user types.
func (p *GoTypePlan) TypeDeclaration() *TypeDeclaration {
	return p.typeDeclaration
}

// UnionDeclaration returns the exact union declaration retained for this
// occurrence, or nil for layouts that are not unions.
func (p *GoTypePlan) UnionDeclaration() *UnionDeclaration {
	return p.unionDeclaration
}

// FieldName returns the retained Go field identifier. It returns the exported
// spelling when firstUpper is true and the package-local spelling otherwise.
func (p *GoTypePlan) FieldName(firstUpper bool) string {
	if firstUpper {
		return p.fieldNameUpper
	}
	return p.fieldNameLower
}

// Description returns the copied design description for this occurrence.
func (p *GoTypePlan) Description() string {
	return p.description
}

// Tag returns the complete retained Go struct tag, including leading space.
func (p *GoTypePlan) Tag() string {
	return p.tag
}

// IsPointer reports whether an enclosing struct field stores this occurrence
// through a pointer under the planned pointer/default policy.
func (p *GoTypePlan) IsPointer() bool {
	return p.fieldPointer
}

// Import returns the package imported directly by this type spelling. The
// boolean is false for native and generated declaration spellings.
func (p *GoTypePlan) Import() (GoTypeImport, bool) {
	return p.directImport, p.hasDirectImport
}

// ImportPreferences returns every distinct authored alias preference and
// generated declaration path reachable from this plan in stable layout order.
// Multiple preferences for one path remain distinct so generation can resolve
// them before freezing its import aliases.
func (p *GoTypePlan) ImportPreferences() []GoTypeImport {
	seen := make(map[GoTypeImport]struct{})
	var imports []GoTypeImport
	p.walkImports(func(candidate *GoTypePlan) {
		var goImport GoTypeImport
		switch {
		case candidate.hasDirectImport:
			goImport = candidate.directImport
		case candidate.typeDeclaration != nil || candidate.unionDeclaration != nil:
			goImport = GoTypeImport{Path: candidate.owner}
		default:
			return
		}
		if _, exists := seen[goImport]; exists {
			return
		}
		seen[goImport] = struct{}{}
		imports = append(imports, goImport)
	})
	return imports
}

// Fields returns a copy of the ordered anonymous struct field plans.
func (p *GoTypePlan) Fields() []*GoTypePlan {
	return append([]*GoTypePlan(nil), p.fields...)
}

// Branches returns a copy of the ordered union branch plans.
func (p *GoTypePlan) Branches() []*GoTypePlan {
	return append([]*GoTypePlan(nil), p.branches...)
}

// Elem returns the retained array or map element plan, or nil for other kinds.
func (p *GoTypePlan) Elem() *GoTypePlan {
	return p.element
}

// Key returns the retained map key plan, or nil for other kinds.
func (p *GoTypePlan) Key() *GoTypePlan {
	return p.key
}

// Equivalent reports whether p and other retain the same complete Go layout.
// Source expression pointers are deliberately excluded: independently built
// compiler copies are equivalent when they bind the same declarations and
// retain identical owners, policies, field spellings, tags, pointer choices,
// imports, and ordered child layouts.
func (p *GoTypePlan) Equivalent(other *GoTypePlan) bool {
	if p == nil || other == nil {
		return p == other
	}
	if p.kind != other.kind || p.owner != other.owner || p.policy != other.policy ||
		p.fieldNameUpper != other.fieldNameUpper || p.fieldNameLower != other.fieldNameLower ||
		p.description != other.description || p.comment != other.comment || p.tag != other.tag ||
		p.fieldPointer != other.fieldPointer || p.definitionPointer != other.definitionPointer ||
		p.referencePointer != other.referencePointer || p.primitive != other.primitive ||
		p.directImport != other.directImport || p.hasDirectImport != other.hasDirectImport ||
		p.customQualifier != other.customQualifier || p.typeDeclaration != other.typeDeclaration ||
		p.unionDeclaration != other.unionDeclaration || len(p.fields) != len(other.fields) ||
		len(p.branches) != len(other.branches) {
		return false
	}
	if !p.key.Equivalent(other.key) || !p.element.Equivalent(other.element) {
		return false
	}
	for index := range p.fields {
		if !p.fields[index].Equivalent(other.fields[index]) {
			return false
		}
	}
	for index := range p.branches {
		if !p.branches[index].Equivalent(other.branches[index]) {
			return false
		}
	}
	return true
}

// Link binds this retained layout to one generated output package after the
// owning generation freezes declarations and import aliases. Link itself is a
// pure binding operation; declaration access remains governed by the catalog's
// freeze contract. The returned formatter contains no expression traversal or
// metadata decisions.
func (p *GoTypePlan) Link(outputPath string, qualifier GoTypeQualifier) LinkedGoType {
	return LinkedGoType{plan: p, outputPath: outputPath, qualifier: qualifier}
}

// Name returns the Go type spelling selected by the retained layout.
func (l LinkedGoType) Name() string {
	switch l.plan.kind {
	case GoPrimitive:
		if !l.plan.hasDirectImport || l.plan.customQualifier == "" {
			return l.plan.primitive
		}
		return strings.ReplaceAll(
			l.plan.primitive,
			l.plan.customQualifier+".",
			l.qualify(l.plan.directImport.Path)+".",
		)
	case GoArray:
		return "[]" + l.Enter(l.plan.element).Ref()
	case GoMap:
		return fmt.Sprintf(
			"map[%s]%s",
			l.Enter(l.plan.key).Ref(),
			l.Enter(l.plan.element).Ref(),
		)
	case GoStruct:
		return l.Def()
	case GoNamed:
		return l.qualifiedDeclaration(l.plan.typeDeclaration.Declaration())
	case GoUnion:
		return l.qualifiedDeclaration(l.plan.unionDeclaration.Declaration())
	case GoEmpty:
		return "struct {}"
	case GoServiceError:
		return l.qualify(l.plan.directImport.Path) + ".ServiceError"
	default:
		panic(fmt.Sprintf("format unknown retained Go type kind %d", l.plan.kind))
	}
}

// Def returns the Go definition selected by the retained layout.
func (l LinkedGoType) Def() string {
	switch l.plan.kind {
	case GoArray:
		element := l.Enter(l.plan.element).Def()
		if l.plan.element.definitionPointer {
			element = "*" + element
		}
		return "[]" + element
	case GoMap:
		key := l.Enter(l.plan.key).Def()
		if l.plan.key.definitionPointer {
			key = "*" + key
		}
		element := l.Enter(l.plan.element).Def()
		if l.plan.element.definitionPointer {
			element = "*" + element
		}
		return fmt.Sprintf("map[%s]%s", key, element)
	case GoStruct:
		lines := []string{"struct {"}
		for _, field := range l.plan.fields {
			fieldType := l.Enter(field).Def()
			if field.fieldPointer {
				fieldType = "*" + fieldType
			}
			var description string
			if field.comment != "" {
				description = field.comment + "\n\t"
			}
			lines = append(lines, fmt.Sprintf(
				"\t%s%s %s%s",
				description,
				field.fieldNameUpper,
				fieldType,
				field.tag,
			))
		}
		return strings.Join(append(lines, "}"), "\n")
	default:
		return l.Name()
	}
}

// Ref returns the retained Go reference spelling, including named object and
// union pointer semantics.
func (l LinkedGoType) Ref() string {
	name := l.Name()
	if l.plan.referencePointer {
		return "*" + name
	}
	return name
}

// Field returns the retained field identifier for this exact occurrence.
func (l LinkedGoType) Field(firstUpper bool) string {
	return l.plan.FieldName(firstUpper)
}

// Package returns the qualifier for this occurrence's owner relative to the
// linked output package, or the empty string for a same-package occurrence.
func (l LinkedGoType) Package() string {
	if l.plan.owner == l.outputPath {
		return ""
	}
	return l.qualify(l.plan.owner)
}

// Enter links an exact retained child while preserving the output package and
// frozen import alias lookup.
func (l LinkedGoType) Enter(child *GoTypePlan) LinkedGoType {
	if child == nil {
		panic("enter nil retained Go type plan")
	}
	return LinkedGoType{plan: child, outputPath: l.outputPath, qualifier: l.qualifier}
}

// Imports returns every recursively retained import except the linked output
// package itself.
func (l LinkedGoType) Imports() []GoTypeImport {
	preferences := l.plan.ImportPreferences()
	seen := make(map[string]struct{})
	imports := make([]GoTypeImport, 0, len(preferences))
	for _, preference := range preferences {
		if preference.Path == l.outputPath {
			continue
		}
		if _, exists := seen[preference.Path]; exists {
			continue
		}
		seen[preference.Path] = struct{}{}
		imports = append(imports, GoTypeImport{
			Name: l.qualify(preference.Path),
			Path: preference.Path,
		})
	}
	return imports
}

// plan copies one exact occurrence and recursively retains anonymous child
// layouts. Named types terminate at their canonical declaration binding.
func (p goTypePlanner) plan(attribute *expr.AttributeExpr, owner, fieldName string, parent *expr.AttributeExpr, definitionPointer bool) (*GoTypePlan, error) {
	layoutAttribute := attribute
	for {
		if _, named := layoutAttribute.Type.(expr.UserType); named {
			break
		}
		composite, ok := layoutAttribute.Type.(expr.CompositeExpr)
		if !ok {
			break
		}
		layoutAttribute = composite.Attribute()
	}
	plan := &GoTypePlan{
		owner:             owner,
		policy:            p.policy,
		occurrence:        attribute,
		fieldNameUpper:    GoifyAtt(attribute, fieldName, true),
		fieldNameLower:    GoifyAtt(attribute, fieldName, false),
		description:       attribute.Description,
		tag:               AttributeTagsWithName(parent, fieldName, attribute),
		definitionPointer: definitionPointer,
	}
	if attribute.Description != "" {
		plan.comment = Comment(attribute.Description)
	}
	if parent != nil {
		field := expr.AsObject(parent.Type).Attribute(fieldName)
		switch {
		case expr.IsUnion(field.Type):
			plan.fieldPointer = p.policy.UnionPointer && (!parent.IsRequired(fieldName) || p.policy.Pointer)
		case !p.policy.SumType:
			plan.fieldPointer = expr.IsPrimitive(field.Type) &&
				(p.policy.Pointer || parent.IsPrimitivePointer(fieldName, p.policy.UseDefault))
		default:
			plan.fieldPointer = goFieldIsPointer(parent, fieldName, p.policy.Pointer, p.policy.UseDefault)
		}
	}

	dataType := layoutAttribute.Type
	_, rawObject := dataType.(*expr.Object)
	plan.referencePointer = !rawObject && (expr.IsObject(dataType) || expr.IsUnion(dataType))
	switch actual := dataType.(type) {
	case expr.Primitive:
		plan.kind = GoPrimitive
		plan.primitive = GoNativeTypeName(actual)
		if custom, importSpec := GetMetaType(layoutAttribute); custom != "" {
			plan.primitive = custom
			if importSpec != nil {
				plan.directImport = GoTypeImport{Name: importSpec.Name, Path: importSpec.Path}
				plan.hasDirectImport = true
				plan.customQualifier = customTypeQualifier(custom, importSpec.Name)
			}
		}
	case *expr.Array:
		plan.kind = GoArray
		element, err := p.plan(actual.ElemType, owner, "", nil, expr.IsObject(actual.ElemType.Type))
		if err != nil {
			return nil, err
		}
		plan.element = element
	case *expr.Map:
		plan.kind = GoMap
		key, err := p.plan(actual.KeyType, owner, "", nil, expr.IsObject(actual.KeyType.Type))
		if err != nil {
			return nil, err
		}
		element, err := p.plan(actual.ElemType, owner, "", nil, expr.IsObject(actual.ElemType.Type))
		if err != nil {
			return nil, err
		}
		plan.key = key
		plan.element = element
	case *expr.Object:
		plan.kind = GoStruct
		plan.fields = make([]*GoTypePlan, len(*actual))
		for index, field := range *actual {
			child, err := p.plan(field.Attribute, owner, field.Name, layoutAttribute, false)
			if err != nil {
				return nil, err
			}
			plan.fields[index] = child
		}
	case expr.UserType:
		switch actual {
		case expr.Empty:
			plan.kind = GoEmpty
		case expr.ErrorResult:
			plan.kind = GoServiceError
			goaImport := GoaImport("")
			plan.directImport = GoTypeImport{Name: goaImport.Name, Path: goaImport.Path}
			plan.hasDirectImport = true
		default:
			plan.kind = GoNamed
			binding, err := p.binding(layoutAttribute, owner, GoNamed)
			if err != nil {
				return nil, err
			}
			plan.owner = binding.Owner
			plan.typeDeclaration = binding.Type
		}
	case *expr.Union:
		plan.kind = GoUnion
		binding, err := p.binding(layoutAttribute, owner, GoUnion)
		if err != nil {
			return nil, err
		}
		plan.owner = binding.Owner
		plan.unionDeclaration = binding.Union
		plan.branches = make([]*GoTypePlan, len(actual.Values))
		for index, branch := range actual.Values {
			child, err := p.plan(branch.Attribute, binding.Owner, branch.Name, nil, false)
			if err != nil {
				return nil, err
			}
			plan.branches[index] = child
		}
	default:
		return nil, fmt.Errorf("plan Go type: unsupported data type %T", actual)
	}
	return plan, nil
}

// binding obtains and validates one exact subsystem-owned declaration record.
func (p goTypePlanner) binding(attribute *expr.AttributeExpr, inheritedOwner string, kind GoTypeKind) (GoTypeBinding, error) {
	if p.bind == nil {
		return GoTypeBinding{}, fmt.Errorf("plan Go %s: declaration binder must not be nil", kind)
	}
	binding, err := p.bind(GoTypeBindingRequest{
		Attribute:      attribute,
		InheritedOwner: inheritedOwner,
		Kind:           kind,
	})
	if err != nil {
		return GoTypeBinding{}, fmt.Errorf("plan Go %s %q: %w", kind, attribute.Type.Name(), err)
	}
	if binding.Owner == "" {
		return GoTypeBinding{}, fmt.Errorf("plan Go %s %q: binding owner must not be empty", kind, attribute.Type.Name())
	}
	switch kind {
	case GoNamed:
		if binding.Type == nil || binding.Union != nil {
			return GoTypeBinding{}, fmt.Errorf("plan Go named type %q: binding requires only a type declaration", attribute.Type.Name())
		}
		if declarationOwner := binding.Type.PackagePath(); declarationOwner != binding.Owner {
			return GoTypeBinding{}, fmt.Errorf(
				"plan Go named type %q: binding owner %q does not match declaration owner %q",
				attribute.Type.Name(), binding.Owner, declarationOwner,
			)
		}
	case GoUnion:
		if binding.Union == nil || binding.Type != nil {
			return GoTypeBinding{}, fmt.Errorf("plan Go union %q: binding requires only a union declaration", attribute.Type.Name())
		}
		if declarationOwner := binding.Union.PackagePath(); declarationOwner != binding.Owner {
			return GoTypeBinding{}, fmt.Errorf(
				"plan Go union %q: binding owner %q does not match declaration owner %q",
				attribute.Type.Name(), binding.Owner, declarationOwner,
			)
		}
	}
	return binding, nil
}

// walk visits retained plans in stable pre-order without consulting expression
// contents.
func (p *GoTypePlan) walk(visit func(*GoTypePlan)) {
	visit(p)
	if p.key != nil {
		p.key.walk(visit)
	}
	if p.element != nil {
		p.element.walk(visit)
	}
	for _, field := range p.fields {
		field.walk(visit)
	}
	for _, branch := range p.branches {
		branch.walk(visit)
	}
}

// walkImports visits type spellings owned by the referring file. A named
// union's declaration file, not each file that refers to the union, owns the
// imports required by its branch definitions.
func (p *GoTypePlan) walkImports(visit func(*GoTypePlan)) {
	visit(p)
	if p.kind == GoUnion {
		return
	}
	if p.key != nil {
		p.key.walkImports(visit)
	}
	if p.element != nil {
		p.element.walkImports(visit)
	}
	for _, field := range p.fields {
		field.walkImports(visit)
	}
}

// customTypeQualifier returns the package identifier authored in a custom Go
// type. An explicit metadata alias wins; otherwise the first selector supplies
// the identifier while pointer and container syntax remains untouched.
func customTypeQualifier(typeName, alias string) string {
	if alias != "" {
		return alias
	}
	dot := strings.IndexByte(typeName, '.')
	if dot < 0 {
		return ""
	}
	start := dot
	for start > 0 {
		char, size := utf8.DecodeLastRuneInString(typeName[:start])
		if !goIdentifierRune(char) {
			break
		}
		start -= size
	}
	return typeName[start:dot]
}

// goIdentifierRune reports whether char may occur in a Go identifier.
func goIdentifierRune(char rune) bool {
	return char == '_' || unicode.IsLetter(char) || unicode.IsDigit(char)
}

// qualify resolves one retained external import and rejects an unusable alias.
func (l LinkedGoType) qualify(importPath string) string {
	if l.qualifier == nil {
		panic(fmt.Sprintf("format retained Go type import %q without qualifier lookup", importPath))
	}
	qualifier := l.qualifier(importPath)
	if qualifier == "" {
		panic(fmt.Sprintf("format retained Go type import %q with empty qualifier", importPath))
	}
	return qualifier
}

// qualifiedDeclaration renders one exact frozen declaration relative to the
// linked output package.
func (l LinkedGoType) qualifiedDeclaration(declaration *NameDeclaration) string {
	name := declaration.Name()
	if l.plan.owner == l.outputPath {
		return name
	}
	return l.qualify(l.plan.owner) + "." + name
}
