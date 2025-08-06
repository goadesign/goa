// {{ .ServicePackage }}srvc implements the {{ .ServicePackage }} service.
type {{ .ServicePackage }}srvc struct {
	logger *log.Logger
{{- range .Methods }}
{{- if and (eq .Info.Action "collect") (eq .Info.Type "array") (eq .Transport "ws") }}
	// State for accumulating items in {{ .Name }}
	collectedItems []string
{{- end }}
{{- end }}
}

// New{{ .Title }} returns the {{ .ServicePackage }} service implementation.
func New{{ .Title }}() {{ .ServicePackage }}.Service {
	return &{{ .ServicePackage }}srvc{}
}
{{- if eq .Name "testws" }}

// HandleStream handles the JSON-RPC WebSocket streaming connection
func (s *{{ .ServicePackage }}srvc) HandleStream(ctx context.Context, stream {{ .ServicePackage }}.Stream) error {
	// For testing purposes, we only send broadcasts when explicitly called through the broadcast method
	// In a real application, you might send broadcasts based on external events or timers
	
	// Loop to handle incoming requests
	for {
		// Recv reads and dispatches the next request
		if err := stream.Recv(ctx); err != nil {
			// Check if it's a normal close or an error
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
{{- end }}
{{- range .Methods }}
{{- if or (not .IsNotification) (and .IsNotification .IsStreaming) }}

// {{ .GoName }} implements {{ .Name }}.
{{ template "partial_method_signature" . }} {
	log.Printf("{{ .GoName }} called")
{{- if .IsStreaming }}
{{- if .IsSSE }}
{{ template "partial_streaming_sse" . }}
{{- else if .IsWebSocket }}
{{ template "partial_streaming_websocket" . }}
{{- end }}
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
{{- end }}