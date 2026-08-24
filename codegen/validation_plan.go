// This file records every validation check and generated function call before
// Goa chooses Go names. It later writes those checks using the stored field
// shapes, rules, and chosen function names.
package codegen

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"

	"goa.design/goa/v3/expr"
)

type (
	// ValidatorBindingRequest describes one validation call for a nested user
	// type. Attribute is available only while planning. Layout supplies its
	// generated package and chosen declaration.
	ValidatorBindingRequest struct {
		// Attribute is the nested user type whose validation call is being prepared.
		Attribute *expr.AttributeExpr
		// Layout is the planned Go layout for Attribute.
		Layout *GoTypePlan
		// View selects which result fields the nested validator checks. Service
		// types and view-specific result copies use the default view, represented by
		// the empty string.
		View string
	}

	// ValidatorDeclarationBinder returns the package-level validation function
	// chosen before Goa starts writing files.
	ValidatorDeclarationBinder func(ValidatorBindingRequest) (*NameDeclaration, error)

	// ValidationPlanOptions configures one root validation operation.
	ValidationPlanOptions struct {
		// Required reports whether the root value is required.
		Required bool
		// Alias validates the root as the underlying value of a user-type alias.
		Alias bool
		// Bind resolves every nested non-alias user validation call.
		Bind ValidatorDeclarationBinder
	}

	// ValidationPlan stores the copied checks for one service or view value,
	// including calls to validators for nested fields. It keeps expression
	// pointers only to recognize the attribute supplied by the caller.
	ValidationPlan struct {
		layout       *GoTypePlan
		root         *validationPlanNode
		declarations []*NameDeclaration
	}

	// LinkedValidationPlan renders a ValidationPlan after Goa has chosen all
	// generated function names and package aliases.
	LinkedValidationPlan struct {
		plan   *ValidationPlan
		layout LinkedGoType
	}

	// validationPlanNode stores checks for one value and for its fields,
	// collection entries, or union branches.
	validationPlanNode struct {
		occurrence *expr.AttributeExpr
		layout     *GoTypePlan
		rules      validationRulePlan
		guard      bool
		call       *validatorCallPlan
		fields     []validationFieldPlan
		array      *validationArrayPlan
		mapValue   *validationMapPlan
		union      *validationUnionPlan
	}

	// validationRulePlan stores local effective validation values in template
	// execution order.
	validationRulePlan struct {
		values           []any
		format           string
		pattern          string
		exclusiveMinimum *float64
		minimum          *float64
		exclusiveMaximum *float64
		maximum          *float64
		minLength        *int
		maxLength        *int
		required         []validationRequiredPlan
		pointer          bool
		dereference      bool
		aliasCast        string
		stringValue      bool
		arrayValue       bool
		mapValue         bool
	}

	// validationRequiredPlan stores one generated required-field check.
	validationRequiredPlan struct {
		name      string
		fieldName string
		unionKind bool
	}

	// validationFieldPlan stores one object child and its context path segment.
	validationFieldPlan struct {
		name string
		node *validationPlanNode
	}

	// validationArrayPlan stores one element operation and presence policy.
	validationArrayPlan struct {
		element          *validationPlanNode
		checkNilElements bool
	}

	// validationMapPlan stores map key and value validation operations.
	validationMapPlan struct {
		key   *validationPlanNode
		value *validationPlanNode
	}

	// validationUnionPlan stores generated sum-type branch operations.
	validationUnionPlan struct {
		cases []validationUnionCasePlan
	}

	// validationUnionCasePlan stores one sum-type accessor and branch program.
	validationUnionCasePlan struct {
		typeTag   string
		fieldName string
		node      *validationPlanNode
	}

	// validatorCallPlan stores one exact nested validator declaration.
	validatorCallPlan struct {
		declaration *NameDeclaration
	}

	// validationPlanner performs all expression reads while validation checks and
	// function calls are copied into a plan.
	validationPlanner struct {
		bind         ValidatorDeclarationBinder
		declarations []*NameDeclaration
	}
)

