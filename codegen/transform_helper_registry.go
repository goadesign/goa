// This file finds conversion functions that can share one declaration in a
// generated Go package. It compares the saved Go types, conversion rules, and
// child function calls before final package names are chosen.
package codegen

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// TransformHelperRegistry collects every conversion plan that writes helper
	// functions to one generated Go package. Finalize groups only helpers with
	// the same generated parameter type, result type, conversion rules, and
	// ordered child calls.
	TransformHelperRegistry struct {
		candidates []*transformHelperCandidate
		finalized  bool
	}

	// TransformHelperGroup describes helper occurrences that can use one Go
	// function declaration. Definition and Order return facts from the same
	// occurrence; Bind assigns the supplied declaration to every occurrence in
	// the group.
	TransformHelperGroup struct {
		definition TransformHelperDefinition
		order      PackageNameOrder
		candidates []*transformHelperCandidate
	}

	// TransformHelperOrder returns the package name order for one helper at its
	// exact field or collection location. The caller supplies the stable identity
	// of the conversion that contains the helper.
	TransformHelperOrder func(TransformHelperDefinitionLocation) PackageNameOrder

	// transformHelperCandidate stores one planned function call and the Go types
	// selected at that call's field or collection position.
	transformHelperCandidate struct {
		index        int
		plan         *TransformPlan
		helper       TransformHelper
		definition   TransformHelperDefinition
		order        PackageNameOrder
		sourceLayout *GoTypePlan
		targetLayout *GoTypePlan
		children     []*transformHelperCandidate
	}

	// transformGoLayoutPair stops comparison when two recursive generated types
	// refer back to layouts that were already compared.
	transformGoLayoutPair struct {
		left  *GoTypePlan
		right *GoTypePlan
	}

	// transformSemanticAttributePair stops comparison when two recursive Goa
	// types refer back to fields that were already compared.
	transformSemanticAttributePair struct {
		left  *expr.AttributeExpr
		right *expr.AttributeExpr
	}
)

// NewTransformHelperRegistry creates an empty package-level helper registry.
func NewTransformHelperRegistry() *TransformHelperRegistry {
	return &TransformHelperRegistry{}
}

// Collect adds every helper occurrence in plan. sourceLayout and targetLayout
// must describe the complete source and target values supplied to plan and
// retain the fields beneath named types. order must combine the conversion's
// stable identity with the location supplied to it.
func (r *TransformHelperRegistry) Collect(plan *TransformPlan, sourceLayout, targetLayout *GoTypePlan, order TransformHelperOrder) error {
	if r.finalized {
		return fmt.Errorf("transform helper registry is already finalized")
	}
	if plan == nil {
		return fmt.Errorf("transform helper plan must not be nil")
	}
	if sourceLayout == nil || targetLayout == nil {
		return fmt.Errorf("transform helper layouts must not be nil")
	}
	if order == nil {
		return fmt.Errorf("transform helper order must not be nil")
	}
	definitions := make(map[int]TransformHelperDefinition, len(plan.helpers))
	for _, definition := range plan.definitions {
		for _, index := range definition.helpers {
			definitions[index] = definition
		}
	}
	for _, helper := range plan.helpers {
		definition, ok := definitions[helper.ID.index]
		if !ok {
			return fmt.Errorf("transform helper occurrence %d has no definition", helper.Occurrence)
		}
		source, err := transformLayoutAtLocation(sourceLayout, plan.source, helper.location)
		if err != nil {
			return fmt.Errorf("find source layout for transform helper occurrence %d: %w", helper.Occurrence, err)
		}
		target, err := transformLayoutAtLocation(targetLayout, plan.target, helper.location)
		if err != nil {
			return fmt.Errorf("find target layout for transform helper occurrence %d: %w", helper.Occurrence, err)
		}
		candidateOrder := order(helper.location)
		if err := validatePackageNameOrder(candidateOrder); err != nil {
			return fmt.Errorf("order transform helper occurrence %d: %w", helper.Occurrence, err)
		}
		r.candidates = append(r.candidates, &transformHelperCandidate{
			index:        len(r.candidates),
			plan:         plan,
			helper:       helper,
			definition:   definition,
			order:        candidateOrder,
			sourceLayout: source,
			targetLayout: target,
		})
	}
	return nil
}

