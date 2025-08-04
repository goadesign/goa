{{- if eq .Kind "primitive" -}}
{{- .Primitive -}}
{{- else if eq .Kind "array" -}}
{{- if and .ArrayElem (eq .ArrayElem.Kind "primitive") -}}
ArrayOf({{ .ArrayElem.Primitive }})
{{- else -}}
func() {
	Field(1, "items", ArrayOf({{ if .ArrayElem }}{{ template "partial_type" .ArrayElem }}{{ else }}String{{ end }}))
	Required("items")
}
{{- end -}}
{{- else if eq .Kind "object" -}}
func() {
{{- range .Fields }}
	Field({{ .Position }}, "{{ .Name }}", {{ template "partial_type" .Type }}{{ if .Description }}, "{{ .Description }}"{{ end }})
{{- end }}
{{- if .Validations }}
{{- range .Validations }}
{{- if eq .Type "MinLength" }}
	MinLength({{ .Value }})
{{- else if eq .Type "MaxLength" }}
	MaxLength({{ .Value }})
{{- end }}
{{- end }}
{{- end }}
{{- $required := collectRequired .Fields }}
{{- if $required }}
	Required({{ range $i, $f := $required }}{{ if $i }}, {{ end }}"{{ $f }}"{{ end }})
{{- end }}
{{- if .NeedsID }}
	ID("id")
{{- end }}
}
{{- else if eq .Kind "map" -}}
func() {
	Field(1, "data", MapOf({{ if .MapKey }}{{ template "partial_type" .MapKey }}{{ else }}String{{ end }}, {{ if .MapValue }}{{ template "partial_type" .MapValue }}{{ else }}Any{{ end }}))
	Required("data")
}
{{- else -}}
Any
{{- end -}}