
{{ comment .Description }}
type Service interface {
{{- if hasJSONRPCWebSocket . }}
	{{ comment "HandleStream handles the JSON-RPC WebSocket connection. Calling Recv() on the stream will dispatch the request to the appropriate method below." }}
	HandleStream(context.Context, Stream) error
{{- end }}
{{- range .Methods }}
	{{ comment .Description }}
	{{- if .SkipResponseBodyEncodeDecode }}
	{{ comment "\nIf body implements [io.WriterTo], that implementation will be used instead. Consider [goa.design/goa/v3/pkg.SkipResponseWriter] to adapt existing implementations." }}
	{{- end }}
	{{- if .ViewedResult }}
		{{- if not .ViewedResult.ViewName }}
			{{ comment "The \"view\" return value must have one of the following views" }}
			{{- range .ViewedResult.Views }}
				{{- if .Description }}
					{{ printf "//	- %q: %s" .Name .Description }}
				{{- else }}
					{{ printf "//	- %q" .Name }}
				{{- end }}
			{{- end }}
		{{- end }}
	{{- end }}
	{{- if and .ServerStream (not .IsJSONRPC) }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}, {{ .ServerStream.Interface }}) (err error)
	{{- else }}
		{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}{{ if .SkipRequestBodyEncodeDecode }}, io.ReadCloser{{ end }}) ({{ if .Result }}res {{ .ResultRef }}, {{ end }}{{ if .SkipResponseBodyEncodeDecode }}body io.ReadCloser, {{ end }}{{ if .Result }}{{ if .ViewedResult }}{{ if not .ViewedResult.ViewName }}view string, {{ end }}{{ end }}{{ end }}err error)
	{{- end }}
{{- end }}
}

{{- if .Schemes }}
// Auther defines the authorization functions to be implemented by the service.
type Auther interface {
	{{- range .Schemes.DedupeByType }}
	{{ printf "%sAuth implements the authorization logic for the %s security scheme." .Type .Type | comment }}
	{{ .Type }}Auth(ctx context.Context, {{ if eq .Type "Basic" }}user, pass{{ else if eq .Type "APIKey" }}key{{ else }}token{{ end }} string, schema *security.{{ .Type }}Scheme) (context.Context, error)
	{{- end }}
}
{{- end }}

{{- range .Methods }}
	{{- if .ServerStream }}
		{{ template "stream_interface" (streamInterfaceFor "server" . .ServerStream) }}
		{{ template "stream_interface" (streamInterfaceFor "client" . .ClientStream) }}
	{{- end }}
{{- end }}

{{- if hasJSONRPCWebSocket . }}
	{{ template "jsonrpc_websocket_stream" . }}
{{- end }}

// APIName is the name of the API as defined in the design.
const APIName = {{ printf "%q" .APIName }}

// APIVersion is the version of the API as defined in the design.
const APIVersion = {{ printf "%q" .APIVersion }}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = {{ printf "%q" .Name }}

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [{{ len .Methods }}]string{ {{ range .Methods }}{{ printf "%q" .Name }}, {{ end }} }

{{- define "stream_interface" }}
{{ printf "%s is the interface a %q endpoint %s stream must satisfy." .Stream.Interface .Endpoint .Type | comment }}
type {{ .Stream.Interface }} interface {
	{{- if .Stream.SendTypeRef }}
		{{ comment .Stream.SendDesc }}
		{{ .Stream.SendName }}({{ .Stream.SendTypeRef }}) error
		{{ comment .Stream.SendWithContextDesc }}
		{{ .Stream.SendWithContextName }}(context.Context, {{ .Stream.SendTypeRef }}) error
	{{- end }}
	{{- if .Stream.RecvTypeRef }}
		{{ comment .Stream.RecvDesc }}
		{{ .Stream.RecvName }}() ({{ .Stream.RecvTypeRef }}, error)
		{{ comment .Stream.RecvWithContextDesc }}
		{{ .Stream.RecvWithContextName }}(context.Context) ({{ .Stream.RecvTypeRef }}, error)
	{{- end }}
	{{- if .Stream.MustClose }}
		{{ comment "Close closes the stream." }}
		Close() error
	{{- end }}
	{{- if and .IsViewedResult (eq .Type "server") }}
		{{ comment "SetView sets the view used to render the result before streaming." }}
		SetView(view string)
	{{- end }}
}
{{- end }}

{{ define "jsonrpc_websocket_stream" }}
{{ printf "Stream defines the interface for managing a streaming WebSocket connection in the %s server.  It allows sending results, sending errors, receiving requests, and closing the connection.  This interface is used by the service to interact with clients over WebSocket using JSON-RPC." .Name | comment }}
type Stream interface {
{{- $hasErrors := false }}
{{- range .Methods }}
	{{- if .Result }}
	{{ printf "Send%s sends a JSON-RPC response for the %s method." .VarName .Name | comment }}
	Send{{ .VarName }}(ctx context.Context, result {{ .ResultRef }}) error
	{{- end }}
	{{- if .Errors }}{{ $hasErrors = true }}{{ end }}
{{- end }}
{{- if $hasErrors }}
	// SendError sends a JSON-RPC error response.
	SendError(ctx context.Context, id string, err error) error
{{- end }}
{{- $hasStreamingPayload := false }}
{{- range .Methods }}
	{{- if .StreamingPayload }}{{ $hasStreamingPayload = true }}{{ end }}
{{- end }}
{{- if $hasStreamingPayload }}
	{{ printf "Recv reads JSON-RPC requests from the %s service stream and dispatches them to the appropriate method." .Name | comment }}
	Recv(ctx context.Context) error
{{- end }}
	{{ printf "Close closes the %s service websocket connection." .Name | comment }}
	Close() error
} 
{{ end }}