// NewValidationPlan records every check needed for attribute before Goa chooses
// the final generated names. layout must describe the same attribute and the
// requested service or view representation.
func NewValidationPlan(attribute *expr.AttributeExpr, layout *GoTypePlan, options ValidationPlanOptions) (*ValidationPlan, error) {
	if attribute == nil {
		return nil, fmt.Errorf("plan validation: attribute must not be nil")
	}
	if layout == nil {
		return nil, fmt.Errorf("plan validation: Go type layout must not be nil")
	}
	if !layout.MatchesOccurrence(attribute) {
		return nil, fmt.Errorf("plan validation: Go type layout does not match the root attribute occurrence")
	}
	if !layout.Policy().SumType {
		return nil, fmt.Errorf("plan validation: service/view validation requires a sum-type Go layout")
	}
	planner := validationPlanner{bind: options.Bind}
	root, err := planner.plan(attribute, layout, options.Required, options.Alias, false, "root")
	if err != nil {
		return nil, err
	}
	return &ValidationPlan{
		layout:       layout,
		root:         root,
		declarations: planner.declarations,
	}, nil
}

// NeedsValidation reports whether Goa would generate at least one validation
// check for attribute with the given Go field layout.
func NeedsValidation(attribute *expr.AttributeExpr, policy GoLayoutPolicy) bool {
	return attributeNeedsValidation(attribute, policy, make(map[expr.UserType]struct{}))
}

// ValidatorDeclarations returns the exact nested validator declarations in
// stable call order. Repeated calls deliberately repeat the same pointer.
func (p *ValidationPlan) ValidatorDeclarations() []*NameDeclaration {
	return append([]*NameDeclaration(nil), p.declarations...)
}

// ImportPreferences returns each package needed by the stored validation
// checks. Goa and standard library packages keep the names used by the
// templates. A package containing another generated validator includes the
// name Goa should try first.
func (p *ValidationPlan) ImportPreferences() []GoTypeImport {
	seen := make(map[string]struct{})
	var imports []GoTypeImport
	add := func(goImport GoTypeImport) {
		if _, exists := seen[goImport.Path]; exists {
			return
		}
		seen[goImport.Path] = struct{}{}
		imports = append(imports, goImport)
	}
	if p.root.usesUTF8() {
		add(GoTypeImport{Path: "unicode/utf8"})
	}
	if !p.root.empty() {
		goa := GoaImport("")
		add(GoTypeImport{Name: goa.Name, Path: goa.Path})
	}
	for _, declaration := range p.declarations {
		owner := declaration.packagePath()
		if owner == p.layout.Owner() {
			continue
		}
		add(GoTypeImport{
			Name: strings.ToLower(Goify(path.Base(owner), false)),
			Path: owner,
		})
	}
	return imports
}

// Link joins p with the Go types and package aliases that Goa chose for the
// generated file.
func (p *ValidationPlan) Link(layout LinkedGoType) (LinkedValidationPlan, error) {
	if layout.plan != p.layout {
		return LinkedValidationPlan{}, fmt.Errorf("link validation: linked Go type does not belong to this validation plan")
	}
	return LinkedValidationPlan{plan: p, layout: layout}, nil
}

// Render returns validation code for target. context is the root name included
// in validation errors and may differ from the Go target expression.
func (p LinkedValidationPlan) Render(target, context string) string {
	return p.renderNode(p.plan.root, target, context)
}

// Imports returns each external validation package once, using its final alias.
// Imports already supplied by the linked Go layout are not
// repeated unless validation calls require them too.
func (p LinkedValidationPlan) Imports() []GoTypeImport {
	preferences := p.plan.ImportPreferences()
	if len(preferences) == 0 {
		return nil
	}
	imports := make([]GoTypeImport, len(preferences))
	for index, preference := range preferences {
		imports[index] = GoTypeImport{
			Name: p.layout.qualify(preference.Path),
			Path: preference.Path,
		}
	}
	return imports
}

