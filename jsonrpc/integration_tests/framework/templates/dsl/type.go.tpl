{{- /* Template for generating Goa DSL type expressions */ -}}
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
	{{- $required := collectRequired .Fields }}
	{{- if $required }}
	Required({{ range $i, $f := $required }}{{ if $i }}, {{ end }}"{{ $f }}"{{ end }})
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

{{- define "type" -}}
	{{- if eq .Kind "primitive" -}}
		{{- .Primitive -}}
	{{- else if eq .Kind "array" -}}
		{{- template "array_type" . -}}
	{{- else if eq .Kind "object" -}}
		{{- template "object_type" . -}}
	{{- else if eq .Kind "map" -}}
		{{- template "map_type" . -}}
	{{- else -}}
Any
	{{- end -}}
{{- end -}}

{{- define "array_type" -}}
ArrayOf({{ if .ArrayElem }}{{ template "type" .ArrayElem }}{{ else }}Any{{ end }})
{{- end -}}

{{- define "object_type" -}}
func() {
	{{- range .Fields }}
	Field({{ .Position }}, "{{ .Name }}", {{ template "type" .Type }}{{ if .Description }}, "{{ .Description }}"{{ end }})
	{{- end }}
}
{{- end -}}

{{- define "map_type" -}}
MapOf({{ if .MapKey }}{{ template "type" .MapKey }}{{ else }}String{{ end }}, {{ if .MapValue }}{{ template "type" .MapValue }}{{ else }}Any{{ end }})
{{- end -}}
