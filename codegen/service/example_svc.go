// This file renders starter service implementations and imports only the
// generated types referenced by each implementation's service methods.
package service

import (
	"os"
	"path"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
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
	}
)

// ExampleServiceFiles returns a basic service implementation for every
// service expression.
func ExampleServiceFiles(genpkg string, root *expr.RootExpr, services *ServicesData) []*codegen.File {
	// determine the unique API package name different from the service names
	scope := codegen.NewNameScope()
	for _, svc := range root.Services {
		s := services.Get(svc.Name)
		if s == nil {
			panic("unknown service, " + svc.Name) // bug
		}
		scope.Unique(s.PkgName)
	}
	apipkg := scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")

	var fw []*codegen.File
	for _, svc := range root.Services {
		if f := exampleServiceFile(genpkg, root, svc, services, apipkg); f != nil {
			fw = append(fw, f)
		}
	}
	return fw
}

// exampleServiceFile returns a basic implementation of the given service.
func exampleServiceFile(genpkg string, _ *expr.RootExpr, svc *expr.ServiceExpr, services *ServicesData, apipkg string) *codegen.File {
	data := services.Get(svc.Name)
	svcName := data.PathName
	fpath := svcName + ".go"
	if _, err := os.Stat(fpath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	specs := []*codegen.ImportSpec{
		{Path: "io"},
		{Path: "context"},
		{Path: "fmt"},
		{Path: "strings"},
		{Path: path.Join(genpkg, svcName), Name: data.PkgName},
		{Path: "goa.design/clue/log"},
		{Path: "goa.design/goa/v3/security"},
	}
	specs = append(specs, services.AttributeImports(path.Dir(genpkg), serviceReferenceAttributes(svc)...)...)
	sections := []*codegen.SectionTemplate{
		codegen.Header("", apipkg, specs),
		{
			Name:   "basic-service-struct",
			Source: serviceTemplates.Read(exampleServiceStructT),
			Data:   data,
		}, {
			Name:   "basic-service-init",
			Source: serviceTemplates.Read(exampleServiceInitT),
			Data:   data,
		},
	}
	if len(data.Schemes) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "security-authfuncs",
			Source: serviceTemplates.Read(exampleSecurityAuthfuncsT),
			Data:   data,
		})
	}
	resolver := newServiceResolver(services.generation, services.aliases, svc, path.Dir(genpkg))
	for _, m := range svc.Methods {
		sections = append(sections, basicEndpointSection(m, data, resolver))
	}

	// Add HandleStream method for JSON-RPC WebSocket services (not SSE)
	if hasJSONRPCWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "jsonrpc-handle-stream",
			Source: serviceTemplates.Read(jsonrpcHandleStreamT),
			Data:   data,
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
func basicEndpointSection(m *expr.MethodExpr, svcData *Data, resolver *declarationResolver) *codegen.SectionTemplate {
	md := svcData.Method(m.Name)
	ed := &basicEndpointData{
		MethodData:     md,
		ServiceVarName: svcData.VarName,
	}
	if m.Payload.Type != expr.Empty {
		ed.PayloadFullRef = resolver.Ref(m.Payload, "")
	}
	if m.Result.Type != expr.Empty {
		ed.ResultFullName = resolver.Name(m.Result, "", false, true)
		ed.ResultFullRef = resolver.Ref(m.Result, "")
		ed.ResultIsStruct = expr.IsObject(m.Result.Type)
		if md.ViewedResult != nil {
			view := expr.DefaultView
			if v, ok := m.Result.Meta.Last(expr.ViewMetaKey); ok {
				view = v
			}
			ed.ResultView = view
		}
	}
	if md.ServerStream != nil {
		ed.StreamInterface = svcData.PkgName + "." + md.ServerStream.Interface
	}
	return &codegen.SectionTemplate{
		Name:   "basic-endpoint",
		Source: serviceTemplates.Read(endpointT),
		Data:   ed,
	}
}