// plan copies one recursive operation. nested distinguishes a user-type field
// call from a root definition whose anonymous layout is expanded in place.
func (p *validationPlanner) plan(attribute *expr.AttributeExpr, layout *GoTypePlan, required, alias, nested bool, path string) (*validationPlanNode, error) {
	if !layout.MatchesOccurrence(attribute) {
		return nil, fmt.Errorf("plan validation for %s: Go type layout occurrence does not match", path)
	}
	policy := layout.Policy()
	if userType, named := attribute.Type.(expr.UserType); named && !alias && nested {
		if !userTypeNeedsValidation(userType, policy, make(map[expr.UserType]struct{})) {
			return &validationPlanNode{occurrence: attribute, layout: layout}, nil
		}
		if p.bind == nil {
			return nil, fmt.Errorf("plan validation for %s: validator binder must not be nil", path)
		}
		declaration, err := p.bind(ValidatorBindingRequest{
			Attribute: attribute,
			Layout:    layout,
			View:      "",
		})
		if err != nil {
			return nil, fmt.Errorf("plan validation for %s: %w", path, err)
		}
		if declaration == nil {
			return nil, fmt.Errorf("plan validation for %s: validator declaration must not be nil", path)
		}
		if declaration.owner == nil {
			return nil, fmt.Errorf("plan validation for %s: validator declaration is not owned", path)
		}
		if declaration.packagePath() != layout.Owner() {
			return nil, fmt.Errorf(
				"plan validation for %s: validator owner %q does not match layout owner %q",
				path, declaration.packagePath(), layout.Owner(),
			)
		}
		p.declarations = append(p.declarations, declaration)
		return &validationPlanNode{
			occurrence: attribute,
			layout:     layout,
			call:       &validatorCallPlan{declaration: declaration},
		}, nil
	}

	node := &validationPlanNode{
		occurrence: attribute,
		layout:     layout,
		rules:      planValidationRules(attribute, layout, required, alias),
	}
	switch {
	case expr.IsObject(attribute.Type):
		object := expr.AsObject(attribute.Type)
		fields := layout.Fields()
		if len(fields) != len(*object) {
			return nil, fmt.Errorf("plan validation for %s: object layout has %d fields, expected %d", path, len(fields), len(*object))
		}
		node.fields = make([]validationFieldPlan, 0, len(fields))
		for index, field := range *object {
			child, err := p.plan(
				field.Attribute,
				fields[index],
				attribute.IsRequired(field.Name),
				expr.IsAlias(field.Attribute.Type),
				true,
				fmt.Sprintf("field %q", field.Name),
			)
			if err != nil {
				return nil, err
			}
			if child.empty() {
				continue
			}
			node.fields = append(node.fields, validationFieldPlan{name: field.Name, node: child})
		}
	case expr.IsArray(attribute.Type):
		array := expr.AsArray(attribute.Type)
		childLayout := layout.Elem()
		if childLayout == nil {
			return nil, fmt.Errorf("plan validation for %s: array layout has no element", path)
		}
		childPolicy := policy
		if expr.IsPrimitive(array.ElemType.Type) {
			childPolicy.Pointer = childLayout.definitionPointer
		}
		childLayout = childLayout.withPolicy(childPolicy)
		child, err := p.plan(array.ElemType, childLayout, true, expr.IsAlias(array.ElemType.Type), true, path+"[*]")
		if err != nil {
			return nil, err
		}
		checkNilElements := array.NonNullableElems &&
			(childLayout.definitionPointer || IsNilable(array.ElemType.Type))
		if !child.empty() || checkNilElements {
			node.array = &validationArrayPlan{
				element:          child,
				checkNilElements: checkNilElements,
			}
		}
	case expr.IsMap(attribute.Type):
		mapping := expr.AsMap(attribute.Type)
		childPolicy := policy
		childPolicy.Pointer = false
		keyLayout := layout.Key()
		valueLayout := layout.Elem()
		if keyLayout == nil || valueLayout == nil {
			return nil, fmt.Errorf("plan validation for %s: map layout is incomplete", path)
		}
		key, err := p.plan(mapping.KeyType, keyLayout.withPolicy(childPolicy), true, expr.IsAlias(mapping.KeyType.Type), true, path+".key")
		if err != nil {
			return nil, err
		}
		value, err := p.plan(mapping.ElemType, valueLayout.withPolicy(childPolicy), true, expr.IsAlias(mapping.ElemType.Type), true, path+"[key]")
		if err != nil {
			return nil, err
		}
		if !key.empty() || !value.empty() {
			node.mapValue = &validationMapPlan{key: key, value: value}
		}
	case expr.IsUnion(attribute.Type):
		union := expr.AsUnion(attribute.Type)
		branches := layout.Branches()
		if len(branches) != len(union.Values) {
			return nil, fmt.Errorf("plan validation for %s: union layout has %d branches, expected %d", path, len(branches), len(union.Values))
		}
		var cases []validationUnionCasePlan
		for index, branch := range union.Values {
			branchPolicy := policy
			branchPolicy.Pointer = branchPolicy.Pointer && expr.IsObject(branch.Attribute.Type)
			child, err := p.plan(
				branch.Attribute,
				branches[index].withPolicy(branchPolicy),
				true,
				expr.IsAlias(branch.Attribute.Type),
				true,
				fmt.Sprintf("union branch %q", branch.Name),
			)
			if err != nil {
				return nil, err
			}
			if child.empty() {
				continue
			}
			cases = append(cases, validationUnionCasePlan{
				typeTag:   branch.Name,
				fieldName: Goify(branch.Name, true),
				node:      child,
			})
		}
		if len(cases) > 0 {
			node.union = &validationUnionPlan{cases: cases}
		}
	}
	if nested && !node.empty() {
		node.guard = validationNeedsNilGuard(attribute, required, policy)
	}
	return node, nil
}

