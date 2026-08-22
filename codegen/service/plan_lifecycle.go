// This file prepares every service design used by one run. It rejects missing
// or repeated designs and chooses each shared Go name once.
package service

import (
	"fmt"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// HTTPMethodNames contains the Go names used by one service method in an HTTP
	// package. The HTTP generator reuses these names instead of choosing new ones.
	HTTPMethodNames struct {
		// Method is the name used for the service endpoint field and receiver method.
		Method string
		// ServerStream is the service's server stream type name.
		ServerStream string
		// ClientStream is the service's client stream type name.
		ClientStream string
	}
)

// NewPlans reads every service design in generation. It returns the data used
// to write each service and chooses shared Go declaration names once.
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

// NewPlan reads the only service design in generation. Call NewPlans when a run
// contains several designs so shared files and methods receive names only once.
func NewPlan(root *expr.RootExpr, generation *codegen.Generation, examples *expr.ExampleGenerator) (*Plan, error) {
	plans, err := NewPlans(generation, PlanInput{Root: root, Examples: examples})
	if err != nil {
		return nil, err
	}
	return plans[0], nil
}

// Root returns the service design used by this plan. Other file writers use it
// to reject a plan created for a different design.
func (p *Plan) Root() *expr.RootExpr {
	return p.facts.root
}

// ProjectedResult returns a copy of the result fields included in the views for
// method. It reports an error when method is absent or has no views.
func (p *Plan) ProjectedResult(method *expr.MethodExpr) (*expr.AttributeExpr, error) {
	for _, service := range p.facts.services {
		facts := service.methodByExpr[method]
		if facts == nil {
			continue
		}
		if facts.viewedResult == nil {
			return nil, fmt.Errorf("service method %q does not have a viewed result", method.Name)
		}
		projected := expr.AsObject(facts.viewedResult.wrapped.Attribute().Type).Attribute("projected")
		return expr.DupAtt(projected), nil
	}
	if method == nil {
		return nil, fmt.Errorf("service method is not part of this plan")
	}
	return nil, fmt.Errorf("service method %q is not part of this plan", method.Name)
}

// HTTPMethodNames returns the Go names already chosen for method. It returns an
// error when the service design does not contain method.
func (p *Plan) HTTPMethodNames(method *expr.MethodExpr) (HTTPMethodNames, error) {
	for _, service := range p.facts.services {
		facts := service.methodByExpr[method]
		if facts == nil {
			continue
		}
		return HTTPMethodNames{
			Method:       facts.varName,
			ServerStream: facts.serverStreamVarName,
			ClientStream: facts.clientStreamVarName,
		}, nil
	}
	if method == nil {
		return HTTPMethodNames{}, fmt.Errorf("service method is not part of this plan")
	}
	return HTTPMethodNames{}, fmt.Errorf("service method %q is not part of this plan", method.Name)
}

// collectRootFacts reads one service design and chooses names used only by that
// design before shared files receive their names.
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
