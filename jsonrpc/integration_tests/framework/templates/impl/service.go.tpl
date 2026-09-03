// {{ .ServicePackage }}srvc implements the {{ .ServicePackage }} service.
type {{ .ServicePackage }}srvc struct {
	logger *log.Logger
}

// New{{ .Title }} returns the {{ .ServicePackage }} service implementation.
func New{{ .Title }}() {{ .ServicePackage }}.Service {
	return &{{ .ServicePackage }}srvc{}
}
{{- range .Methods }}

// {{ .GoName }} implements {{ .Name }}.
{{ template "partial_method_signature" . }} {
	log.Printf("{{ .GoName }} called")
{{- if .IsStreaming }}
{{ template "partial_streaming_sse" . }}
{{- else if .IsNotification }}
{{ template "partial_notify" . }}
{{- else if .ReturnsError }}
{{ template "partial_error" . }}
{{- else }}
{{- if eq .Info.Action "echo" }}
{{ template "partial_echo" . }}
{{- else if eq .Info.Action "transform" }}
{{ template "partial_transform" . }}
{{- else if eq .Info.Action "generate" }}
{{ template "partial_generate" . }}
{{- else }}
	// Unknown action: {{ .Info.Action }}
	{{- if .IsStreaming }}
	return fmt.Errorf("not implemented")
	{{- else }}
	return {{ if .HasResult }}nil, {{ end }}fmt.Errorf("not implemented")
	{{- end }}
{{- end }}
{{- end }}
}
{{- end }}
