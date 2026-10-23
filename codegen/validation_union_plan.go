// Union validation keeps the chosen branch check and its field checks together.
// The renderer uses these retained layouts after generated names become fixed.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/expr"
)

type (
	// validationUnionPlan stores generated sum-type branch operations.
	validationUnionPlan struct {
		cases    []validationUnionCasePlan
		receiver *GoTypePlan
		optional bool
	}

	// validationUnionCasePlan stores one sum-type accessor and branch program.
	validationUnionCasePlan struct {
		typeTag   string
		fieldName string
		node      *validationPlanNode
	}

	// validatorCallPlan stores one exact nested validator declaration.
	validatorCallPlan struct {
		declaration   *NameDeclaration
		address       bool
		optionalUnion *GoTypePlan
	}
)

// planUnion records selection validation even when no branch has field rules.
// Optional value fields retain their empty-kind absence; present pointers and
// required values use the generated union's own validity check.
func (p *validationPlanner) planUnion(
	attribute *expr.AttributeExpr,
	layout, structure *GoTypePlan,
	required, valuePointer bool,
	path string,
) (*validationUnionPlan, error) {
	policy := layout.Policy()
	union := expr.AsUnion(attribute.Type)
	branches := structure.Branches()
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
			branches[index].ReferenceIsPointer(),
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
	planned := &validationUnionPlan{
		cases:    cases,
		optional: !required && !valuePointer,
	}
	if structure != layout {
		// Defined types do not inherit the underlying union's methods.
		// Retain that declaration for the receiver conversion.
		planned.receiver = structure
	}
	return planned, nil
}
