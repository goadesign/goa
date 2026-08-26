// This file records every package referenced by JSON-RPC templates before Go
// names become final, then writes those packages into each output file.
package codegen

import (
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// setFileImports replaces a file's header imports with the JSON-RPC imports
// and shared HTTP imports recorded for that exact path.
func setFileImports(file *codegen.File, service *servicePlan) *codegen.File {
	header := file.SectionTemplates[0]
	header.Data.(map[string]any)["Imports"] = []*codegen.ImportSpec(nil)
	codegen.AddImport(header, service.fileImportSpecs(file.Path)...)
	codegen.AddImport(header, service.data.FileImports(file.Path)...)
	return file
}

// planJSONRPCFileImports records the packages referenced by each JSON-RPC file
// and reserves their names in the package that will write the file.
func planJSONRPCFileImports(
	service *servicePlan,
	transport *expr.HTTPServiceExpr,
	client, server *codegen.GeneratedPackage,
	serviceImport, viewsImport *codegen.ImportSpec,
) error {
	service.fileImports = make(map[string]*codegen.GeneratedImportPlan)
	service.clientPackage = client
	service.serverPackage = server
	service.serviceImportPath = serviceImport.Path
	base := path.Join(codegen.Gendir, "jsonrpc", service.pathName)
	clientPath := path.Join(base, "client")
	serverPath := path.Join(base, "server")

	if err := service.recordFileImports(client, path.Join(clientPath, "client.go"), jsonRPCClientImports(service.jsonRPCServiceFacts), nil); err != nil {
		return err
	}
	serverGenerated := []*codegen.ImportSpec{serviceImport}
	if service.hasViewedResult {
		serverGenerated = append(serverGenerated, viewsImport)
	}
	if err := service.recordFileImports(server, path.Join(serverPath, "server.go"), jsonRPCServerImports(service.jsonRPCServiceFacts), serverGenerated); err != nil {
		return err
	}

	var clientCodecGenerated []*codegen.ImportSpec
	if service.hasViewedResult {
		clientCodecGenerated = []*codegen.ImportSpec{serviceImport, viewsImport}
	}
	if err := service.recordFileImports(client, path.Join(clientPath, "encode_decode.go"), jsonRPCClientCodecImports(transport), clientCodecGenerated); err != nil {
		return err
	}
	if err := service.recordFileImports(server, path.Join(serverPath, "encode_decode.go"), nil, nil); err != nil {
		return err
	}

	if !service.hasSSE {
		return nil
	}
	streamGenerated := jsonRPCStreamGeneratedImports(service.jsonRPCServiceFacts, serviceImport, viewsImport)
	if err := service.recordFileImports(client, path.Join(clientPath, "stream.go"), jsonRPCClientStreamImports(service.jsonRPCServiceFacts), streamGenerated); err != nil {
		return err
	}
	return service.recordFileImports(server, path.Join(serverPath, "sse.go"), jsonRPCServerStreamImports(service.jsonRPCServiceFacts), streamGenerated)
}

// nativeServiceName returns the service package name used by JSON-RPC
// templates in one output package. It returns an empty string when those
// templates do not refer to the service package.
func (s *servicePlan) nativeServiceName(client bool) string {
	if client {
		if !s.hasViewedResult && !s.streamUsesServicePackage {
			return ""
		}
		return s.clientPackage.ImportName(s.serviceImportPath)
	}
	return s.serverPackage.ImportName(s.serviceImportPath)
}

// recordFileImports reserves runtime and generated package names, then saves
// the packages used by one source file.
func (s *servicePlan) recordFileImports(
	output *codegen.GeneratedPackage,
	filePath string,
	runtimePackages, generatedPackages []*codegen.ImportSpec,
) error {
	imports := codegen.NewGeneratedImportPlan(output)
	if err := imports.Require(runtimePackages...); err != nil {
		return err
	}
	if err := imports.AddGenerated(generatedPackages...); err != nil {
		return err
	}
	s.fileImports[jsonRPCFilePath(filePath)] = imports
	return nil
}

// fileImportSpecs returns the final package names recorded for one generated
// file. Unknown paths are generator bugs.
func (s *servicePlan) fileImportSpecs(filePath string) []*codegen.ImportSpec {
	planned, ok := s.fileImports[jsonRPCFilePath(filePath)]
	if !ok {
		panic("JSON-RPC file has no import plan")
	}
	return planned.Imports()
}

// jsonRPCClientImports returns packages named by the generated client file.
func jsonRPCClientImports(facts jsonRPCServiceFacts) []*codegen.ImportSpec {
	paths := []string{"bytes", "context", "net/http", "sync", codegen.GoaNamedImport("http", "goahttp").Path, codegen.GoaImport("").Path}
	if facts.hasSSE {
		paths = append(paths, "errors", "fmt", "io", "mime")
	}
	if facts.hasNotification {
		paths = append(paths, "errors", "io")
	}
	return jsonRPCImports(paths...)
}

// jsonRPCServerImports returns packages named by the generated server file.
func jsonRPCServerImports(facts jsonRPCServiceFacts) []*codegen.ImportSpec {
	paths := []string{
		"context", "fmt", "net/http", codegen.GoaNamedImport("http", "goahttp").Path,
		codegen.GoaImport("jsonrpc").Path, codegen.GoaImport("").Path,
	}
	if facts.hasHTTP {
		paths = append(paths, "bufio", "errors", "io")
	}
	if facts.hasSSE {
		paths = append(paths, "bytes", "errors", "io", "sync")
	}
	if facts.hasHTTP && facts.hasSSE {
		paths = append(paths, "mime", "strconv", "strings")
	}
	return jsonRPCImports(paths...)
}

// jsonRPCClientCodecImports returns packages named by JSON-RPC response and
// viewed-result decoders added to the shared HTTP codec file.
func jsonRPCClientCodecImports(service *expr.HTTPServiceExpr) []*codegen.ImportSpec {
	var paths []string
	for _, endpoint := range service.HTTPEndpoints {
		if !endpoint.UsesSSE() && !endpoint.IsJSONRPCNotification() {
			paths = append(paths, "bytes", "errors", "io", "net/http", codegen.GoaNamedImport("http", "goahttp").Path, codegen.GoaImport("jsonrpc").Path)
		}
		if !jsonRPCMethodHasViewedResult(endpoint.MethodExpr) {
			continue
		}
		paths = append(paths, "bytes", "encoding/json", "io", "net/http", codegen.GoaNamedImport("http", "goahttp").Path, codegen.GoaImport("").Path)
		if endpoint.SSE != nil && (endpoint.SSE.IDField != "" || endpoint.SSE.EventField != "" || endpoint.SSE.RetryField != "") {
			paths = append(paths, "fmt")
		}
		if endpoint.SSE != nil && endpoint.SSE.RetryField != "" {
			paths = append(paths, "strconv")
		}
	}
	return jsonRPCImports(paths...)
}

// jsonRPCClientStreamImports returns packages named by every generated event
// stream client and adds strconv only when a retry value is decoded.
func jsonRPCClientStreamImports(facts jsonRPCServiceFacts) []*codegen.ImportSpec {
	paths := []string{
		"bufio", "bytes", "context", "encoding/json", "errors", "fmt", "io", "net/http", "strings", "sync",
		codegen.GoaNamedImport("http", "goahttp").Path, codegen.GoaImport("jsonrpc").Path,
	}
	if facts.hasSSERetry {
		paths = append(paths, "strconv")
	}
	return jsonRPCImports(paths...)
}

// jsonRPCServerStreamImports returns packages named by the generated event
// stream server and adds packages used only by retry or variable-view code.
func jsonRPCServerStreamImports(facts jsonRPCServiceFacts) []*codegen.ImportSpec {
	paths := []string{"context", codegen.GoaImport("jsonrpc").Path}
	if facts.hasSSERetry {
		paths = append(paths, "fmt")
	}
	if facts.hasVariableViewedStream {
		paths = append(paths, codegen.GoaImport("").Path)
	}
	return jsonRPCImports(paths...)
}

// jsonRPCStreamGeneratedImports returns generated service and view packages
// referenced by event types and viewed-result code in stream files.
func jsonRPCStreamGeneratedImports(
	facts jsonRPCServiceFacts,
	serviceImport, viewsImport *codegen.ImportSpec,
) []*codegen.ImportSpec {
	if !facts.streamUsesServicePackage {
		return nil
	}
	imports := []*codegen.ImportSpec{serviceImport}
	if facts.streamUsesResultViewPackage {
		imports = append(imports, viewsImport)
	}
	return imports
}

// jsonRPCImports returns one import for each path using the package names that
// JSON-RPC templates write directly.
func jsonRPCImports(paths ...string) []*codegen.ImportSpec {
	imports := make([]*codegen.ImportSpec, 0, len(paths))
	for _, importPath := range paths {
		var spec *codegen.ImportSpec
		switch importPath {
		case codegen.GoaImport("").Path:
			spec = codegen.GoaImport("")
		case codegen.GoaNamedImport("http", "goahttp").Path:
			spec = codegen.GoaNamedImport("http", "goahttp")
		default:
			spec = codegen.SimpleImport(importPath)
		}
		imports = append(imports, spec)
	}
	return imports
}

// jsonRPCMethodHasViewedResult reports whether method's result comes from a
// result type that defines views.
func jsonRPCMethodHasViewedResult(method *expr.MethodExpr) bool {
	userType, ok := method.Result.Type.(expr.UserType)
	if !ok {
		return false
	}
	_, ok = userType.Origin().(*expr.ResultTypeExpr)
	return ok
}

// jsonRPCFilePath returns the stable map key used for a generated source file.
func jsonRPCFilePath(filePath string) string {
	return strings.ReplaceAll(filePath, "\\", "/")
}
