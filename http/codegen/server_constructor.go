// This file keeps required server dependencies and their example constructors
// in one generation plan. Plugins provide a retained Go type; server fields,
// constructor arguments and example calls use that same type and declaration.
package codegen

import (
	"cmp"
	"fmt"
	"go/token"
	"path"
	"slices"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// ServerConstructorDependency describes one required constructor argument.
	// Generated servers retain the value in a private field with the same name.
	ServerConstructorDependency struct {
		// Name is the private field and constructor parameter name.
		Name string
		// TypeRef is the dependency type in the server's generated package.
		TypeRef string
		// Type retains the immutable type used for every emitted reference.
		Type *codegen.GoTypePlan
		// ExampleConstructor names the application factory emitted by goa example.
		ExampleConstructor *codegen.NameDeclaration
		// ExampleConstructorRef qualifies the factory in an example command.
		ExampleConstructorRef string
	}
)

// DeclareServerConstructorDependency adds a required value to this service's
// server constructor and returns its application factory declaration. The
// factory accepts no arguments and returns the dependency type. Goa example
// emits a factory that panics until the application supplies its construction.
// Declare dependencies before generation freezes names; nil never chooses an
// alternate server contract. The dependency type must already be planned.
// Required dependency arguments follow built-in arguments in name order.
func (p *Plan) DeclareServerConstructorDependency(service *expr.HTTPServiceExpr, name string, dependencyType *codegen.GoTypePlan, preferredFactory string, order codegen.PackageNameOrder) (*codegen.NameDeclaration, error) {
	if p.services != nil || p.generation.Frozen() {
		return nil, fmt.Errorf("server constructor dependency must be declared before generation freeze and plan linking")
	}
	extensions, ok := p.extensions[service]
	if !ok {
		return nil, fmt.Errorf("server constructor dependency requires a service from this plan")
	}
	if !token.IsIdentifier(name) || token.IsExported(name) || name == "_" {
		return nil, fmt.Errorf("server constructor dependency requires an unexported Go identifier")
	}
	if slices.Contains([]string{"e", "endpoints", "mux", "decoder", "encoder", "errhandler", "formatter", "upgrader", "configurer", "s"}, name) {
		return nil, fmt.Errorf("server constructor dependency %q conflicts with a built-in constructor parameter or field", name)
	}
	if dependencyType == nil {
		return nil, fmt.Errorf("server constructor dependency %q requires a retained Go type", name)
	}
	if slices.ContainsFunc(extensions.dependencies, func(dependency ServerConstructorDependency) bool {
		return dependency.Name == name
	}) {
		return nil, fmt.Errorf("server constructor dependency %q is already declared", name)
	}
	rootPath := path.Dir(p.generation.GenPkg())
	factory := codegen.NewPreferredName(codegen.NameFunction, preferredFactory, codegen.ExportedName, order)
	if err := p.generation.Package(rootPath).DeclareName(factory); err != nil {
		return nil, err
	}
	// Ordinary HTTP owns server.go here. JSON-RPC records that file's imports
	// in its own plan while consuming these same constructor facts.
	if p.transport == httpTransport {
		serverPath := path.Join(codegen.Gendir, transportDirectory(p.transport), p.servicePaths[service], "server", "server.go")
		if err := p.fileImports[serverPath].AddTypeReference(dependencyType); err != nil {
			return nil, err
		}
	}
	examplePath := p.dependencyExamplePath(service)
	exampleImports, found := p.fileImports[examplePath]
	if !found {
		exampleImports = codegen.NewGeneratedImportPlan(p.generation.Package(rootPath))
	}
	if err := exampleImports.AddTypeReference(dependencyType); err != nil {
		return nil, err
	}
	for _, imports := range extensions.dependencyExampleImports {
		if err := imports.AddGenerated(codegen.NewImport(examplePackageImportName(p.root), rootPath)); err != nil {
			return nil, err
		}
	}
	p.fileImports[examplePath] = exampleImports
	extensions.dependencies = append(extensions.dependencies, ServerConstructorDependency{
		Name:               name,
		Type:               dependencyType,
		ExampleConstructor: factory,
	})
	return factory, nil
}

