// This file records each Go type before generated package names are final.
// Formatting later uses the copied field names, package paths, and child types
// without reading the Goa expressions again.
package codegen

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"goa.design/goa/v3/expr"
)

type (
	// GoTypeKind states how a planned value is represented in Go.
	GoTypeKind uint8

	// GoTypeImport describes one package used in a planned Go type.
	GoTypeImport struct {
		// Name is the package name requested before the type, when present.
		Name string
		// Path is the Go import path.
		Path string
	}

	// GoTypeBindingRequest asks which generated declaration and package contain
	// one attribute's named type or union.
	GoTypeBindingRequest struct {
		// Attribute is the expression being planned.
		Attribute *expr.AttributeExpr
		// InheritedOwner is the package path inherited from the enclosing type.
		InheritedOwner string
		// Kind says whether Attribute contains a named type or a union.
		Kind GoTypeKind
	}

	// GoTypeBinding gives a planned attribute the package path and generated
	// declaration that will represent it. Service types set Type or Union;
	// other generators set Declaration.
	GoTypeBinding struct {
		// Owner is the import path of the package containing the declaration.
		Owner string
		// Type is the generated declaration for a named user type.
		Type *TypeDeclaration
		// Union is the generated declaration for a union.
		Union *UnionDeclaration
		// Declaration is the generated type declaration when the owning generator
		// does not register a service TypeDeclaration or UnionDeclaration.
		Declaration *NameDeclaration
		name        string
	}

	// GoTypeBinder returns the package path and generated declaration for a named
	// type or union. PlanGoType does not choose these values itself.
	GoTypeBinder func(GoTypeBindingRequest) (GoTypeBinding, error)

	// GoLayoutPolicy contains the pointer and validation choices used throughout
	// one planned Go type.
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
		// ArrayElementPointer uses pointers for required primitive array elements
		// when generated input validation must distinguish null from a zero value.
		ArrayElementPointer bool
		// SumType reports whether unions use Goa's generated struct form.
		SumType bool
	}

	// GoTypePlanOptions supplies the package, field name, and rules used to plan
	// one attribute.
	GoTypePlanOptions struct {
		// Owner is the package path that will contain the top-level attribute.
		Owner string
		// FieldName is the design field name of the top-level attribute, when set.
		FieldName string
		// Policy contains the pointer and validation choices selected by the caller.
		Policy GoLayoutPolicy
		// Bind returns the generated declaration for every named type and union.
		Bind GoTypeBinder
		// RetainNamedValue records the fields or elements beneath named types for
		// callers that render complete values.
		RetainNamedValue bool
	}

	// GoTypePlan stores the complete Go form copied from one attribute. It keeps
	// expression pointers only so callers can find which plans came from the same
	// attribute; its methods do not read those expressions.
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
		referenceNilable  bool
		primitive         string
		directImport      GoTypeImport
		hasDirectImport   bool
		customQualifier   string
		typeDeclaration   *TypeDeclaration
		unionDeclaration  *UnionDeclaration
		declaration       *NameDeclaration
		fixedName         string
		fields            []*GoTypePlan
		branches          []*GoTypePlan
		element           *GoTypePlan
		key               *GoTypePlan
		value             *GoTypePlan
	}

	// GoTypeQualifier returns the final package name written before a type from
	// the given import path.
	GoTypeQualifier func(importPath string) string

	// LinkedGoType formats a planned type for one output package after all type
	// names and imported package names are final.
	LinkedGoType struct {
		plan       *GoTypePlan
		outputPath string
		qualifier  GoTypeQualifier
	}

	// goTypePlanner reads attributes and builds GoTypePlan values.
	goTypePlanner struct {
		policy           GoLayoutPolicy
		bind             GoTypeBinder
		namedValues      map[expr.UserType]*GoTypePlan
		retainNamedValue bool
	}
)

const (
	// GoPrimitive is a built-in or explicitly imported primitive type.
	GoPrimitive GoTypeKind = iota + 1
	// GoArray is a slice with one planned element type.
	GoArray
	// GoMap is a map with planned key and element types.
	GoMap
	// GoStruct is an anonymous struct with fields in source order.
	GoStruct
	// GoNamed is a user type with a generated type declaration.
	GoNamed
	// GoUnion is a union with a generated union declaration.
	GoUnion
	// GoEmpty is Goa's built-in empty service type.
	GoEmpty
	// GoServiceError is Goa's built-in service error type.
	GoServiceError
)

