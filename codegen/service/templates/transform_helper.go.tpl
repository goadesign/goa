{{- $name := .Name -}}
{{- if .Declaration -}}{{- $name = .Declaration.Name -}}{{- end }}
{{ printf "%s builds a value of type %s from a value of type %s." $name .ResultTypeRef .ParamTypeRef | comment }}
func {{ if .Declaration }}{{ .Declaration.Name }}{{ else }}{{ .Name }}{{ end }}(v {{ .ParamTypeRef }}) {{ .ResultTypeRef }} {
        {{ .Code }}
        return res
}
