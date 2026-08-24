// This file derives imports from the gRPC endpoint sections rendered into one
// generated file.
package codegen

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// addEndpointImports adds the packages recorded for one service to a generated
// file and omits the package that contains the file itself.
func addEndpointImports(file *codegen.File, services *ServicesData, service *grpcServicePlan) *codegen.File {
	outputPath := strings.TrimPrefix(strings.ReplaceAll(file.Path, "\\", "/"), codegen.Gendir+"/")
	outputPackage := path.Join(services.GenPkg(), path.Dir(outputPath))
	owner := services.generation.Package(outputPackage)
	imports := make([]*codegen.ImportSpec, 0, len(service.imports))
	for _, importPath := range service.imports {
		if importPath != outputPackage {
			imports = append(imports, owner.Import(importPath))
		}
	}
	codegen.AddImport(file.SectionTemplates[0], imports...)
	return file
}

// grpcEndpointAttributes returns the named service attributes referenced by
// the supplied gRPC endpoint sections.
func grpcEndpointAttributes(endpoints ...*expr.GRPCEndpointExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range endpoints {
		method := endpoint.MethodExpr
		attributes = append(attributes, method.Payload, method.StreamingPayload, method.Result, method.StreamingResult)
		for _, methodError := range method.Errors {
			attributes = append(attributes, methodError.AttributeExpr)
		}
	}
	return attributes
}
