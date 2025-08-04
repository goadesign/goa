
{{ comment .Description }}
type Service interface {
{{- if isJSONRPCWebSocket . }}
	{{ comment "HandleStream handles the JSON-RPC WebSocket streaming connection. Calling Recv() on the stream will dispatch requests to the appropriate methods below." }}
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
	{{- if .ServerStream }}
		{{- if and .IsJSONRPC (not .IsJSONRPCSSE) (eq .ServerStream.Kind 2) }}
			{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}) ({{ if .Result }}res {{ .ResultRef }}, {{ end }}err error)
		{{- else }}
			{{- if and .IsJSONRPC (not .IsJSONRPCSSE) (eq .ServerStream.Kind 3) .PayloadRef }}
				{{- /* JSON-RPC WebSocket server streaming with non-streaming payload */ -}}
				{{ .VarName }}(context.Context, {{ .PayloadRef }}, {{ .ServerStream.Interface }}) (err error)
			{{- else }}
				{{ .VarName }}(context.Context{{ if .Payload }}, {{ .PayloadRef }}{{ end }}, {{ .ServerStream.Interface }}) (err error)
			{{- end }}
		{{- end }}
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

{{- range .Methods }}
	{{- if .ServerStream }}
		{{ template "stream_interface" (streamInterfaceFor "server" . .ServerStream) }}
		{{ template "stream_interface" (streamInterfaceFor "client" . .ClientStream) }}
	{{- end }}
{{- end }}

{{- if hasJSONRPCStreaming . }}
	{{- if isJSONRPCWebSocket . }}
	{{ template "jsonrpc_websocket_stream" . }}
	{{- else }}
	{{ template "jsonrpc_sse_stream" . }}
	{{- end }}
{{- end }}

{{- define "stream_interface" }}
{{- if and .IsJSONRPCSSE (eq .Type "server") }}
{{ printf "%sEvent is the interface implemented by the result type for the %s method." .MethodVarName .Endpoint | comment }}
type {{ .MethodVarName }}Event interface {
	is{{ .MethodVarName }}Event()
}

{{ printf "is%sEvent implements the %sEvent interface." .MethodVarName .MethodVarName | comment }}
func ({{ .Stream.SendTypeRef }}) is{{ .MethodVarName }}Event() {}

{{ printf "%s is the interface a %q endpoint %s stream must satisfy for JSON-RPC SSE." .Stream.Interface .Endpoint .Type | comment }}
type {{ .Stream.Interface }} interface {
	{{- if .Stream.SendTypeRef }}
	{{ comment "Send sends an event (notification or response) to the client." }}
	{{ comment "For notifications, the result should not have an ID field." }}
	{{ comment "For responses, the result must have an ID field." }}
	Send(ctx context.Context, event {{ .MethodVarName }}Event) error
	{{- end }}
	{{ comment "SendError sends a JSON-RPC error response." }}
	SendError(ctx context.Context, id string, err error) error
}
{{- else }}
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
{{- end }}

{{- define "jsonrpc_websocket_stream" }}
{{ printf "Stream defines the interface for managing a WebSocket streaming connection in the %s server. It allows sending results, sending errors, receiving requests, and closing the connection. This interface is used by the service to interact with clients over WebSocket using JSON-RPC." .Name | comment }}
type Stream interface {
{{- $hasErrors := false }}
{{- $hasResults := false }}
{{- $resultTypes := "" }}
{{- range .Methods }}
	{{- if .Result }}
		{{- $hasResults = true }}
		{{- if $resultTypes }}
			{{- $resultTypes = printf "%s, %s" $resultTypes .ResultRef }}
		{{- else }}
			{{- $resultTypes = .ResultRef }}
		{{- end }}
	{{- end }}
	{{- if .Errors }}{{ $hasErrors = true }}{{ end }}
{{- end }}
{{- if $hasResults }}
	// Send sends an event to the client.
	// Accepted types: {{ $resultTypes }}
	Send(Event) error
{{- end }}
{{- if $hasErrors }}
	// SendError sends a JSON-RPC error response.
	SendError(ctx context.Context, id string, err error) error
{{- end }}
	{{ printf "Recv reads JSON-RPC requests from the %s service WebSocket stream and dispatches them to the appropriate method." .Name | comment }}
	Recv(ctx context.Context) error
}

{{- if $hasResults }}
{{ printf "Event is the interface implemented by all result types that can be sent via the %s Stream." .Name | comment }}
type Event interface {
	is{{ .VarName }}Event()
}

	{{- range .Methods }}
		{{- if .Result }}
// is{{ $.VarName }}Event implements the Event interface.
func ({{ .ResultRef }}) is{{ $.VarName }}Event() {}
		{{- end }}
	{{- end }}
{{- end }}
{{- end }}

{{- define "jsonrpc_sse_stream" }}
{{- $hasResults := false }}
{{- $hasErrors := false }}
{{- $resultTypes := "" }}
{{- range .Methods }}
	{{- if .Result }}
		{{- $hasResults = true }}
		{{- if $resultTypes }}
			{{- $resultTypes = printf "%s, %s" $resultTypes .ResultRef }}
		{{- else }}
			{{- $resultTypes = .ResultRef }}
		{{- end }}
	{{- end }}
	{{- if .Errors }}{{ $hasErrors = true }}{{ end }}
{{- end }}
{{ printf "Stream defines the interface for managing an SSE streaming connection in the %s server. It allows sending notifications and final responses. This interface is used by the service to interact with clients over SSE using JSON-RPC." .Name | comment }}
type Stream interface {
{{- if $hasResults }}
	{{ comment "Send sends an event (notification or response) to the client." }}
	{{ comment "For notifications, the result should not have an ID field." }}
	{{ comment "For responses, the result must have an ID field." }}
	{{ printf "Accepted types: %s" $resultTypes | comment }}
	Send(ctx context.Context, event Event) error
{{- end }}
{{- if $hasErrors }}
	{{ comment "SendError sends a JSON-RPC error response." }}
	SendError(ctx context.Context, id string, err error) error
{{- end }}
}

{{- if $hasResults }}
{{ printf "Event is the interface implemented by all result types that can be sent via the %s Stream." .Name | comment }}
type Event interface {
	is{{ .VarName }}Event()
}

	{{- range .Methods }}
		{{- if .Result }}
{{ printf "is%sEvent implements the Event interface." $.VarName | comment }}
func ({{ .ResultRef }}) is{{ $.VarName }}Event() {}
		{{- end }}
	{{- end }}
{{- end }}
{{- end }}