// planValidationRules copies every local effective validation rule.
func planValidationRules(attribute *expr.AttributeExpr, layout *GoTypePlan, required, alias bool) validationRulePlan {
	validation := expr.EffectiveValidation(attribute)
	if validation == nil {
		return validationRulePlan{}
	}
	policy := layout.Policy()
	unaliased := unalias(attribute.Type)
	pointer := policy.Pointer || !required && (attribute.DefaultValue == nil || !policy.UseDefault)
	rules := validationRulePlan{
		format:           string(validation.Format),
		pattern:          validation.Pattern,
		exclusiveMinimum: copyValidationFloat(validation.ExclusiveMinimum),
		minimum:          copyValidationFloat(validation.Minimum),
		exclusiveMaximum: copyValidationFloat(validation.ExclusiveMaximum),
		maximum:          copyValidationFloat(validation.Maximum),
		minLength:        copyValidationInt(validation.MinLength),
		maxLength:        copyValidationInt(validation.MaxLength),
		pointer:          pointer,
		dereference: pointer && expr.IsPrimitive(attribute.Type) &&
			unaliased.Kind() != expr.BytesKind && unaliased.Kind() != expr.AnyKind,
		stringValue: unaliased.Kind() == expr.StringKind,
		arrayValue:  expr.IsArray(attribute.Type),
		mapValue:    expr.IsMap(attribute.Type),
	}
	if validation.Values != nil {
		rules.values = make([]any, len(validation.Values))
		for index, value := range validation.Values {
			rules.values[index] = copyValidationValue(value)
		}
	}
	if custom, _ := GetMetaType(attribute); custom != "" {
		rules.format = ""
	}
	if alias {
		rules.aliasCast = unaliased.Name()
	}
	object := expr.AsObject(attribute.Type)
	fields := layout.Fields()
	for _, name := range generatedRequiredValidationNames(attribute, validation, policy) {
		var fieldName string
		for index, field := range *object {
			if field.Name == name {
				fieldName = fields[index].FieldName(true)
				break
			}
		}
		requiredAttribute := object.Attribute(name)
		rules.required = append(rules.required, validationRequiredPlan{
			name:      name,
			fieldName: fieldName,
			unionKind: expr.IsUnion(requiredAttribute.Type) &&
				policy.SumType && !(policy.UnionPointer && policy.Pointer),
		})
	}
	return rules
}

