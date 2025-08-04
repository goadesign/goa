{{- /* Template for generating Goa DSL type expressions */ -}}
{{- if eq .Kind "primitive" -}}
	{{- .Primitive -}}
{{- else if eq .Kind "array" -}}
	{{- if .ArrayElem -}}
		{{- if eq .ArrayElem.Kind "primitive" -}}
ArrayOf({{ template "type" .ArrayElem }})
		{{- else -}}
ArrayOf(func() {
{{ template "type" .ArrayElem | indent 1 }}
})
		{{- end -}}
	{{- else -}}
ArrayOf(Any)
	{{- end -}}
{{- else if eq .Kind "object" -}}
func() {
	{{- range .Fields }}
	Field({{ .Position }}, "{{ .Name }}", {{ template "type" .Type }}{{ if .Description }}, "{{ .Description }}"{{ end }})
	{{- end }}
	{{- if .Validations }}
		{{- range .Validations }}
			{{- if eq .Type "MinLength" }}
	MinLength({{ .Value }})
			{{- else if eq .Type "MaxLength" }}
	MaxLength({{ .Value }})
			{{- else if eq .Type "Pattern" }}
	Pattern({{ printf "%q" .Value }})
			{{- end }}
		{{- end }}
	{{- end }}
	{{- $required := collectRequired .Fields }}
	{{- if $required }}
	Required({{ range $i, $f := $required }}{{ if $i }}, {{ end }}"{{ $f }}"{{ end }})
	{{- end }}
}
{{- else if eq .Kind "map" -}}
	{{- if and .MapKey .MapValue -}}
MapOf({{ template "type" .MapKey }}, {{ template "type" .MapValue }})
	{{- else -}}
MapOf(String, Any)
	{{- end -}}
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