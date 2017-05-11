package rest

import (
	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design/rest"
)

// MarshalTypes return the file containing the type definitions used by the HTTP
// transport.
func MarshalTypes(r *rest.ResourceExpr) codegen.File {
	return nil
}

const typeT = `
{{- range . }}
	{{- if .Description }}
	// {{ .Description }}
	{{- end }}
	{{ .VarName }} {{ .TypeDef }}
{{- end }}
`

const payloadConstructorT = `{{ printf "%s instantiates and validates the %s service %s endpoint payload." .Constructor .ServiceName .EndpointName | comment }}
func {{ .Constructor }}({{ if .BodyTypeRef }}body {{ .BodyTypeRef }}, {{ end }}{{ range .Params }}{{ .VarName }} {{ .TypeRef }}, {{ end }}) (*service.{{ .TypeName }}, error) {
	{{ if .ValidateBody -}}
        if err := body.Validate(); err != nil {
		return nil, err
	}
	{{- end }}
	p := service.{{ .TypeName }}{
	{{ range .Params }}{{ .FieldName }}: {{ .VarName }},
	{{ end -}}
	}

	{{- range $res, $bod := .ResultToBody }}
	p.{{ $res }} = body.{{ $bod }}
	{{ end -}}

	{{ if .Validate -}}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	{{- end }}
	return &p, nil
}
`