// generatedRequiredValidationNames retains required checks emitted for policy.
func generatedRequiredValidationNames(attribute *expr.AttributeExpr, validation *expr.ValidationExpr, policy GoLayoutPolicy) []string {
	object := expr.AsObject(attribute.Type)
	var names []string
	for _, name := range validation.Required {
		required := object.Attribute(name)
		if required == nil {
			continue
		}
		if !policy.Pointer && expr.IsPrimitive(required.Type) &&
			required.Type.Kind() != expr.BytesKind && required.Type.Kind() != expr.AnyKind {
			continue
		}
		if policy.IgnoreRequired && expr.IsPrimitive(required.Type) {
			continue
		}
		names = append(names, name)
	}
	return names
}

// validationNeedsNilGuard reports whether generated checks must first verify
// that the value is not nil.
func validationNeedsNilGuard(attribute *expr.AttributeExpr, required bool, policy GoLayoutPolicy) bool {
	if expr.IsArray(attribute.Type) || expr.IsMap(attribute.Type) {
		return false
	}
	if expr.IsUnion(attribute.Type) {
		return policy.UnionPointer && (!required || policy.Pointer)
	}
	return policy.Pointer || !required && (attribute.DefaultValue == nil || !policy.UseDefault)
}

// userTypeNeedsValidation reports whether a user-defined type or any value
// inside it needs a generated check. seen stops recursive types.
func userTypeNeedsValidation(userType expr.UserType, policy GoLayoutPolicy, seen map[expr.UserType]struct{}) bool {
	origin := userType.Origin()
	if _, exists := seen[origin]; exists {
		return false
	}
	seen[origin] = struct{}{}
	defer delete(seen, origin)
	return attributeNeedsValidation(userType.Attribute(), policy, seen)
}

// attributeNeedsValidation reports whether Goa would generate a check for the
// attribute or a value inside it.
func attributeNeedsValidation(attribute *expr.AttributeExpr, policy GoLayoutPolicy, seen map[expr.UserType]struct{}) bool {
	validation := expr.EffectiveValidation(attribute)
	if validation != nil {
		if len(validation.Values) > 0 || validation.Pattern != "" ||
			validation.ExclusiveMinimum != nil || validation.Minimum != nil ||
			validation.ExclusiveMaximum != nil || validation.Maximum != nil ||
			validation.MinLength != nil || validation.MaxLength != nil {
			return true
		}
		if validation.Format != "" {
			if custom, _ := GetMetaType(attribute); custom == "" {
				return true
			}
		}
		if len(generatedRequiredValidationNames(attribute, validation, policy)) > 0 {
			return true
		}
	}
	switch {
	case expr.IsObject(attribute.Type):
		for _, field := range *expr.AsObject(attribute.Type) {
			if nested, ok := field.Attribute.Type.(expr.UserType); ok && !expr.IsAlias(nested) {
				if userTypeNeedsValidation(nested, policy, seen) {
					return true
				}
				continue
			}
			if attributeNeedsValidation(field.Attribute, policy, seen) {
				return true
			}
		}
	case expr.IsArray(attribute.Type):
		array := expr.AsArray(attribute.Type)
		if array.NonNullableElems &&
			(IsNilable(array.ElemType.Type) || arrayElementIsPointer(array, policy.ArrayElementPointer)) {
			return true
		}
		return attributeNeedsValidation(array.ElemType, policy, seen)
	case expr.IsMap(attribute.Type):
		mapping := expr.AsMap(attribute.Type)
		mapPolicy := policy
		mapPolicy.Pointer = false
		return attributeNeedsValidation(mapping.KeyType, mapPolicy, seen) ||
			attributeNeedsValidation(mapping.ElemType, mapPolicy, seen)
	case expr.IsUnion(attribute.Type):
		for _, branch := range expr.AsUnion(attribute.Type).Values {
			branchPolicy := policy
			branchPolicy.Pointer = policy.Pointer && expr.IsObject(branch.Attribute.Type)
			if nested, ok := branch.Attribute.Type.(expr.UserType); ok && !expr.IsAlias(nested) {
				if userTypeNeedsValidation(nested, branchPolicy, seen) {
					return true
				}
				continue
			}
			if attributeNeedsValidation(branch.Attribute, branchPolicy, seen) {
				return true
			}
		}
	}
	return false
}

