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
