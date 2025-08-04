{{- /* Template for generating a Goa method definition */ -}}
Method("{{ .Name }}", func() {
	Description("{{ .Description }}")
{{- if .Payload }}
	Payload({{ template "inline_type" .Payload }})
{{- end }}
{{- if .StreamingPayload }}
	StreamingPayload({{ template "inline_type" .StreamingPayload }})
{{- end }}
{{- if .StreamingResult }}
	StreamingResult({{ template "inline_type" .StreamingResult }})
{{- end }}
{{- if and .Result (not .IsNotification) }}
	Result({{ template "inline_type" .Result }})
{{- end }}
{{- if .ReturnsError }}
	Error("test_error")
{{- end }}
	JSONRPC(func() {
{{- if .IsSSE }}
		ServerSentEvents()
{{- else if and .IsWebSocket .IsBidirectional }}
		ID("id")
{{- end }}
	})
})

{{- define "inline_type" -}}
{{- if eq .Kind "primitive" -}}
{{- if .Validations -}}
func() {
	Attribute({{ .Primitive }})
{{- range .Validations }}
{{- if eq .Type "MinLength" }}
	MinLength({{ .Value }})
{{- else if eq .Type "MaxLength" }}
	MaxLength({{ .Value }})
{{- end }}
{{- end }}
}
{{- else -}}
{{- .Primitive -}}
{{- end -}}
{{- else if eq .Kind "array" -}}
{{- if and .ArrayElem (eq .ArrayElem.Kind "primitive") -}}
ArrayOf({{ .ArrayElem.Primitive }})
{{- else -}}
func() {
	Field(1, "items", ArrayOf({{ if .ArrayElem }}{{ template "inline_type" .ArrayElem }}{{ else }}String{{ end }}))
	Required("items")
}
{{- end -}}
{{- else if eq .Kind "object" -}}
func() {
{{- range .Fields }}
{{- $fieldName := .Name }}
	Field({{ .Position }}, "{{ .Name }}", {{ template "inline_type" .Type }}{{ if .Description }}, "{{ .Description }}"{{ end }})
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
}
{{- else if eq .Kind "map" -}}
func() {
	Field(1, "data", MapOf({{ if .MapKey }}{{ template "inline_type" .MapKey }}{{ else }}String{{ end }}, {{ if .MapValue }}{{ template "inline_type" .MapValue }}{{ else }}Any{{ end }}))
	Required("data")
}
{{- else -}}
Any
{{- end -}}
{{- end -}}