// renderNode writes the Go checks for node and its children without reading the
// original design expression.
func (p LinkedValidationPlan) renderNode(node *validationPlanNode, target, context string) string {
	if node.call != nil {
		name := p.validatorName(node.call.declaration)
		var buffer bytes.Buffer
		if err := userValT.Execute(&buffer, map[string]any{
			"call": fmt.Sprintf("%s(%s)", name, target),
			"goa":  p.goaPackage(),
		}); err != nil {
			panic(err)
		}
		return fmt.Sprintf("if %s != nil {\n\t%s\n}", target, buffer.String())
	}
	var sections []string
	if local := p.renderValidationRules(node.rules, target, context, !node.guard); local != "" {
		sections = append(sections, local)
	}
	for _, field := range node.fields {
		validation := p.renderNode(
			field.node,
			target+"."+field.node.layout.FieldName(true),
			context+"."+field.name,
		)
		if validation != "" {
			sections = append(sections, validation)
		}
	}
	if node.array != nil {
		validation := p.renderNode(node.array.element, "e", context+"[*]")
		var buffer bytes.Buffer
		if err := arrayValT.Execute(&buffer, map[string]any{
			"target":           target,
			"validation":       validation,
			"checkNilElements": node.array.checkNilElements,
			"context":          literalValidationPath(context),
			"goa":              p.goaPackage(),
		}); err != nil {
			panic(err)
		}
		sections = append(sections, buffer.String())
	}
	if node.mapValue != nil {
		keyValidation := p.renderNode(node.mapValue.key, "k", context+".key")
		if keyValidation != "" {
			keyValidation = "\n" + keyValidation
		}
		valueValidation := p.renderNode(node.mapValue.value, "v", context+"[key]")
		if valueValidation != "" {
			valueValidation = "\n" + valueValidation
		}
		var buffer bytes.Buffer
		if err := mapValT.Execute(&buffer, map[string]any{
			"target":          target,
			"keyValidation":   keyValidation,
			"valueValidation": valueValidation,
		}); err != nil {
			panic(err)
		}
		sections = append(sections, buffer.String())
	}
	if node.union != nil {
		cases := make([]map[string]any, len(node.union.cases))
		for index, unionCase := range node.union.cases {
			cases[index] = map[string]any{
				"typeTag":    unionCase.typeTag,
				"fieldName":  unionCase.fieldName,
				"validation": p.renderNode(unionCase.node, "actual", context+".value"),
			}
		}
		var buffer bytes.Buffer
		if err := unionSumValT.Execute(&buffer, map[string]any{"target": target, "cases": cases}); err != nil {
			panic(err)
		}
		sections = append(sections, buffer.String())
	}
	code := strings.Join(sections, "\n")
	if node.guard && code != "" {
		condition := fmt.Sprintf("if %s != nil {\n", target)
		code = condition + code + "\n}"
	}
	return code
}

