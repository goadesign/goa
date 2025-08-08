{{- /* Template for generating method signature */ -}}
{{- if .IsStreaming -}}
	{{- if .IsSSE -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context{{ if .HasPayload }}, p {{ .PayloadRef }}{{ end }}, stream {{ $.ServicePackage }}.{{ .StreamInterface }}) error
	{{- else if .IsWebSocket -}}
		{{- if .IsBidirectional -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context{{ if .HasPayload }}, p {{ .PayloadRef }}{{ end }}, stream {{ $.ServicePackage }}.{{ .StreamInterface }}) error
		{{- else if eq .StreamKind "payload" -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context, stream {{ $.ServicePackage }}.{{ .StreamInterface }}) error
		{{- else -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context{{ if .HasPayload }}, p {{ .PayloadRef }}{{ end }}, stream {{ $.ServicePackage }}.{{ .StreamInterface }}) error
		{{- end -}}
	{{- else -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context{{ if .HasPayload }}, p {{ .PayloadRef }}{{ end }}) {{ if .HasResult }}({{ .ResultRef }}, error){{ else }}error{{ end }}
	{{- end -}}
{{- else -}}
func (s *{{ $.ServicePackage }}srvc) {{ .GoName }}(ctx context.Context{{ if .HasPayload }}, p {{ .PayloadRef }}{{ end }}) {{ if .HasResult }}({{ .ResultRef }}, error){{ else }}error{{ end }}
{{- end -}}