// PlanGoType copies the Go form of attribute before generated type and imported
// package names are final. Callers format the result after Generation.Freeze
// chooses those names.
func PlanGoType(attribute *expr.AttributeExpr, options GoTypePlanOptions) (*GoTypePlan, error) {
	if attribute == nil {
		return nil, fmt.Errorf("plan Go type: attribute must not be nil")
	}
	if options.Owner == "" {
		return nil, fmt.Errorf("plan Go type: inherited owner must not be empty")
	}
	planner := goTypePlanner{
		policy:           options.Policy,
		bind:             options.Bind,
		namedValues:      make(map[expr.UserType]*GoTypePlan),
		retainNamedValue: options.RetainNamedValue,
	}
	return planner.plan(attribute, options.Owner, options.FieldName, nil, false)
}

// String returns the name of the Go type kind used in error messages.
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

// Kind returns how this planned value is represented in Go.
func (p *GoTypePlan) Kind() GoTypeKind {
	return p.kind
}

// Owner returns the import path of the package containing this type.
func (p *GoTypePlan) Owner() string {
	return p.owner
}

// Policy returns the pointer and validation choices used for this type.
func (p *GoTypePlan) Policy() GoLayoutPolicy {
	return p.policy
}

// MatchesOccurrence reports whether PlanGoType built this plan from attribute.
// It compares pointers without reading the expression.
func (p *GoTypePlan) MatchesOccurrence(attribute *expr.AttributeExpr) bool {
	return p.occurrence == attribute
}

// PlansForOccurrence returns every child plan built from attribute. The same
// attribute may produce several plans with different field names, package
// paths, or pointer choices.
func (p *GoTypePlan) PlansForOccurrence(attribute *expr.AttributeExpr) []*GoTypePlan {
	var matches []*GoTypePlan
	p.walk(func(candidate *GoTypePlan) {
		if candidate.occurrence == attribute {
			matches = append(matches, candidate)
		}
	})
	return matches
}

// TypeDeclaration returns the generated declaration for a named user type. It
// returns nil for every other kind.
func (p *GoTypePlan) TypeDeclaration() *TypeDeclaration {
	return p.typeDeclaration
}

// UnionDeclaration returns the generated declaration for a union. It returns
// nil for every other kind.
func (p *GoTypePlan) UnionDeclaration() *UnionDeclaration {
	return p.unionDeclaration
}

// FieldName returns the copied Go field name. It returns an exported name when
// firstUpper is true and an unexported name otherwise.
func (p *GoTypePlan) FieldName(firstUpper bool) string {
	if firstUpper {
		return p.fieldNameUpper
	}
	return p.fieldNameLower
}

// Description returns the description copied from the attribute.
func (p *GoTypePlan) Description() string {
	return p.description
}

// Tag returns the complete copied Go struct tag, including its leading space.
func (p *GoTypePlan) Tag() string {
	return p.tag
}

// IsPointer reports whether an enclosing struct stores this value through a
// pointer under the selected pointer and default rules.
func (p *GoTypePlan) IsPointer() bool {
	return p.fieldPointer
}

// ReferenceIsPointer reports whether this type is written with a leading
// pointer when it is used as a value. Named objects and unions use pointers;
// primitive aliases, arrays, and maps use values.
func (p *GoTypePlan) ReferenceIsPointer() bool {
	return p.referencePointer
}

// ReferenceCanBeNil reports whether the generated reference can represent an
// absent value without another pointer.
func (p *GoTypePlan) ReferenceCanBeNil() bool {
	return p.referenceNilable
}

// Import returns the package written directly in this type name. The second
// result is false for built-in types and generated declarations.
func (p *GoTypePlan) Import() (GoTypeImport, bool) {
	return p.directImport, p.hasDirectImport
}

