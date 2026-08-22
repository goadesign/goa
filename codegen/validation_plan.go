// This file retains service and view validation operations before generated
// package names freeze. Linked validation rendering consumes only copied rules,
// symbolic Go layouts, and exact validator declarations.
package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa/v3/expr"
)

type (
	// ValidatorBindingRequest identifies one nested user-type validation call.
	// Attribute is available only while planning; Layout supplies its already
	// bound Go owner and declaration identity.
	ValidatorBindingRequest struct {
		// Attribute is the exact nested user-type occurrence.
		Attribute *expr.AttributeExpr
		// Layout is the exact symbolic Go layout for Attribute.
		Layout *GoTypePlan
		// View is the selected nested validator view. Service and projected-view
		// validation use the default view, represented by the empty string.
		View string
	}

	// ValidatorDeclarationBinder returns the exact package-level validator
	// declaration selected before generation freeze.
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

	// ValidationPlan is an immutable symbolic service/view validation program.
	// Expression pointers are retained only for occurrence identity; all rules,
	// paths, requiredness, layouts, and validator calls are copied during
	// NewValidationPlan.
	ValidationPlan struct {
		layout       *GoTypePlan
		root         *validationPlanNode
		declarations []*NameDeclaration
	}

	// LinkedValidationPlan renders a ValidationPlan after generated declaration
	// names and import aliases freeze.
	LinkedValidationPlan struct {
		plan   *ValidationPlan
		layout LinkedGoType
	}

	// validationPlanNode is one retained recursive validation operation.
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

	// validationRulePlan retains local effective validation values in template
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

	// validationRequiredPlan retains one generated required-field check.
	validationRequiredPlan struct {
		name      string
		fieldName string
		unionKind bool
	}

	// validationFieldPlan retains one object child and its context path segment.
	validationFieldPlan struct {
		name string
		node *validationPlanNode
	}

	// validationArrayPlan retains one element operation and presence policy.
	validationArrayPlan struct {
		element          *validationPlanNode
		nonNullableElems bool
	}

	// validationMapPlan retains map key and value validation operations.
	validationMapPlan struct {
		key   *validationPlanNode
		value *validationPlanNode
	}

	// validationUnionPlan retains generated sum-type branch operations.
	validationUnionPlan struct {
		cases []validationUnionCasePlan
	}

	// validationUnionCasePlan retains one sum-type accessor and branch program.
	validationUnionCasePlan struct {
		typeTag   string
		fieldName string
		node      *validationPlanNode
	}

	// validatorCallPlan retains one exact nested validator declaration.
	validatorCallPlan struct {
		declaration *NameDeclaration
	}

	// validationPlanner owns all expression reads during validation planning.
	validationPlanner struct {
		bind         ValidatorDeclarationBinder
		declarations []*NameDeclaration
	}
)

// NewValidationPlan selects every service/view validation operation for
// attribute before generated package names freeze. layout must be the exact
// sum-type Go plan built for attribute with the desired service or view policy.
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

// ValidatorDeclarations returns the exact nested validator declarations in
// stable call order. Repeated calls deliberately repeat the same pointer.
func (p *ValidationPlan) ValidatorDeclarations() []*NameDeclaration {
	return append([]*NameDeclaration(nil), p.declarations...)
}

// Link binds p to its exact linked Go layout after declaration and import
// aliases freeze.
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

