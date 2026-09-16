{{- $name := .Name -}}
{{- if .Declaration -}}
{{- $name = .Declaration.Name -}}
{{- end }}
{{ printf "%s builds a value of type %s from a value of type %s." $name .ResultTypeRef .ParamTypeRef | comment }}
func {{ $name }}(v {{ .ParamTypeRef }}) {{ .ResultTypeRef }} {
	{{ .Code }}
	return res
}
