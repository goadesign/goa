package service

import (
	"fmt"
	"path/filepath"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
)

// File returns the service file for the given service.
func File(genpkg string, service *design.ServiceExpr) *codegen.File {
	path := filepath.Join(codegen.Gendir, codegen.SnakeCase(service.Name), "service.go")
	svc := Services.Get(service.Name)
	header := codegen.Header(
		service.Name+" service",
		svc.PkgName,
		[]*codegen.ImportSpec{
			{Path: "context"},
			{Path: "goa.design/goa"},
			{Path: genpkg + "/" + codegen.SnakeCase(service.Name) + "/" + "views", Name: svc.ViewsPkg},
		})
	def := &codegen.SectionTemplate{
		Name:   "service",
		Source: serviceT,
		Data:   svc,
		FuncMap: map[string]interface{}{
			"streamInterfaceFor": streamInterfaceFor,
		},
	}

	sections := []*codegen.SectionTemplate{header, def}
	seen := make(map[string]struct{})

	for _, m := range svc.Methods {
		if m.PayloadDef != "" {
			if _, ok := seen[m.Payload]; !ok {
				seen[m.Payload] = struct{}{}
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "service-payload",
					Source: payloadT,
					Data:   m,
				})
			}
		}
		if m.ResultDef != "" {
			if _, ok := seen[m.Result]; !ok {
				seen[m.Result] = struct{}{}
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "service-result",
					Source: resultT,
					Data:   m,
				})
			}
		}
	}

	for _, ut := range svc.UserTypes {
		if _, ok := seen[ut.Name]; !ok {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "service-user-type",
				Source: userTypeT,
				Data:   ut,
			})
		}
	}

	var errorTypes []*UserTypeData
	for _, et := range svc.ErrorTypes {
		if et.Type == design.ErrorResult {
			continue
		}
		if _, ok := seen[et.Name]; !ok {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "error-user-type",
				Source: userTypeT,
				Data:   et,
			})
			errorTypes = append(errorTypes, et)
		}
	}

	for _, et := range errorTypes {
		if et.Type == design.ErrorResult {
			continue
		}
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "service-error",
			Source:  errorT,
			FuncMap: map[string]interface{}{"errorName": errorName},
			Data:    et,
		})
	}
	for _, er := range svc.ErrorInits {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "error-init-func",
			Source: errorInitT,
			Data:   er,
		})
	}

	// transform result type functions
	for _, t := range svc.ViewedResultTypes {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "viewed-result-type-to-service-result-type",
			Source: typeInitT,
			Data:   t.ResultInit,
		})
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "service-result-type-to-viewed-result-type",
			Source: typeInitT,
			Data:   t.Init,
		})
	}
	var projh []*codegen.TransformFunctionData
	for _, t := range svc.ProjectedTypes {
		for _, i := range t.TypeInits {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "projected-type-to-service-type",
				Source: typeInitT,
				Data:   i,
			})
		}
		for _, i := range t.Projections {
			projh = codegen.AppendHelpers(projh, i.Helpers)
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "service-type-to-projected-type",
				Source: typeInitT,
				Data:   i,
			})
		}
	}

	for _, h := range projh {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "transform-helpers",
			Source: transformHelperT,
			Data:   h,
		})
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func errorName(et *UserTypeData) string {
	obj := design.AsObject(et.Type)
	if obj != nil {
		for _, att := range *obj {
			if _, ok := att.Attribute.Metadata["struct:error:name"]; ok {
				return fmt.Sprintf("e.%s", codegen.Goify(att.Name, true))
			}
		}
	}
	return fmt.Sprintf("%q", et.Name)
}

// streamInterfaceFor builds the data to generate the client and server stream
// interfaces for the given endpoint.
func streamInterfaceFor(kind string, m *MethodData, stream *StreamData) map[string]interface{} {
	return map[string]interface{}{
		"Kind":           kind,
		"Endpoint":       m.Name,
		"Stream":         stream,
		"IsViewedResult": m.ViewedResult != nil,
	}
}

