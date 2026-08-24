// This file derives imports from the HTTP endpoints rendered into one
// generated file. Streaming-only files pass only their streaming endpoints.
package codegen

import (
	"path"
	"sort"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// addPlannedFileImports adds the service-type packages recorded for file before
// generation names were frozen.
func addPlannedFileImports(file *codegen.File, services *ServicesData) *codegen.File {
	if file == nil {
		return nil
	}
	codegen.AddImport(file.SectionTemplates[0], services.fileImports[filepathKey(file.Path)]...)
	return file
}

// generatedFileOutputPackage returns the import path of the package that owns
// a file written below the generated directory.
func generatedFileOutputPackage(services *ServicesData, filePath string) string {
	outputPath := strings.TrimPrefix(strings.ReplaceAll(filePath, "\\", "/"), codegen.Gendir+"/")
	return path.Join(services.GenPkg(), path.Dir(outputPath))
}

// serviceDataForOutput copies the package-name fields that a template writes
// so they match the imports selected by its actual output package.
func serviceDataForOutput(data *ServiceData, services *ServicesData, outputPackage string) *ServiceData {
	serviceCopy := *data.Service
	serviceCopy.PkgName = services.ServiceImport(outputPackage, data.Service.Name).Name
	copy := *data
	copy.Service = &serviceCopy
	copy.Endpoints = make([]*EndpointData, len(data.Endpoints))
	for index, endpoint := range data.Endpoints {
		endpointCopy := *endpoint
		endpointCopy.ServicePkgName = serviceCopy.PkgName
		copy.Endpoints[index] = &endpointCopy
	}
	return &copy
}

// exampleServiceDataForOutput gives local variables the unique service package
// path chosen for this generation. Example files may use several services at
// once, and two service names can produce the same Go name.
func exampleServiceDataForOutput(data *ServiceData, services *ServicesData, outputPackage string) *ServiceData {
	copy := serviceDataForOutput(data, services, outputPackage)
	copy.Service.VarName = codegen.Goify(copy.Service.PathName, false)
	return copy
}

// serviceReferenceAttributes returns the named service attributes referenced
// by generated HTTP or JSON-RPC endpoint sections, including the nested result
// field selected as SSE event data.
func serviceReferenceAttributes(endpoints ...*expr.HTTPEndpointExpr) []*expr.AttributeExpr {
	var attributes []*expr.AttributeExpr
	for _, endpoint := range endpoints {
		method := endpoint.MethodExpr
		attributes = append(attributes, method.Payload, method.StreamingPayload, method.Result, method.StreamingResult)
		if endpoint.SSE != nil && endpoint.SSE.DataField != "" {
			event := method.Result
			if method.HasMixedResults() {
				event = method.StreamingResult
			}
			if object := expr.AsObject(event.Type); object != nil {
				attributes = append(attributes, object.Attribute(endpoint.SSE.DataField))
			}
		}
		for _, methodError := range method.Errors {
			attributes = append(attributes, methodError.AttributeExpr)
		}
	}
	return attributes
}

// planHTTPAttributeImports records the metadata and relocated generated types
// referenced by transport conversion code in one output package.
func planHTTPAttributeImports(generation *codegen.Generation, outputPackage *codegen.GeneratedPackage, attributes ...*expr.AttributeExpr) ([]string, error) {
	seen := make(map[expr.UserType]struct{})
	paths := make(map[string]struct{})
	var visit func(*expr.AttributeExpr) error
	visit = func(attribute *expr.AttributeExpr) error {
		if attribute == nil || attribute.Type == expr.Empty {
			return nil
		}
		if _, spec := codegen.GetMetaType(attribute); spec != nil && spec.Path != outputPackage.ImportPath() {
			if err := outputPackage.DeclareImport(spec); err != nil {
				return err
			}
			paths[spec.Path] = struct{}{}
		}
		switch actual := attribute.Type.(type) {
		case expr.UserType:
			if location := codegen.UserTypeLocation(actual); location != nil {
				importPath := path.Join(generation.GenPkg(), location.RelImportPath)
				if importPath != outputPackage.ImportPath() {
					if err := outputPackage.ReserveGeneratedImport(codegen.NewImport(
						strings.ToLower(codegen.Goify(path.Base(importPath), false)),
						importPath,
					)); err != nil {
						return err
					}
					paths[importPath] = struct{}{}
				}
			}
			origin := actual.Origin()
			if _, ok := seen[origin]; ok {
				return nil
			}
			seen[origin] = struct{}{}
			return visit(actual.Attribute())
		case *expr.Object:
			for _, named := range *actual {
				if err := visit(named.Attribute); err != nil {
					return err
				}
			}
		case *expr.Array:
			return visit(actual.ElemType)
		case *expr.Map:
			if err := visit(actual.KeyType); err != nil {
				return err
			}
			return visit(actual.ElemType)
		case *expr.Union:
			for _, named := range actual.Values {
				if err := visit(named.Attribute); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, attribute := range attributes {
		if err := visit(attribute); err != nil {
			return nil, err
		}
	}
	result := make([]string, 0, len(paths))
	for importPath := range paths {
		result = append(result, importPath)
	}
	sort.Strings(result)
	return result, nil
}

// filepathKey normalizes generated paths so file writers on every platform
// read the same planned import record.
func filepathKey(filePath string) string {
	return strings.ReplaceAll(filePath, "\\", "/")
}

// httpWebSocketEndpoints returns only the endpoints whose stream sections are
// rendered into WebSocket files.
func httpWebSocketEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesWebSocket() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

// httpSSEEndpoints returns only the endpoints whose stream sections are
// rendered into Server-Sent Events files.
func httpSSEEndpoints(svc *expr.HTTPServiceExpr) []*expr.HTTPEndpointExpr {
	var endpoints []*expr.HTTPEndpointExpr
	for _, endpoint := range svc.HTTPEndpoints {
		if endpoint.UsesSSE() {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}