// Finalize groups all collected helpers. It may be called once, after every
// transform plan for the generated package has been collected and before the
// package chooses final declaration names.
func (r *TransformHelperRegistry) Finalize() ([]*TransformHelperGroup, error) {
	if r.finalized {
		return nil, fmt.Errorf("transform helper registry is already finalized")
	}
	r.finalized = true
	byID := make(map[TransformHelperID]*transformHelperCandidate, len(r.candidates))
	for _, candidate := range r.candidates {
		byID[candidate.helper.ID] = candidate
	}
	for _, candidate := range r.candidates {
		operation := candidate.plan.operations[candidate.helper.ID.index+1]
		candidate.children = make([]*transformHelperCandidate, len(operation.calls))
		for index, call := range operation.calls {
			child := byID[call.helper]
			if child == nil {
				return nil, fmt.Errorf("transform helper occurrence %d calls an uncollected helper", candidate.helper.Occurrence)
			}
			candidate.children[index] = child
		}
	}

	classes := make([]int, len(r.candidates))
	for {
		next := transformHelperClasses(r.candidates, classes)
		if slices.Equal(classes, next) {
			classes = next
			break
		}
		classes = next
	}
	groups := make([]*TransformHelperGroup, 0)
	for index, class := range classes {
		for len(groups) <= class {
			groups = append(groups, &TransformHelperGroup{})
		}
		group := groups[class]
		candidate := r.candidates[index]
		if len(group.candidates) == 0 || comparePackageNameOrders(candidate.order, group.order) < 0 {
			group.definition = candidate.definition
			group.definition.Location = candidate.helper.location
			group.order = candidate.order
		}
		group.candidates = append(group.candidates, candidate)
	}
	return groups, nil
}

// Order returns the stable package name order from the same helper occurrence
// returned by Definition.
func (g *TransformHelperGroup) Order() PackageNameOrder {
	return g.order
}

// Definition returns a detached description used to choose the group's Go
// function name. Changing the returned attributes does not change the plans.
func (g *TransformHelperGroup) Definition() TransformHelperDefinition {
	definition := g.definition
	sourceCopier := expr.NewAttributeGraphCopier()
	targetCopier := expr.NewAttributeGraphCopier()
	definition.Source = sourceCopier.Copy(definition.Source)
	definition.Target = targetCopier.Copy(definition.Target)
	definition.helpers = nil
	return definition
}

// Bind assigns declaration to every helper occurrence in the group. It checks
// the complete group before changing any plan, so an error leaves all helpers
// unchanged.
func (g *TransformHelperGroup) Bind(declaration *NameDeclaration) error {
	if declaration == nil {
		return fmt.Errorf("transform helper group declaration must not be nil")
	}
	if declaration.Kind() != NameFunction {
		return fmt.Errorf("transform helper group declaration must be a function, got %s", declaration.Kind())
	}
	for _, candidate := range g.candidates {
		bound := candidate.plan.helpers[candidate.helper.ID.index].Declaration
		if bound != nil && bound != declaration {
			return fmt.Errorf("transform helper occurrence %d already has a different declaration", candidate.helper.Occurrence)
		}
	}
	for _, candidate := range g.candidates {
		candidate.plan.helpers[candidate.helper.ID.index].Declaration = declaration
		candidate.plan.refreshTransformDefinitionDeclaration(candidate.definition.ID)
	}
	return nil
}

// refreshTransformDefinitionDeclaration records a declaration only when every
// occurrence in the plan-local definition uses that same declaration.
func (p *TransformPlan) refreshTransformDefinitionDeclaration(id TransformHelperDefinitionID) {
	definition := &p.definitions[id.index]
	declaration := p.helpers[definition.helpers[0]].Declaration
	for _, index := range definition.helpers[1:] {
		if p.helpers[index].Declaration != declaration {
			definition.Declaration = nil
			return
		}
	}
	definition.Declaration = declaration
}

// transformHelperClasses assigns a group from the function's own conversion
// rules and the groups assigned to its child calls on the previous pass.
func transformHelperClasses(candidates []*transformHelperCandidate, previous []int) []int {
	classes := make([]int, len(candidates))
	representatives := make([]int, 0, len(candidates))
	for index, candidate := range candidates {
		class := -1
		for candidateClass, representative := range representatives {
			if transformHelperCandidatesEqual(candidate, candidates[representative], previous) {
				class = candidateClass
				break
			}
		}
		if class < 0 {
			class = len(representatives)
			representatives = append(representatives, index)
		}
		classes[index] = class
	}
	return classes
}