// renderValidationRules renders copied local rules through the shared
// validation templates using the package names assigned to the linked file.
func (p LinkedValidationPlan) renderValidationRules(rules validationRulePlan, target, context string, localGuards bool) string {
	targetValue := target
	if rules.dereference {
		targetValue = "*" + targetValue
	}
	if rules.aliasCast != "" {
		targetValue = fmt.Sprintf("%s(%s)", rules.aliasCast, targetValue)
	}
	utf8Package := ""
	if rules.stringValue && (rules.minLength != nil || rules.maxLength != nil) {
		utf8Package = p.utf8Package()
	}
	data := map[string]any{
		"isPointer": rules.pointer && localGuards,
		"context":   literalValidationPath(context),
		"target":    target,
		"targetVal": targetValue,
		"goa":       p.goaPackage(),
		"utf8":      utf8Package,
		"string":    rules.stringValue,
		"array":     rules.arrayValue,
		"map":       rules.mapValue,
	}
	var rendered []string
	if rules.values != nil {
		data["values"] = rules.values
		rendered = appendValidationTemplate(rendered, enumValT, data)
	}
	if rules.format != "" {
		data["format"] = rules.format
		rendered = appendValidationTemplate(rendered, formatValT, data)
	}
	if rules.pattern != "" {
		data["pattern"] = rules.pattern
		rendered = appendValidationTemplate(rendered, patternValT, data)
	}
	if rules.exclusiveMinimum != nil {
		data["exclMin"] = *rules.exclusiveMinimum
		data["isExclMin"] = true
		rendered = appendValidationTemplate(rendered, exclMinMaxValT, data)
	}
	if rules.minimum != nil {
		data["min"] = *rules.minimum
		data["isMin"] = true
		rendered = appendValidationTemplate(rendered, minMaxValT, data)
	}
	if rules.exclusiveMaximum != nil {
		data["exclMax"] = *rules.exclusiveMaximum
		data["isExclMin"] = false
		rendered = appendValidationTemplate(rendered, exclMinMaxValT, data)
	}
	if rules.maximum != nil {
		data["max"] = *rules.maximum
		data["isMin"] = false
		rendered = appendValidationTemplate(rendered, minMaxValT, data)
	}
	if rules.minLength != nil {
		data["minLength"] = rules.minLength
		data["isMinLength"] = true
		delete(data, "maxLength")
		rendered = appendValidationTemplate(rendered, lengthValT, data)
	}
	if rules.maxLength != nil {
		data["maxLength"] = rules.maxLength
		data["isMinLength"] = false
		delete(data, "minLength")
		rendered = appendValidationTemplate(rendered, lengthValT, data)
	}
	for _, required := range rules.required {
		if required.unionKind {
			rendered = append(rendered, fmt.Sprintf(
				"if %s.%s.Kind() == \"\" {\n        err = %s.MergeErrors(err, %s.MissingFieldError(%q, %q))\n}",
				target, required.fieldName, p.goaPackage(), p.goaPackage(), required.name, context,
			))
			continue
		}
		rendered = append(rendered, fmt.Sprintf(
			"if %s.%s == nil {\n        err = %s.MergeErrors(err, %s.MissingFieldError(%q, %q))\n}",
			target, required.fieldName, p.goaPackage(), p.goaPackage(), required.name, context,
		))
	}
	return strings.Join(rendered, "\n")
}

// appendValidationTemplate executes one shared local validation template.
func appendValidationTemplate(rendered []string, validationTemplate *template.Template, data map[string]any) []string {
	var buffer bytes.Buffer
	if err := validationTemplate.Execute(&buffer, data); err != nil {
		panic(err)
	}
	if validation := strings.Trim(buffer.String(), "\n"); validation != "" {
		return append(rendered, validation)
	}
	return rendered
}

// validatorName qualifies one exact validator declaration for the linked file.
func (p LinkedValidationPlan) validatorName(declaration *NameDeclaration) string {
	name := declaration.Name()
	owner := declaration.packagePath()
	if owner == p.layout.outputPath {
		return name
	}
	return p.layout.qualify(owner) + "." + name
}

