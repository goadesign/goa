// This file prepares every service design used by one run. It rejects missing
// or repeated designs and chooses each shared Go name once.
package service

import (
	"fmt"
	"path"
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
	servicePaths, err := allocateServicePackagePaths(generation.GenPkg(), inputs)
	if err != nil {
		return nil, err
	}
	plans := make([]*Plan, len(inputs))
	for index, input := range inputs {
		facts, err := collectRootFacts(input.Root, generation, input.Examples, servicePaths)
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

// ExampleImports returns copies of the application and interceptor imports
// selected while this service plan was created.
func (p *Plan) ExampleImports() []*codegen.ImportSpec {
	imports := make([]*codegen.ImportSpec, len(p.facts.exampleImports))
	for index, spec := range p.facts.exampleImports {
		copy := *spec
		imports[index] = &copy
	}
	return imports
}

// ProjectedResult returns a copy of the result fields included in the views for
// method. It reports an error when method is absent or has no views.
func (p *Plan) ProjectedResult(method *expr.MethodExpr) (*expr.AttributeExpr, error) {
	viewed, err := p.projectedResultFacts(method)
	if err != nil {
		return nil, err
	}
	projected := expr.AsObject(viewed.wrapped.Attribute().Type).Attribute("projected")
	return expr.DupAtt(projected), nil
}

// ProjectedResultDeclaration returns the generated view type used by method.
// HTTP and gRPC generators use the same record so references and names follow
// the exact type chosen by the service plan.
func (p *Plan) ProjectedResultDeclaration(method *expr.MethodExpr) (*codegen.TypeDeclaration, error) {
	viewed, err := p.projectedResultFacts(method)
	if err != nil {
		return nil, err
	}
	return viewed.projected.declaration, nil
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

// MethodPayloadLayout returns the Go fields stored by method's payload. For a
// named payload, it returns the definition containing those fields instead of
// the outer reference to the named type.
func (p *Plan) MethodPayloadLayout(method *expr.MethodExpr) (*codegen.GoTypePlan, error) {
	for _, service := range p.facts.services {
		facts := service.methodByExpr[method]
		if facts == nil {
			continue
		}
		if facts.payload == nil || facts.payload.layout == nil {
			return nil, fmt.Errorf("service method %q does not have a payload", method.Name)
		}
		if facts.payload.definition != nil {
			return facts.payload.definition, nil
		}
		return facts.payload.layout, nil
	}
	if method == nil {
		return nil, fmt.Errorf("service method is not part of this plan")
	}
	return nil, fmt.Errorf("service method %q is not part of this plan", method.Name)
}

// MethodResultLayout returns the Go fields stored by method's result. For a
// named result, it returns the definition containing those fields instead of
// the outer reference to the named type.
func (p *Plan) MethodResultLayout(method *expr.MethodExpr) (*codegen.GoTypePlan, error) {
	for _, service := range p.facts.services {
		facts := service.methodByExpr[method]
		if facts == nil {
			continue
		}
		if facts.result == nil || facts.result.layout == nil {
			return nil, fmt.Errorf("service method %q does not have a result", method.Name)
		}
		if facts.result.definition != nil {
			return facts.result.definition, nil
		}
		return facts.result.layout, nil
	}
	if method == nil {
		return nil, fmt.Errorf("service method is not part of this plan")
	}
	return nil, fmt.Errorf("service method %q is not part of this plan", method.Name)
}

// StreamingResultLayout returns the Go type layout used by method's stream.
// An explicitly empty streaming result uses the ordinary result layout, which
// is the type implemented by the generated client and server stream methods.
// Transport planners use this fact to decide whether their decoded value is
// directly assignable or needs a generated conversion.
func (p *Plan) StreamingResultLayout(method *expr.MethodExpr) (*codegen.GoTypePlan, error) {
	for _, service := range p.facts.services {
		facts := service.methodByExpr[method]
		if facts == nil {
			continue
		}
		if facts.streamingResult != nil && facts.streamingResult.present {
			return facts.streamingResult.layout, nil
		}
		if facts.result == nil || facts.result.layout == nil {
			return nil, fmt.Errorf("service method %q does not have a streaming result", method.Name)
		}
		return facts.result.layout, nil
	}
	if method == nil {
		return nil, fmt.Errorf("service method is not part of this plan")
	}
	return nil, fmt.Errorf("service method %q is not part of this plan", method.Name)
}

// ServicePackageImports returns the generated service and views package
// preferences recorded before Generation.Freeze. Transport generators use it
// for service-level files that may not contain a method, such as an HTTP file
// server.
func (p *Plan) ServicePackageImports(
	serviceExpression *expr.ServiceExpr,
) (servicePackage, viewsPackage *codegen.ImportSpec, err error) {
	for _, service := range p.facts.services {
		if service.service != serviceExpression {
			continue
		}
		serviceCopy := *service.packageImport
		viewsCopy := *service.viewsImport
		return &serviceCopy, &viewsCopy, nil
	}
	if serviceExpression == nil {
		return nil, nil, fmt.Errorf("service is not part of this plan")
	}
	return nil, nil, fmt.Errorf("service %q is not part of this plan", serviceExpression.Name)
}

// MethodPackageImports returns the generated service package preference and,
// for a viewed result, its views package preference. These are the names and
// paths recorded before Generation.Freeze; an importing output package may
// receive a numbered qualifier when another import requests the same name.
func (p *Plan) MethodPackageImports(
	method *expr.MethodExpr,
) (servicePackage, viewsPackage *codegen.ImportSpec, err error) {
	for _, service := range p.facts.services {
		facts := service.methodByExpr[method]
		if facts == nil {
			continue
		}
		serviceCopy := *service.packageImport
		if facts.viewedResult == nil {
			return &serviceCopy, nil, nil
		}
		viewsCopy := *service.viewsImport
		return &serviceCopy, &viewsCopy, nil
	}
	if method == nil {
		return nil, nil, fmt.Errorf("service method is not part of this plan")
	}
	return nil, nil, fmt.Errorf("service method %q is not part of this plan", method.Name)
}

// projectedResultFacts returns the stored view plan for one method and reports
// whether the method is missing or has no viewed result.
func (p *Plan) projectedResultFacts(method *expr.MethodExpr) (*viewedResultFacts, error) {
	for _, service := range p.facts.services {
		facts := service.methodByExpr[method]
		if facts == nil {
			continue
		}
		if facts.viewedResult == nil {
			return nil, fmt.Errorf("service method %q does not have a viewed result", method.Name)
		}
		return facts.viewedResult, nil
	}
	if method == nil {
		return nil, fmt.Errorf("service method is not part of this plan")
	}
	return nil, fmt.Errorf("service method %q is not part of this plan", method.Name)
}

// collectRootFacts reads one service design and chooses names used only by that
// design before shared files receive their names.
func collectRootFacts(root *expr.RootExpr, generation *codegen.Generation, examples *expr.ExampleGenerator, servicePaths map[string]string) (*rootFacts, error) {
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
		serviceFacts.packagePath = servicePaths[service.Name]
		serviceFacts.viewsPath = serviceFacts.packagePath + "/views"
		serviceFacts.packageImport = codegen.NewImport(
			strings.ToLower(codegen.Goify(service.Name, false)),
			serviceFacts.packagePath,
		)
		serviceFacts.viewsImport = codegen.NewImport(
			serviceFacts.packageImport.Name+"views",
			serviceFacts.viewsPath,
		)
		facts.services = append(facts.services, serviceFacts)
		facts.serviceByID[service.Name] = serviceFacts
	}
	rootPath := path.Dir(generation.GenPkg())
	facts.exampleImports = append(facts.exampleImports, codegen.NewImport(facts.examplePackageName, rootPath))
	for _, service := range facts.services {
		if len(service.serverInterceptors) > 0 || len(service.clientInterceptors) > 0 {
			facts.exampleImports = append(facts.exampleImports, codegen.NewImport("interceptors", rootPath+"/interceptors"))
			break
		}
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