// linkConstructorDependencies formats each retained dependency in the server
// package. Parameters that conflict with a designed file argument stop generation.
func (p *Plan) linkConstructorDependencies(service *expr.HTTPServiceExpr, data *ServiceData) error {
	dependencies := p.extensions[service].dependencies
	data.ConstructorDependencies = append([]ServerConstructorDependency(nil), dependencies...)
	slices.SortFunc(data.ConstructorDependencies, func(left, right ServerConstructorDependency) int {
		return cmp.Compare(left.Name, right.Name)
	})
	output := p.serverPackages[service]
	for index := range data.ConstructorDependencies {
		dependency := &data.ConstructorDependencies[index]
		for _, endpoint := range data.Endpoints {
			if endpoint.MultipartRequestDecoder != nil && dependency.Name == endpoint.MultipartRequestDecoder.VarName {
				return fmt.Errorf("server constructor dependency %q conflicts with a multipart decoder argument", dependency.Name)
			}
		}
		for _, fileServer := range data.FileServers {
			if dependency.Name == fileServer.ArgName {
				return fmt.Errorf("server constructor dependency %q conflicts with a file-system argument", dependency.Name)
			}
		}
		dependency.TypeRef = dependency.Type.Link(
			path.Join(p.generation.GenPkg(), transportDirectory(p.transport), p.servicePaths[service], "server"),
			output.ImportName,
		).Ref()
	}
	return nil
}

// dependencyExampleFiles emits factories for services with declared dependencies.
// Existing application files remain untouched by the example command.
func (p *Plan) dependencyExampleFiles() []*codegen.File {
	rootPath := path.Dir(p.generation.GenPkg())
	output := p.generation.Package(rootPath)
	var files []*codegen.File
	for _, service := range transportExpressions(p.root, p.transport).Services {
		dependencies := p.services.Get(service.Name()).ConstructorDependencies
		if len(dependencies) == 0 {
			continue
		}
		filePath := p.dependencyExamplePath(service)
		sections := []*codegen.SectionTemplate{
			codegen.Header("Required server dependencies", examplePackageImportName(p.root), p.services.fileImports[filePath]),
		}
		for _, dependency := range dependencies {
			sections = append(sections, &codegen.SectionTemplate{
				Name: "server-dependency-constructor",
				Source: `// {{ .Name }} constructs the required server dependency before requests begin.
// Replace this body with application configuration and dependency construction.
func {{ .Name }}() {{ .TypeRef }} {
    panic("{{ .Name }} requires application configuration")
}
`,
				Data: struct{ Name, TypeRef string }{
					Name:    dependency.ExampleConstructor.Name(),
					TypeRef: dependency.Type.Link(rootPath, output.ImportName).Ref(),
				},
			})
		}
		files = append(files, &codegen.File{Path: filePath, SectionTemplates: sections, SkipExist: true})
	}
	return files
}

// dependencyExamplePath keeps each transport's application dependency factories
// in one file so ordinary HTTP and JSON-RPC contributions cannot replace each other.
func (p *Plan) dependencyExamplePath(service *expr.HTTPServiceExpr) string {
	return p.servicePaths[service] + "_" + transportDirectory(p.transport) + "_dependencies.go"
}

// retainDependencyExampleImports records a configured server's import plan.
// A dependency declared before or after example planning adds the application
// import only to server startup files that call its factory.
func (p *Plan) retainDependencyExampleImports(service *expr.HTTPServiceExpr, imports *codegen.GeneratedImportPlan) error {
	extensions := p.extensions[service]
	if len(extensions.dependencies) > 0 {
		if err := imports.AddGenerated(codegen.NewImport(examplePackageImportName(p.root), path.Dir(p.generation.GenPkg()))); err != nil {
			return err
		}
	}
	extensions.dependencyExampleImports = append(extensions.dependencyExampleImports, imports)
	return nil
}

// exampleServerServiceDataForOutput copies service names for an example server
// and qualifies only the dependency factories that its startup code calls.
// CLI generation keeps its original service-only data and imports.
func exampleServerServiceDataForOutput(data *ServiceData, services *ServicesData, outputPackage string) *ServiceData {
	result := exampleServiceDataForOutput(data, services, outputPackage)
	result.ConstructorDependencies = append([]ServerConstructorDependency(nil), data.ConstructorDependencies...)
	for index := range result.ConstructorDependencies {
		dependency := &result.ConstructorDependencies[index]
		dependency.ExampleConstructorRef = services.PackageImport(outputPackage, path.Dir(services.GenPkg())).Name + "." + dependency.ExampleConstructor.Name()
	}
	return result
}