// transformHelperCandidatesEqual compares the facts that determine one
// function body, then compares its child calls using the previous pass.
func transformHelperCandidatesEqual(left, right *transformHelperCandidate, previous []int) bool {
	if left.plan.hooks != nil || right.plan.hooks != nil {
		if left.plan != right.plan || left.definition.ID != right.definition.ID {
			return false
		}
	}
	if !transformGoLayoutsEqual(left.sourceLayout, right.sourceLayout, true, make(map[transformGoLayoutPair]struct{})) {
		return false
	}
	if !transformGoLayoutsEqual(left.targetLayout, right.targetLayout, true, make(map[transformGoLayoutPair]struct{})) {
		return false
	}
	if !transformSemanticAttributesEqual(left.helper.Source, right.helper.Source, make(map[transformSemanticAttributePair]struct{})) {
		return false
	}
	if !transformSemanticAttributesEqual(left.helper.Target, right.helper.Target, make(map[transformSemanticAttributePair]struct{})) {
		return false
	}
	if len(left.children) != len(right.children) {
		return false
	}
	for index := range left.children {
		if previous[left.children[index].index] != previous[right.children[index].index] {
			return false
		}
	}
	return true
}

// transformGoLayoutsEqual compares every generated type fact read by a
// conversion. The root field name and enclosing field pointer are excluded
// because a helper receives the field value as its own function parameter.
func transformGoLayoutsEqual(left, right *GoTypePlan, root bool, seen map[transformGoLayoutPair]struct{}) bool {
	if left == nil || right == nil {
		return left == right
	}
	pair := transformGoLayoutPair{left: left, right: right}
	if _, ok := seen[pair]; ok {
		return true
	}
	seen[pair] = struct{}{}
	if left.kind != right.kind || left.owner != right.owner || left.policy != right.policy ||
		left.referencePointer != right.referencePointer ||
		left.referenceNilable != right.referenceNilable || left.primitive != right.primitive ||
		left.directImport != right.directImport || left.hasDirectImport != right.hasDirectImport ||
		left.customQualifier != right.customQualifier || left.declaration != right.declaration ||
		left.fixedName != right.fixedName || len(left.fields) != len(right.fields) ||
		len(left.branches) != len(right.branches) {
		return false
	}
	if !root && (left.fieldNameUpper != right.fieldNameUpper || left.fieldNameLower != right.fieldNameLower ||
		left.fieldPointer != right.fieldPointer || left.definitionPointer != right.definitionPointer || left.tag != right.tag) {
		return false
	}
	if !transformGoLayoutsEqual(left.key, right.key, false, seen) ||
		!transformGoLayoutsEqual(left.element, right.element, false, seen) ||
		!transformGoLayoutsEqual(left.value, right.value, false, seen) {
		return false
	}
	for index := range left.fields {
		if !transformGoLayoutsEqual(left.fields[index], right.fields[index], false, seen) {
			return false
		}
	}
	for index := range left.branches {
		if !transformGoLayoutsEqual(left.branches[index], right.branches[index], false, seen) {
			return false
		}
	}
	return true
}

// transformSemanticAttributesEqual compares the Goa fields read by the normal
// conversion writers. The saved Go declarations identify named types, so
// separately copied Goa types do not differ by copy address.
func transformSemanticAttributesEqual(left, right *expr.AttributeExpr, seen map[transformSemanticAttributePair]struct{}) bool {
	if left == nil || right == nil {
		return left == right
	}
	pair := transformSemanticAttributePair{left: left, right: right}
	if _, ok := seen[pair]; ok {
		return true
	}
	seen[pair] = struct{}{}
	if !reflect.DeepEqual(left.DefaultValue, right.DefaultValue) {
		return false
	}
	leftObject := expr.AsObject(left.Type)
	rightObject := expr.AsObject(right.Type)
	if (leftObject == nil) != (rightObject == nil) {
		return false
	}
	if leftObject != nil {
		leftMapped := expr.NewMappedAttributeExpr(left)
		rightMapped := expr.NewMappedAttributeExpr(right)
		if len(*leftObject) != len(*rightObject) {
			return false
		}
		for _, field := range *leftObject {
			name := strings.SplitN(field.Name, ":", 2)[0]
			if leftMapped.IsRequired(name) != rightMapped.IsRequired(name) {
				return false
			}
		}
	}
	return transformSemanticDataTypesEqual(left.Type, right.Type, seen)
}

