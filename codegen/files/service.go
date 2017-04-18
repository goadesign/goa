package files

import (
	"path/filepath"
	"text/template"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design"
)

type (
	// serviceData contains the data necessary to render the service
	// template.
	serviceData struct {
		// Name is the service name.
		Name string
		// VarName is the service struct name.
		VarName string
		// Methods lists the service struct methods.
		Methods []*serviceMethod
		// UserTypes lists the user types.
		UserTypes map[string]string
	}

	// serviceMethod describes a single service method.
	serviceMethod struct {
		// Name is the method name.
		Name string
		// VarName is the method struct name.
		VarName string
		// Payload is the payload of the method.
		Payload servicePayload
		// HasPayload is true if the payload type is not empty.
		HasPayload bool
		// Result is the result of the method.
		Result serviceResult
		// HasResult is true if the result type is not empty.
		HasResult bool
	}

	servicePayload struct {
		// Name is the payload name.
		Name string
		// Fileds lists the payload fields.
		Fields map[string]string
	}

	serviceResult struct {
		// Name is the result name.
		Name string
		// Fileds lists the result fields.
		Fields map[string]string
	}

	// serviceFile is the codgen file for a given service.
	serviceFile struct {
		service *design.ServiceExpr
	}
)

// serviceTmpl is the template used to render the body of the service file.
var serviceTmpl = template.Must(template.New("service").Parse(serviceT))

// Service returns the service file for the given service.
func Service(service *design.ServiceExpr) codegen.File {
	return &serviceFile{service}
}

// Sections returns the service file sections.
func (s *serviceFile) Sections(genPkg string) []*codegen.Section {
	var (
		data *serviceData
	)
	{
		methods := make([]*serviceMethod, len(s.service.Endpoints))
		userTypes := make(map[string]string)
		for i, v := range s.service.Endpoints {
			var walker func(*design.AttributeExpr) error
			walker = func(at *design.AttributeExpr) error {
				if ut, ok := at.Type.(design.UserType); ok {
					if _, ok := userTypes[ut.Name()]; ok {
						return nil
					}
					userTypes[ut.Name()] = codegen.GoTypeDef(ut.Attribute().Type)
					codegen.Walk(ut.Attribute(), walker)
				}
				return nil
			}

			payloadFields := make(map[string]string)
			if o := design.AsObject(v.Payload); o != nil {
				codegen.WalkAttributes(o, func(name string, at *design.AttributeExpr) error {
					payloadFields[name] = codegen.GoNativeType(at.Type)
					codegen.Walk(at, walker)
					return nil
				})
			}

			resultFields := make(map[string]string)
			if o := design.AsObject(v.Result); o != nil {
				codegen.WalkAttributes(o, func(name string, at *design.AttributeExpr) error {
					resultFields[name] = codegen.GoNativeType(at.Type)
					codegen.Walk(at, walker)
					return nil
				})
			}

			methods[i] = &serviceMethod{
				Name:    v.Name,
				VarName: codegen.Goify(v.Name, true),
				Payload: servicePayload{
					Name:   codegen.Goify(v.Payload.Name(), true),
					Fields: payloadFields,
				},
				HasPayload: v.Payload != design.Empty,
				Result: serviceResult{
					Name:   codegen.Goify(v.Result.Name(), true),
					Fields: resultFields,
				},
				HasResult: v.Result != design.Empty,
			}
		}
		data = &serviceData{
			Name:      s.service.Name,
			VarName:   codegen.Goify(s.service.Name, true),
			Methods:   methods,
			UserTypes: userTypes,
		}
	}

	var (
		header, body *codegen.Section
	)
	{
		header = codegen.Header(s.service.Name+"Services", "services",
			[]*codegen.ImportSpec{
				{Path: "context"},
				{Path: "goa.design/goa.v2"},
				{Path: genPkg + "/services"},
			})
		body = &codegen.Section{
			Template: serviceTmpl,
			Data:     data,
		}
	}

	return []*codegen.Section{header, body}
}

// OutputPath is the path to the generated service file relative to the output
// directory.
func (s *serviceFile) OutputPath(reserved map[string]bool) string {
	svc := codegen.SnakeCase(s.service.Name)
	return UniquePath(filepath.Join("services", svc+"%d.go"), reserved)
}

// serviceT is the template used to write an service definition.
const serviceT = `
{{- define "interface" }}
	// {{ .VarName }} is the {{ .Name }} service interface.
	{{ .VarName }} interface {
{{ range .Methods }}		// {{ .VarName }} implements the {{ .Name }} endpoint.
		{{ .Name }}(context.Context{{ if .HasPayload }}, *{{ .Payload.Name }}{{ end }}) {{ if .HasResult }}({{ .Result.Name }}, error){{ else }}error{{ end }}
{{ end }}	}
{{- end -}}

{{- define "payloads" }}
{{ range .Methods }}{{ if .HasPayload }}
	{{ .Payload.Name }} struct {
{{ range $key, $att := .Payload.Fields }}		{{ $key }} {{ $att }}
{{ end }}	}
{{ end }}{{ end -}}
{{- end -}}

{{- define "results" -}}
{{ range .Methods }}{{ if .HasResult }}
	{{ .Result.Name }} struct {
{{ range $key, $att := .Result.Fields }}		{{ $key }} {{ $att }}
{{ end }}	}
{{ end }}{{ end -}}
{{- end -}}

{{- define "types" -}}
{{ range $key, $ut := .UserTypes }}
	{{ $key }} {{ $ut }}
{{ end -}}
{{- end -}}

type (
{{- template "interface" . -}}
{{- template "payloads" . -}}
{{- template "results" . -}}
{{- template "types" . -}}
)
`