// serviceT is the template used to write an service definition.
const serviceT = `
{{ comment .Description }}
type Service interface {
{{- range .Methods }}
	{{ comment .Description }}
	{{- if .ViewedResult }}
		{{- if not .ViewedResult.ViewName }}
		{{ comment "The \"view\" return value must have one of the following views" }}
		{{- range .ViewedResult.Views }}
			{{- if .Description }}
			{{ printf "* %q: %s" .Name .Description | comment }}
			{{- else }}
			{{ printf "* %q" .Name | comment }}
			{{- end }}
		{{- end }}
		{{- end }}
	{{- end }}
	{{- if .ServerStream }}
	{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}, {{ .ServerStream.Interface }}) (err error)
	{{- else }}
	{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}) ({{ if .Result }}res {{ .ResultRef }}, {{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}{{ end }}err error)
	{{- end }}
{{- end }}
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = {{ printf "%q" .Name }}

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [{{ len .Methods }}]string{ {{ range .Methods }}{{ printf "%q" .Name }}, {{ end }} }
{{- range .Methods }}
	{{- if or .ServerStream .ClientStream }}
		{{ template "stream_interface" (streamInterfaceFor "server" . .ServerStream) }}
		{{ template "stream_interface" (streamInterfaceFor "client" . .ClientStream) }}
	{{- end }}
{{- end }}

{{- define "stream_interface" }}
{{ printf "%s is the interface a %q endpoint %s stream must satisfy." .Stream.Interface .Endpoint .Kind | comment }}
type {{ .Stream.Interface }} interface {
	{{- if .Stream.SendRef }}
		{{ printf "Send streams instances of %q." .Stream.SendName | comment }}
		Send({{ .Stream.SendRef }}) error
		{{ comment "Close closes the stream." }}
		Close() error
	{{- end }}
	{{- if .Stream.RecvRef }}
		{{ printf "Recv reads instances of %q from the stream." .Stream.RecvName | comment }}
		Recv() ({{ .Stream.RecvRef }}, error)
	{{- end }}
	{{- if and .IsViewedResult (eq .Kind "server") }}
		{{ comment "SetView sets the view used to render the result before streaming." }}
		SetView(view string)
	{{- end }}
}
{{- end }}
`

const payloadT = `{{ comment .PayloadDesc }}
type {{ .Payload }} {{ .PayloadDef }}
`

const resultT = `{{ comment .ResultDesc }}
type {{ .Result }} {{ .ResultDef }}
`

const userTypeT = `{{ comment .Description }}
type {{ .VarName }} {{ .Def }}
`

const errorT = `// Error returns an error description.
func (e {{ .Ref }}) Error() string {
	return {{ printf "%q" .Description }}
}

// ErrorName returns {{ printf "%q" .Name }}.
func (e {{ .Ref }}) ErrorName() string {
	return {{ errorName . }}
}
`

// input: map[string]{"Type": TypeData, "Error": ErrorData}
const errorInitT = `{{ printf "%s builds a %s from an error." .Name .TypeName |  comment }}
func {{ .Name }}(err error) {{ .TypeRef }} {
	return &{{ .TypeName }}{
		Name: {{ printf "%q" .ErrName }},
		ID: goa.NewErrorID(),
		Message: err.Error(),
	{{- if .Temporary }}
		Temporary: true,
	{{- end }}
	{{- if .Timeout }}
		Timeout: true,
	{{- end }}
	{{- if .Fault }}
		Fault: true,
	{{- end }}
	}
}
`

// input: InitData
const typeInitT = `{{ comment .Description }}
func {{ .Name }}({{ range .Args }}{{ .Name }} {{ .Ref }}, {{ end }}) {{ .ReturnTypeRef }} {
	{{ .Code }}
}
`