// transformSemanticDataTypesEqual compares field matching, required fields,
// collection rules, union branches, and nested defaults without comparing Goa
// value addresses.
func transformSemanticDataTypesEqual(left, right expr.DataType, seen map[transformSemanticAttributePair]struct{}) bool {
	if left == nil || right == nil {
		return left == right
	}
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false
	}
	switch left := left.(type) {
	case expr.Primitive:
		return left == right.(expr.Primitive)
	case *expr.Array:
		right := right.(*expr.Array)
		return left.NonNullableElems == right.NonNullableElems &&
			transformSemanticAttributesEqual(left.ElemType, right.ElemType, seen)
	case *expr.Object:
		right := right.(*expr.Object)
		if len(*left) != len(*right) {
			return false
		}
		for index, field := range *left {
			other := (*right)[index]
			if field.Name != other.Name || !transformSemanticAttributesEqual(field.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case *expr.Map:
		right := right.(*expr.Map)
		return transformSemanticAttributesEqual(left.KeyType, right.KeyType, seen) &&
			transformSemanticAttributesEqual(left.ElemType, right.ElemType, seen)
	case *expr.Union:
		right := right.(*expr.Union)
		if left.TypeKey != right.TypeKey || left.ValueKey != right.ValueKey || len(left.Values) != len(right.Values) {
			return false
		}
		for index, branch := range left.Values {
			other := right.Values[index]
			if branch.Name != other.Name || !transformSemanticAttributesEqual(branch.Attribute, other.Attribute, seen) {
				return false
			}
		}
		return true
	case expr.UserType:
		return transformSemanticAttributesEqual(left.Attribute(), right.(expr.UserType).Attribute(), seen)
	default:
		panic(fmt.Sprintf("cannot compare transform type %T", left))
	}
}

// transformLayoutAtLocation follows the authored field, collection, and union
// path saved by TransformPlan and returns the generated layout at that point.
func transformLayoutAtLocation(layout *GoTypePlan, attribute *expr.AttributeExpr, location TransformHelperDefinitionLocation) (*GoTypePlan, error) {
	remaining := location.encoded
	for len(remaining) > 0 {
		kind := remaining[0]
		remaining = remaining[1:]
		var name strings.Builder
		for {
			if len(remaining) < 2 {
				return nil, fmt.Errorf("invalid transform helper location")
			}
			if remaining[0] == 0 {
				if remaining[1] == 0xff {
					name.WriteByte(0)
					remaining = remaining[2:]
					continue
				}
				if remaining[1] == 0 {
					remaining = remaining[2:]
					break
				}
			}
			name.WriteByte(remaining[0])
			remaining = remaining[1:]
		}
		layout, attribute = transformLayoutValue(layout, attribute)
		switch kind {
		case transformObjectFieldLocation:
			object := expr.AsObject(attribute.Type)
			if object == nil || layout.kind != GoStruct {
				return nil, fmt.Errorf("object field %q does not select a generated struct", name.String())
			}
			found := false
			for index, field := range *object {
				if strings.SplitN(field.Name, ":", 2)[0] == name.String() {
					layout = layout.fields[index]
					attribute = field.Attribute
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("object field %q is missing", name.String())
			}
		case transformArrayElementLocation:
			array := expr.AsArray(attribute.Type)
			if array == nil || layout.element == nil {
				return nil, fmt.Errorf("array element does not select a generated element")
			}
			layout, attribute = layout.element, array.ElemType
		case transformMapKeyLocation, transformMapValueLocation:
			mapped := expr.AsMap(attribute.Type)
			if mapped == nil {
				return nil, fmt.Errorf("map component does not select a generated map")
			}
			if kind == transformMapKeyLocation {
				layout, attribute = layout.key, mapped.KeyType
			} else {
				layout, attribute = layout.element, mapped.ElemType
			}
		case transformUnionBranchLocation:
			union := expr.AsUnion(attribute.Type)
			if union == nil || layout.kind != GoUnion {
				return nil, fmt.Errorf("union branch %q does not select a generated union", name.String())
			}
			found := false
			for index, branch := range union.Values {
				if branch.Name == name.String() {
					layout = layout.branches[index]
					attribute = branch.Attribute
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("union branch %q is missing", name.String())
			}
		default:
			return nil, fmt.Errorf("unknown transform helper location step %d", kind)
		}
	}
	return layout, nil
}

// transformLayoutValue enters every consecutive named type while keeping the
// outer named layout available as the helper parameter or result type.
func transformLayoutValue(layout *GoTypePlan, attribute *expr.AttributeExpr) (*GoTypePlan, *expr.AttributeExpr) {
	for {
		userType, ok := attribute.Type.(expr.UserType)
		if !ok || userType == expr.Empty {
			return layout, attribute
		}
		attribute = userType.Attribute()
		if layout.kind == GoNamed && layout.value != nil {
			layout = layout.value
		}
	}
}