// goaPackage returns the final name of Goa's generated-error package.
func (p LinkedValidationPlan) goaPackage() string {
	return p.layout.qualify(GoaImport("").Path)
}

// utf8Package returns the final name of the standard UTF-8 package.
func (p LinkedValidationPlan) utf8Package() string {
	return p.layout.qualify("unicode/utf8")
}

// empty reports whether node emits any validation code.
func (n *validationPlanNode) empty() bool {
	return n.call == nil && n.rules.empty() && len(n.fields) == 0 &&
		n.array == nil && n.mapValue == nil && n.union == nil
}

// usesUTF8 reports whether this validation tree counts runes in a string.
func (n *validationPlanNode) usesUTF8() bool {
	if n.rules.stringValue && (n.rules.minLength != nil || n.rules.maxLength != nil) {
		return true
	}
	for _, field := range n.fields {
		if field.node.usesUTF8() {
			return true
		}
	}
	if n.array != nil && n.array.element.usesUTF8() {
		return true
	}
	if n.mapValue != nil && (n.mapValue.key.usesUTF8() || n.mapValue.value.usesUTF8()) {
		return true
	}
	if n.union != nil {
		for _, unionCase := range n.union.cases {
			if unionCase.node.usesUTF8() {
				return true
			}
		}
	}
	return false
}

// empty reports whether these rules would write no Go checks.
func (p validationRulePlan) empty() bool {
	return p.values == nil && p.format == "" && p.pattern == "" &&
		p.exclusiveMinimum == nil && p.minimum == nil &&
		p.exclusiveMaximum == nil && p.maximum == nil &&
		p.minLength == nil && p.maxLength == nil && len(p.required) == 0
}

// withPolicy copies the prepared type and changes only the rules used to write
// its Go value. It does not modify the original type description.
func (p *GoTypePlan) withPolicy(policy GoLayoutPolicy) *GoTypePlan {
	clone := *p
	clone.policy = policy
	if p.key != nil {
		clone.key = p.key.withPolicy(policy)
	}
	if p.element != nil {
		clone.element = p.element.withPolicy(policy)
	}
	if len(p.fields) > 0 {
		clone.fields = make([]*GoTypePlan, len(p.fields))
		for index, field := range p.fields {
			clone.fields[index] = field.withPolicy(policy)
		}
	}
	if len(p.branches) > 0 {
		clone.branches = make([]*GoTypePlan, len(p.branches))
		for index, branch := range p.branches {
			clone.branches[index] = branch.withPolicy(policy)
		}
	}
	return &clone
}

// copyValidationFloat copies one optional scalar rule value.
func copyValidationFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

// copyValidationInt copies one optional length rule value.
func copyValidationInt(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

// copyValidationValue detaches the mutable collection shapes accepted by Goa
// enum validations. Primitive values are immutable and remain shared.
func copyValidationValue(value any) any {
	switch actual := value.(type) {
	case expr.Val:
		copied := make(expr.Val, len(actual))
		for name, child := range actual {
			copied[name] = copyValidationValue(child)
		}
		return copied
	case expr.ArrayVal:
		copied := make(expr.ArrayVal, len(actual))
		for index, child := range actual {
			copied[index] = copyValidationValue(child)
		}
		return copied
	case expr.MapVal:
		copied := make(expr.MapVal, len(actual))
		for key, child := range actual {
			copied[copyValidationValue(key)] = copyValidationValue(child)
		}
		return copied
	case []any:
		copied := make([]any, len(actual))
		for index, child := range actual {
			copied[index] = copyValidationValue(child)
		}
		return copied
	case []byte:
		return append([]byte(nil), actual...)
	case map[string]any:
		copied := make(map[string]any, len(actual))
		for name, child := range actual {
			copied[name] = copyValidationValue(child)
		}
		return copied
	case map[any]any:
		copied := make(map[any]any, len(actual))
		for key, child := range actual {
			copied[copyValidationValue(key)] = copyValidationValue(child)
		}
		return copied
	default:
		return actual
	}
}
