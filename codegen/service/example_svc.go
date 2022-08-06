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
func ExampleServiceFiles(genpkg string, root *expr.RootExpr) []*codegen.File {

	// determine the unique API package name different from the service names
	scope := codegen.NewNameScope()
	for _, svc := range root.Services {
		s := Services.Get(svc.Name)
		if s == nil {
			panic("unknown service, " + svc.Name) // bug
		}
		scope.Unique(s.PkgName)
	}
	apipkg := scope.Unique(strings.ToLower(codegen.Goify(root.API.Name, false)), "api")

	var fw []*codegen.File
	for _, svc := range root.Services {
		if f := exampleServiceFile(genpkg, root, svc, apipkg); f != nil {
			fw = append(fw, f)
		}
	}
	return fw
}

// exampleServiceFile returns a basic implementation of the given service.
func exampleServiceFile(genpkg string, root *expr.RootExpr, svc *expr.ServiceExpr, apipkg string) *codegen.File {
	data := Services.Get(svc.Name)
	svcName := data.PathName
	fpath := svcName + ".go"
	if _, err := os.Stat(fpath); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	specs := []*codegen.ImportSpec{
		{Path: "io"},
		{Path: "context"},
		{Path: "log"},
		{Path: "fmt"},
		{Path: "strings"},
		{Path: path.Join(genpkg, svcName), Name: data.PkgName},
		{Path: "goa.design/goa/v3/security"},
	}
	sections := []*codegen.SectionTemplate{
		codegen.Header("", apipkg, specs),
		{Name: "basic-service-struct", Source: svcStructT, Data: data},
		{Name: "basic-service-init", Source: svcInitT, Data: data},
	}
	if len(data.Schemes) > 0 {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "security-authfuncs",
			Source: dummyAuthFuncsT,
			Data:   data,
		})
	}
	for _, m := range svc.Methods {
		sections = append(sections, basicEndpointSection(m, data))
	}

	return &codegen.File{
		Path:             fpath,
		SectionTemplates: sections,
		SkipExist:        true,
	}
}

// basicEndpointSection returns a section with a basic implementation for the
// given method.
func basicEndpointSection(m *expr.MethodExpr, svcData *Data) *codegen.SectionTemplate {
	md := svcData.Method(m.Name)
	ed := &basicEndpointData{
		MethodData:     md,
		ServiceVarName: svcData.VarName,
	}
	if m.Payload.Type != expr.Empty {
		ed.PayloadFullRef = svcData.Scope.GoFullTypeRef(m.Payload, svcData.PkgName)
	}
	if m.Result.Type != expr.Empty {
		ed.ResultFullName = svcData.Scope.GoFullTypeName(m.Result, svcData.PkgName)
		ed.ResultFullRef = svcData.Scope.GoFullTypeRef(m.Result, svcData.PkgName)
		ed.ResultIsStruct = expr.IsObject(m.Result.Type)
		if md.ViewedResult != nil {
			view := "default"
			if v, ok := m.Result.Meta["view"]; ok {
				view = v[0]
			}
			ed.ResultView = view
		}
	}
	if md.ServerStream != nil {
		ed.StreamInterface = svcData.PkgName + "." + md.ServerStream.Interface
	}
	return &codegen.SectionTemplate{
		Name:   "basic-endpoint",
		Source: endpointT,
		Data:   ed,
	}
}

const (
	// input: service.Data
	svcStructT = `{{ printf "%s service example implementation.\nThe example methods log the requests and return zero values." .Name | comment }}
type {{ .VarName }}srvc struct {
	logger *log.Logger
}
`

	// input: service.Data
	svcInitT = `{{ printf "New%s returns the %s service implementation." .StructName .Name | comment }}
func New{{ .StructName }}(logger *log.Logger) {{ .PkgName }}.Service {
	return &{{ .VarName }}srvc{logger}
}
`

	// input: basicEndpointData
	endpointT = `{{ comment .Description }}
{{- if .ServerStream }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}, stream {{ .StreamInterface }}) (err error) {
{{- else }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}{{ if .SkipRequestBodyEncodeDecode }}, req io.ReadCloser{{ end }}) ({{ if .ResultFullRef }}res {{ .ResultFullRef }}, {{ end }}{{ if .SkipResponseBodyEncodeDecode }}resp io.ReadCloser, {{ end }}{{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}err error) {
{{- end }}
{{- if .SkipRequestBodyEncodeDecode }}
	// req is the HTTP request body stream.
	defer req.Close()
{{- end }}
{{- if and (and .ResultFullRef .ResultIsStruct) (not .ServerStream) }}
	res = &{{ .ResultFullName }}{}
{{- end }}
{{- if .SkipResponseBodyEncodeDecode }}
	// resp is the HTTP response body stream.
	resp = io.NopCloser(strings.NewReader("{{ .Name }}"))
{{- end }}
{{- if .ViewedResult }}
	{{- if not .ViewedResult.ViewName }}
		{{- if .ServerStream }}
			stream.SetView({{ printf "%q" .ResultView }})
		{{- else }}
			view = {{ printf "%q" .ResultView }}
		{{- end }}
	{{- end }}
{{- end }}
	s.logger.Print("{{ .ServiceVarName }}.{{ .Name }}")
	return
}
`
)
