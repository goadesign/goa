// This file renders starter service implementations and imports only the
// generated types referenced by each implementation's service methods.
package service

import (
	"os"
	"path"

	"goa.design/goa/v3/codegen"
)

type (
	// basicEndpointData contains the data needed to render a basic endpoint
	// implementation in the example service file.
	basicEndpointData struct {
		*MethodData
		// ServiceVarName is the service variable name.
		ServiceVarName string
		// PayloadFullRef is the fully qualified reference to the payload.
		PayloadFullRef string
		// ResultFullName is the fully qualified name of the result.
		ResultFullName string
		// ResultFullRef is the fully qualified reference to the result.
		ResultFullRef string
		// ResultIsStruct indicates that the result type is a struct.
		ResultIsStruct bool
		// ResultView is the view to render the result. It is set only if the
		// result type uses views.
		ResultView string
		// StreamInterface is the stream interface in the service package used
		// by the endpoint implementation.
		StreamInterface string
		// ExampleStructDeclaration is the starter implementation receiver.
		ExampleStructDeclaration *codegen.NameDeclaration
	}

	// exampleServiceData separates the generated service package declaration
	// name from the qualifier used by this example file.
	exampleServiceData struct {
		*Data
		// ServicePkg is the canonical generated service package qualifier.
		ServicePkg string
	}
)

// ExampleServiceFiles returns a basic implementation for every service
// retained by plan.
func ExampleServiceFiles(plan *Plan) []*codegen.File {
	var fw []*codegen.File
	for _, facts := range plan.facts.services {
		if f := exampleServiceFile(plan, facts, plan.facts.examplePackageName); f != nil {
			fw = append(fw, f)
		}
	}
	return fw
}

// exampleServiceFile renders a basic implementation from one retained service.
func exampleServiceFile(plan *Plan, facts *serviceFacts, apipkg string) *codegen.File {
	genpkg := plan.generation.GenPkg()
	services := plan.Services()
	data := services.Get(facts.name)
	svcName := data.PathName
	servicePath := path.Join(genpkg, svcName)
	servicePkg := services.aliases.name(servicePath)
	renderData := &exampleServiceData{Data: data, ServicePkg: servicePkg}
	fpath := svcName + ".go"
	if _, err := os.Stat(fpath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", apipkg, facts.imports.exampleService.specs),
		{
			Name:   "basic-service-struct",
			Source: serviceTemplates.Read(exampleServiceStructT),
			Data:   renderData,
		}, {
			Name:   "basic-service-init",
			Source: serviceTemplates.Read(exampleServiceInitT),
			Data:   renderData,
		},
	}
	if len(data.Schemes) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "security-authfuncs",
			Source: serviceTemplates.Read(exampleSecurityAuthfuncsT),
			Data:   renderData,
		})
	}
	outputPath := path.Dir(genpkg)
	for _, method := range facts.orderedMethods {
		sections = append(sections, basicEndpointSection(method, data, outputPath, services.aliases, servicePkg))
	}

	// Add HandleStream method for JSON-RPC WebSocket services (not SSE)
	if hasJSONRPCWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-handle-stream",
			Source: serviceTemplates.Read(jsonrpcHandleStreamT),
			Data:   renderData,
		})
	}

	return &codegen.File{
		Path:             fpath,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

// basicEndpointSection returns a starter implementation whose payload and
// result references come from the method's frozen generated-package records.
func basicEndpointSection(facts *methodFacts, svcData *Data, outputPath string, aliases *importAliases, servicePkg string) *codegen.SectionTemplate {
	md := svcData.Method(facts.name)
	ed := &basicEndpointData{
		MethodData:               md,
		ServiceVarName:           svcData.VarName,
		ExampleStructDeclaration: svcData.ExampleStructDeclaration,
	}
	if facts.payload != nil && facts.payload.layout.Kind() != codegen.GoEmpty {
		ed.PayloadFullRef = facts.payload.layout.Link(outputPath, retainedTypeQualifier(aliases)).Ref()
	}
	if facts.result != nil && facts.result.layout.Kind() != codegen.GoEmpty {
		linked := facts.result.layout.Link(outputPath, retainedTypeQualifier(aliases))
		ed.ResultFullName = linked.Name()
		ed.ResultFullRef = linked.Ref()
		ed.ResultIsStruct = facts.result.isObject
		if md.ViewedResult != nil {
			ed.ResultView = facts.viewedResult.viewName
		}
	}
	if md.ServerStream != nil {
		ed.StreamInterface = servicePkg + "." + md.ServerStreamDeclaration.Name()
	}
	return &codegen.SectionTemplate{
		Name:   "basic-endpoint",
		Source: serviceTemplates.Read(endpointT),
		Data:   ed,
	}
}
