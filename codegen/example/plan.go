// This file copies the server information used by generated examples and
// records every package imported by their server and client programs.
package example

import (
	"path"
	"path/filepath"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
)

type (
	// Plan stores copied API, service, server, and JSON-RPC names for one
	// example generation.
	Plan struct {
		rootByService map[*service.Plan]*Root
	}

	// Root stores the API name, service names, and server descriptions copied
	// from one design.
	Root struct {
		// APIName is the design API name written in example help text.
		APIName string
		// Services lists every design service in declaration order.
		Services []string
		// Servers lists the copied server values in declaration order.
		Servers []*Data
	}
)

// NewPlan copies the server information from each service plan and records the
// imports used by every generated server and command-line client.
func NewPlan(generation *codegen.Generation, services ...*service.Plan) (*Plan, error) {
	plan := &Plan{rootByService: make(map[*service.Plan]*Root, len(services))}
	for _, servicePlan := range services {
		design := servicePlan.Root()
		plannedRoot := &Root{
			APIName:  design.API.Name,
			Services: make([]string, len(design.Services)),
			Servers:  make([]*Data, len(design.API.Servers)),
		}
		for i, service := range design.Services {
			plannedRoot.Services[i] = service.Name
		}
		for i, server := range design.API.Servers {
			planned := buildServerData(server, design)
			if err := planMainPackages(generation, servicePlan, planned); err != nil {
				return nil, err
			}
			plannedRoot.Servers[i] = planned
		}
		plan.rootByService[servicePlan] = plannedRoot
	}
	return plan, nil
}

// Root returns the copied design description for servicePlan. The second
// result is false when servicePlan was not used to create this plan.
func (p *Plan) Root(servicePlan *service.Plan) (*Root, bool) {
	root, ok := p.rootByService[servicePlan]
	return root, ok
}

// planMainPackages records the imports used by one generated server and its
// command-line client before generation chooses their Go names.
func planMainPackages(generation *codegen.Generation, servicePlan *service.Plan, server *Data) error {
	rootPath := RootPath(generation.GenPkg())
	serverPath := path.Join(rootPath, "cmd", server.Dir)
	serverPackage, err := generation.ClaimOutputPackage(serverPath, filepath.Dir(server.serverMainPath))
	if err != nil {
		return err
	}
	server.serverPackage = serverPackage
	generated := make([]*codegen.ImportSpec, 0, len(server.Services)+2)
	for _, serviceName := range server.Services {
		serviceImport, _, err := servicePlan.ServicePackageImports(servicePlan.Root().Service(serviceName))
		if err != nil {
			return err
		}
		generated = append(generated, serviceImport)
	}
	hasInterceptors := false
	for _, serviceName := range server.Services {
		hasInterceptors = hasInterceptors || len(servicePlan.Root().Service(serviceName).ServerInterceptors) > 0
	}
	for _, spec := range servicePlan.ExampleImports() {
		if spec.Path == path.Join(rootPath, "interceptors") && !hasInterceptors {
			continue
		}
		generated = append(generated, spec)
	}
	if err := registerPackageImports(serverPackage, serverMainFixedImports(), generated); err != nil {
		return err
	}

	if server.DefaultTransport() == nil {
		return nil
	}
	clientPath := path.Join(rootPath, "cmd", server.Dir+"-cli")
	clientPackage, err := generation.ClaimOutputPackage(clientPath, filepath.Dir(server.clientMainPath))
	if err != nil {
		return err
	}
	server.clientPackage = clientPackage
	return registerPackageImports(clientPackage, clientMainFixedImports(server), nil)
}

// registerPackageImports records names written directly in templates first.
// Generated packages receive another name when a template already uses theirs.
func registerPackageImports(owner *codegen.GeneratedPackage, fixed, generated []*codegen.ImportSpec) error {
	for _, spec := range fixed {
		if err := owner.RequireImport(spec); err != nil {
			return err
		}
	}
	for _, spec := range generated {
		if err := owner.ReserveGeneratedImport(spec); err != nil {
			return err
		}
	}
	return nil
}

// packageImports returns the import declarations chosen for one generated
// file after generation has made every package name final.
func packageImports(owner *codegen.GeneratedPackage, planned []*codegen.ImportSpec) []*codegen.ImportSpec {
	imports := make([]*codegen.ImportSpec, len(planned))
	for index, spec := range planned {
		imports[index] = owner.Import(spec.Path)
	}
	return imports
}
