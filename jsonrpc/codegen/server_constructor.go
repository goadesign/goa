// This file adds typed dependencies to JSON-RPC server construction. The HTTP
// plan owns their fields, arguments and examples; JSON-RPC retains the imports
// needed by its own server contribution before names become final.
package codegen

import (
	"fmt"
	"path"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// DeclareServerConstructorDependency adds a required dependency to this server
// and its generated example. It returns the application factory declaration;
// the HTTP plan keeps the shared constructor and field information.
func (p *Plan) DeclareServerConstructorDependency(service *expr.HTTPServiceExpr, name string, dependencyType *codegen.GoTypePlan, preferredFactory string, order codegen.PackageNameOrder) (*codegen.NameDeclaration, error) {
	planned, ok := p.servicesByExpr[service]
	if !ok {
		return nil, fmt.Errorf("JSON-RPC constructor dependency requires a service from this plan")
	}
	declaration, err := p.http.DeclareServerConstructorDependency(service, name, dependencyType, preferredFactory, order)
	if err != nil {
		return nil, err
	}
	filePath := path.Join(codegen.Gendir, "jsonrpc", planned.pathName, "server", "server.go")
	if err := planned.fileImports[filePath].AddTypeReference(dependencyType); err != nil {
		return nil, err
	}
	return declaration, nil
}
