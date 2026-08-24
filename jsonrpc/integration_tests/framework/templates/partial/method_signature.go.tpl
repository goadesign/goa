{{- /* Template for generating method signature */ -}}
{{- if .IsStreaming -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context{{ if .HasPayload }}, p {{ .PayloadRef }}{{ end }}, stream {{ $.ServicePackage }}.{{ .StreamInterface }}) error
{{- else -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context{{ if .HasPayload }}, p {{ .PayloadRef }}{{ end }}) {{ if .HasResult }}({{ .ResultRef }}, error){{ else }}error{{ end }}
{{- end -}}