// Imports returns path-unique external validator imports with their frozen
// qualifiers. Imports already supplied by the linked Go layout are not
// repeated unless validation calls require them too.
func (p LinkedValidationPlan) Imports() []GoTypeImport {
	seen := make(map[string]struct{})
	var imports []GoTypeImport
	for _, declaration := range p.plan.declarations {
		owner := declaration.packagePath()
		if owner == p.layout.outputPath {
			continue
		}
		if _, exists := seen[owner]; exists {
			continue
		}
		seen[owner] = struct{}{}
		imports = append(imports, GoTypeImport{
			Name: p.layout.qualify(owner),
			Path: owner,
		})
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
		childPolicy := policy
		if childPolicy.Pointer && expr.IsPrimitive(array.ElemType.Type) {
			childPolicy.Pointer = false
		}
		childLayout := layout.Elem()
		if childLayout == nil {
			return nil, fmt.Errorf("plan validation for %s: array layout has no element", path)
		}
		childLayout = childLayout.withPolicy(childPolicy)
		child, err := p.plan(array.ElemType, childLayout, true, expr.IsAlias(array.ElemType.Type), true, path+"[*]")
		if err != nil {
			return nil, err
		}
		if !child.empty() || array.NonNullableElems {
			node.array = &validationArrayPlan{
				element:          child,
				nonNullableElems: array.NonNullableElems,
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

// validationNeedsNilGuard retains validateAttribute's wrapper decision.
func validationNeedsNilGuard(attribute *expr.AttributeExpr, required bool, policy GoLayoutPolicy) bool {
	if expr.IsArray(attribute.Type) || expr.IsMap(attribute.Type) {
		return false
	}
	if expr.IsUnion(attribute.Type) {
		return policy.UnionPointer && (!required || policy.Pointer)
	}
	return policy.Pointer || !required && (attribute.DefaultValue == nil || !policy.UseDefault)
}

// userTypeNeedsValidation mirrors the existing nested-validator predicate
// without allocating names or retaining expression-backed render decisions.
func userTypeNeedsValidation(userType expr.UserType, policy GoLayoutPolicy, seen map[expr.UserType]struct{}) bool {
	origin := userType.Origin()
	if _, exists := seen[origin]; exists {
		return false
	}
	seen[origin] = struct{}{}
	return attributeNeedsValidation(userType.Attribute(), true, expr.IsAlias(userType), policy, seen)
}

// attributeNeedsValidation reports whether planning the attribute can emit code.
func attributeNeedsValidation(attribute *expr.AttributeExpr, required, alias bool, policy GoLayoutPolicy, seen map[expr.UserType]struct{}) bool {
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
			if attributeNeedsValidation(field.Attribute, attribute.IsRequired(field.Name), expr.IsAlias(field.Attribute.Type), policy, seen) {
				return true
			}
		}
	case expr.IsArray(attribute.Type):
		array := expr.AsArray(attribute.Type)
		if array.NonNullableElems {
			return true
		}
		return attributeNeedsValidation(array.ElemType, true, expr.IsAlias(array.ElemType.Type), policy, seen)
	case expr.IsMap(attribute.Type):
		mapping := expr.AsMap(attribute.Type)
		mapPolicy := policy
		mapPolicy.Pointer = false
		return attributeNeedsValidation(mapping.KeyType, true, expr.IsAlias(mapping.KeyType.Type), mapPolicy, seen) ||
			attributeNeedsValidation(mapping.ElemType, true, expr.IsAlias(mapping.ElemType.Type), mapPolicy, seen)
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
			if attributeNeedsValidation(branch.Attribute, true, expr.IsAlias(branch.Attribute.Type), branchPolicy, seen) {
				return true
			}
		}
	}
	return false
}

// renderNode renders retained operations without reading expression contents.
func (p LinkedValidationPlan) renderNode(node *validationPlanNode, target, context string) string {
	if node.call != nil {
		name := p.validatorName(node.call.declaration)
		var buffer bytes.Buffer
		if err := userValT.Execute(&buffer, map[string]any{"name": name, "target": target}); err != nil {
			panic(err)
		}
		return fmt.Sprintf("if %s != nil {\n\t%s\n}", target, buffer.String())
	}
	var sections []string
	if local := renderValidationRules(node.rules, target, context); local != "" {
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
			"nonNullableElems": node.array.nonNullableElems,
			"context":          context,
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
		if !strings.HasPrefix(code, condition) {
			code = condition + code + "\n}"
		}
	}
	return code
}

// renderValidationRules renders copied local rules through the canonical
// validation templates.
func renderValidationRules(rules validationRulePlan, target, context string) string {
	targetValue := target
	if rules.dereference {
		targetValue = "*" + targetValue
	}
	if rules.aliasCast != "" {
		targetValue = fmt.Sprintf("%s(%s)", rules.aliasCast, targetValue)
	}
	data := map[string]any{
		"isPointer": rules.pointer,
		"context":   context,
		"target":    target,
		"targetVal": targetValue,
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
				"if %s.%s.Kind() == \"\" {\n        err = goa.MergeErrors(err, goa.MissingFieldError(%q, %q))\n}",
				target, required.fieldName, required.name, context,
			))
			continue
		}
		rendered = append(rendered, fmt.Sprintf(
			"if %s.%s == nil {\n        err = goa.MergeErrors(err, goa.MissingFieldError(%q, %q))\n}",
			target, required.fieldName, required.name, context,
		))
	}
	return strings.Join(rendered, "\n")
}

// appendValidationTemplate executes one canonical local validation template.
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

// empty reports whether node emits any validation code.
func (n *validationPlanNode) empty() bool {
	return n.call == nil && n.rules.empty() && len(n.fields) == 0 &&
		n.array == nil && n.mapValue == nil && n.union == nil
}

// empty reports whether no local rule was retained.
func (p validationRulePlan) empty() bool {
	return p.values == nil && p.format == "" && p.pattern == "" &&
		p.exclusiveMinimum == nil && p.minimum == nil &&
		p.exclusiveMaximum == nil && p.maximum == nil &&
		p.minLength == nil && p.maxLength == nil && len(p.required) == 0
}

// withPolicy returns an occurrence-identical immutable plan view with a
// validation-specific effective policy. It does not modify the shared layout.
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
