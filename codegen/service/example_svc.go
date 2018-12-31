package service

import (
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/expr"
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
	}
)

// ExampleServiceFiles returns a basic service implementation for every
// service expression.
func ExampleServiceFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var fw []*codegen.File
	for _, svc := range root.Services {
		if f := exampleServiceFile(genpkg, root, svc); f != nil {
			fw = append(fw, f)
		}
	}
	return fw
}

// exampleServiceFile returns a basic implementation of the given service.
func exampleServiceFile(genpkg string, root *expr.RootExpr, svc *expr.ServiceExpr) *codegen.File {
	path := codegen.SnakeCase(svc.Name) + ".go"
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil // file already exists, skip it.
	}
	data := Services.Get(svc.Name)
	apiPkg := strings.ToLower(codegen.Goify(root.API.Name, false))
	sections := []*codegen.SectionTemplate{
		codegen.Header("", apiPkg, []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "log"},
			{Path: filepath.Join(genpkg, codegen.SnakeCase(svc.Name)), Name: data.PkgName},
		}),
		{Name: "basic-service-struct", Source: svcStructT, Data: data},
		{Name: "basic-service-init", Source: svcInitT, Data: data},
	}
	for _, m := range svc.Methods {
		sections = append(sections, basicEndpointSection(m, data))
	}

	return &codegen.File{
		Path:             path,
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
			if m.Result.Meta != nil {
				if v, ok := m.Result.Meta["view"]; ok {
					view = v[0]
				}
			}
			ed.ResultView = view
		}
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
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}, stream {{ .ServerStream.Interface }}) (err error) {
{{- else }}
func (s *{{ .ServiceVarName }}srvc) {{ .VarName }}(ctx context.Context{{ if .PayloadFullRef }}, p {{ .PayloadFullRef }}{{ end }}) ({{ if .ResultFullRef }}res {{ .ResultFullRef }}, {{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }} {{ end }}err error) {
{{- end }}
{{- if and (and .ResultFullRef .ResultIsStruct) (not .ServerStream) }}
  res = &{{ .ResultFullName }}{}
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