// ImportPreferences returns each requested imported package name and each
// generated type package found in this plan, in field order. It keeps different
// requested names for the same path so Generation can choose the final name.
func (p *GoTypePlan) ImportPreferences() []GoTypeImport {
	seen := make(map[GoTypeImport]struct{})
	var imports []GoTypeImport
	p.walkImports(func(candidate *GoTypePlan) {
		var goImport GoTypeImport
		switch {
		case candidate.hasDirectImport:
			goImport = candidate.directImport
		case candidate.declaration != nil:
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

// Fields returns a copy of the anonymous struct fields in source order.
func (p *GoTypePlan) Fields() []*GoTypePlan {
	return append([]*GoTypePlan(nil), p.fields...)
}

// Branches returns a copy of the union branches in source order.
func (p *GoTypePlan) Branches() []*GoTypePlan {
	return append([]*GoTypePlan(nil), p.branches...)
}

// Elem returns the planned array or map element type. It returns nil for other
// kinds.
func (p *GoTypePlan) Elem() *GoTypePlan {
	return p.element
}

// Key returns the planned map key type. It returns nil for other kinds.
func (p *GoTypePlan) Key() *GoTypePlan {
	return p.key
}

// Equivalent reports whether p and other produce the same Go type. It compares
// declarations, package paths, pointer choices, field names, tags, imports, and
// child types, but does not compare source expression pointers.
func (p *GoTypePlan) Equivalent(other *GoTypePlan) bool {
	if p == nil || other == nil {
		return p == other
	}
	if p.kind != other.kind || p.owner != other.owner || p.policy != other.policy ||
		p.fieldNameUpper != other.fieldNameUpper || p.fieldNameLower != other.fieldNameLower ||
		p.description != other.description || p.comment != other.comment || p.tag != other.tag ||
		p.fieldPointer != other.fieldPointer || p.definitionPointer != other.definitionPointer ||
		p.referencePointer != other.referencePointer || p.referenceNilable != other.referenceNilable || p.primitive != other.primitive ||
		p.directImport != other.directImport || p.hasDirectImport != other.hasDirectImport ||
		p.customQualifier != other.customQualifier || p.typeDeclaration != other.typeDeclaration ||
		p.unionDeclaration != other.unionDeclaration || p.declaration != other.declaration || p.fixedName != other.fixedName || len(p.fields) != len(other.fields) ||
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

// Link prepares this plan for formatting in outputPath after generated type and
// imported package names are final. The returned value uses only data already
// copied into the plan.
func (p *GoTypePlan) Link(outputPath string, qualifier GoTypeQualifier) LinkedGoType {
	return LinkedGoType{plan: p, outputPath: outputPath, qualifier: qualifier}
}

// Name returns the Go type name selected by the plan.
func (l LinkedGoType) Name() string {
	if l.plan.fixedName != "" {
		return l.plan.fixedName
	}
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
		return l.qualifiedDeclaration(l.plan.declaration)
	case GoUnion:
		return l.qualifiedDeclaration(l.plan.declaration)
	case GoEmpty:
		return "struct {}"
	case GoServiceError:
		return l.qualify(l.plan.directImport.Path) + ".ServiceError"
	default:
		panic(fmt.Sprintf("format unknown retained Go type kind %d", l.plan.kind))
	}
}

// Def returns the complete Go type definition selected by the plan.
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

// Ref returns the Go type reference, including any pointer required for a named
// object or union.
func (l LinkedGoType) Ref() string {
	return l.RefWithPointer(l.plan.referencePointer)
}

// RefWithPointer returns this planned type with the requested top-level
// pointer. Child pointer choices remain unchanged.
func (l LinkedGoType) RefWithPointer(pointer bool) string {
	if pointer {
		return "*" + l.Name()
	}
	return l.Name()
}

// Kind returns how this linked value is represented in Go.
func (l LinkedGoType) Kind() GoTypeKind {
	return l.plan.Kind()
}

// ReferenceIsPointer reports whether the ordinary reference uses a pointer.
func (l LinkedGoType) ReferenceIsPointer() bool {
	return l.plan.ReferenceIsPointer()
}

// ReferenceCanBeNil reports whether the ordinary reference can represent an
// absent value.
func (l LinkedGoType) ReferenceCanBeNil() bool {
	return l.plan.ReferenceCanBeNil()
}

// Field returns the copied Go field name for this planned value.
func (l LinkedGoType) Field(firstUpper bool) string {
	return l.plan.FieldName(firstUpper)
}

// Package returns the package name written before this type when referenced
// from the output package. It returns an empty string when both types are in the
// same package.
func (l LinkedGoType) Package() string {
	if l.plan.owner == l.outputPath {
		return ""
	}
	return l.qualify(l.plan.owner)
}

// Enter returns a formatter for child that uses the same output package and
// imported package name lookup.
func (l LinkedGoType) Enter(child *GoTypePlan) LinkedGoType {
	if child == nil {
		panic("enter nil retained Go type plan")
	}
	return LinkedGoType{plan: child, outputPath: l.outputPath, qualifier: l.qualifier}
}

// Imports returns every package used by this type and its children except the
// output package itself.
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

// plan copies one attribute and recursively plans anonymous child types. Named
// types stop at their generated declaration.
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
	plan.referenceNilable = plan.referencePointer || !rawObject && IsNilable(dataType)
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
		elementPointer := expr.IsObject(actual.ElemType.Type) ||
			arrayElementIsPointer(actual, p.policy.ArrayElementPointer)
		element, err := p.plan(actual.ElemType, owner, "", nil, elementPointer)
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
			plan.declaration = binding.declaration()
			plan.fixedName = binding.name
			if p.retainNamedValue {
				origin := actual.Origin()
				value := p.namedValues[origin]
				if value == nil {
					value = &GoTypePlan{}
					p.namedValues[origin] = value
					planned, err := p.plan(actual.Attribute(), binding.Owner, "", nil, false)
					if err != nil {
						return nil, err
					}
					*value = *planned
				}
				plan.value = value
			}
		}
	case *expr.Union:
		plan.kind = GoUnion
		binding, err := p.binding(layoutAttribute, owner, GoUnion)
		if err != nil {
			return nil, err
		}
		plan.owner = binding.Owner
		plan.unionDeclaration = binding.Union
		plan.declaration = binding.declaration()
		plan.fixedName = binding.name
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

// The planner asks the caller for the generated declaration and checks that its
// package path and type match the request.
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
		if binding.declarationCount() != 1 || binding.Union != nil {
			return GoTypeBinding{}, fmt.Errorf("plan Go named type %q: binding requires one type declaration", attribute.Type.Name())
		}
		if binding.name == "" && binding.declaration().packagePath() != binding.Owner {
			declarationOwner := binding.declaration().packagePath()
			return GoTypeBinding{}, fmt.Errorf(
				"plan Go named type %q: binding owner %q does not match declaration owner %q",
				attribute.Type.Name(), binding.Owner, declarationOwner,
			)
		}
	case GoUnion:
		if binding.declarationCount() != 1 || binding.Type != nil {
			return GoTypeBinding{}, fmt.Errorf("plan Go union %q: binding requires one union declaration", attribute.Type.Name())
		}
		if binding.name == "" && binding.declaration().packagePath() != binding.Owner {
			declarationOwner := binding.declaration().packagePath()
			return GoTypeBinding{}, fmt.Errorf(
				"plan Go union %q: binding owner %q does not match declaration owner %q",
				attribute.Type.Name(), binding.Owner, declarationOwner,
			)
		}
	}
	return binding, nil
}

// declaration returns the generated name selected for this type binding.
func (b GoTypeBinding) declaration() *NameDeclaration {
	if b.Declaration != nil {
		return b.Declaration
	}
	if b.Type != nil {
		return b.Type.Declaration()
	}
	if b.Union != nil {
		return b.Union.Declaration()
	}
	return nil
}

// declarationCount reports how many mutually exclusive name sources are set.
func (b GoTypeBinding) declarationCount() int {
	count := 0
	if b.Type != nil {
		count++
	}
	if b.Union != nil {
		count++
	}
	if b.Declaration != nil {
		count++
	}
	if b.name != "" {
		count++
	}
	return count
}

// walk visits this plan and then its children in their stored order.
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

// walkImports visits the types whose packages must be imported by the current
// file. It stops at a union because the file that declares the union imports
// the packages used by its branches.
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

// customTypeQualifier returns the package name written in a custom Go type. It
// uses alias when provided; otherwise it reads the name before the first dot.
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

// qualify returns the package name written before types from importPath. It
// panics when no usable name was planned.
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

// qualifiedDeclaration adds the declaring package name before a generated type
// when that type is outside the output package.
func (l LinkedGoType) qualifiedDeclaration(declaration *NameDeclaration) string {
	name := declaration.Name()
	if l.plan.owner == l.outputPath {
		return name
	}
	return l.qualify(l.plan.owner) + "." + name
}
