// This file owns the service planning lifecycle across every Goa design root
// in one generation. It validates complete input membership, collects each
// root once, and assigns files shared by multiple roots before names freeze.
package service

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// NewPlans collects every service root owned by generation in one operation.
// Root-local facts remain in separate plans, while declarations and files that
// can be shared across roots are assigned once across the complete input set.
func NewPlans(generation *codegen.Generation, inputs ...PlanInput) ([]*Plan, error) {
	owned := make(map[*expr.RootExpr]struct{})
	for _, candidate := range generation.Roots() {
		if root, ok := candidate.(*expr.RootExpr); ok {
			owned[root] = struct{}{}
		}
	}
	seen := make(map[*expr.RootExpr]struct{}, len(inputs))
	for _, input := range inputs {
		if _, ok := owned[input.Root]; !ok {
			return nil, rootMembershipError(input.Root)
		}
		if _, ok := seen[input.Root]; ok {
			return nil, fmt.Errorf("service root %p is planned more than once", input.Root)
		}
		seen[input.Root] = struct{}{}
	}
	if len(inputs) != len(owned) {
		return nil, fmt.Errorf(
			"service planning requires all %d generation roots, got %d",
			len(owned),
			len(inputs),
		)
	}
	plans := make([]*Plan, len(inputs))
	for index, input := range inputs {
		facts, err := collectRootFacts(input.Root, generation, input.Examples)
		if err != nil {
			return nil, err
		}
		plans[index] = &Plan{generation: generation, facts: facts}
	}
	allFacts := make([]*rootFacts, len(plans))
	for index, plan := range plans {
		allFacts[index] = plan.facts
	}
	if err := collectGeneratedPackageEmissions(allFacts); err != nil {
		return nil, err
	}
	if err := collectExternalConversions(allFacts, generation); err != nil {
		return nil, err
	}
	return plans, nil
}

// NewPlan collects the only service root owned by generation. Generations
// containing multiple service roots must use NewPlans so shared package files
// and receiver methods are planned once across the complete run.
func NewPlan(root *expr.RootExpr, generation *codegen.Generation, examples *expr.ExampleGenerator) (*Plan, error) {
	plans, err := NewPlans(generation, PlanInput{Root: root, Examples: examples})
	if err != nil {
		return nil, err
	}
	return plans[0], nil
}

// collectRootFacts retains one root's service facts and declares its
// root-owned symbols before run-wide file ownership is assigned.
func collectRootFacts(root *expr.RootExpr, generation *codegen.Generation, examples *expr.ExampleGenerator) (*rootFacts, error) {
	examplePackageScope := codegen.NewNameScope()
	for _, service := range root.Services {
		examplePackageScope.Unique(strings.ToLower(codegen.Goify(service.Name, false)))
	}
	facts := &rootFacts{
		root:               root,
		apiName:            root.API.Name,
		apiVersion:         root.API.Version,
		examplePackageName: examplePackageScope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api"),
		serviceByID:        make(map[string]*serviceFacts, len(root.Services)),
		types:              append([]expr.UserType(nil), root.Types...),
		rootTypes:          newRootTypeSet(root),
		examples:           examples,
	}
	for _, service := range root.Services {
		serviceFacts := collectServiceFacts(root, service, examples)
		serviceFacts.packagePath = servicePackagePath(generation.GenPkg(), service)
		serviceFacts.viewsPath = serviceFacts.packagePath + "/views"
		facts.services = append(facts.services, serviceFacts)
		facts.serviceByID[service.Name] = serviceFacts
	}
	if err := collectServiceDeclarations(facts, generation); err != nil {
		return nil, err
	}
	for _, serviceFacts := range facts.services {
		if err := collectServiceNames(serviceFacts, facts.rootTypes, generation); err != nil {
			return nil, err
		}
		if err := collectServiceTypeFacts(serviceFacts, facts.types, facts.rootTypes, generation); err != nil {
			return nil, err
		}
		if err := collectServiceUnionFacts(serviceFacts, facts.rootTypes, generation); err != nil {
			return nil, err
		}
		if err := planServiceTypeLayouts(serviceFacts, facts.rootTypes, generation); err != nil {
			return nil, err
		}
		if err := planServiceFileImports(serviceFacts, facts.rootTypes, generation); err != nil {
			return nil, err
		}
	}
	return facts, nil
}
