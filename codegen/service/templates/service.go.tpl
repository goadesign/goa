
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
		{{- /* Emit client stream interface */ -}}
		{{- if .IsJSONRPC }}
			{{- if .ClientStream }}
			{{ template "stream_interface" (streamInterfaceFor "client" . .ClientStream) }}
			{{- end }}
		{{- else }}
		{{ template "stream_interface" (streamInterfaceFor "client" . .ClientStream) }}
		{{- end }}
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

{{ printf "%s allows streaming instances of %s over SSE." .Stream.Interface .Stream.SendTypeRef | comment }}
type {{ .Stream.Interface }} interface {
	{{- if .Stream.SendTypeRef }}
	{{ comment .Stream.SendDesc }}
	{{ comment "IMPORTANT: Send only sends JSON-RPC notifications. Use SendAndClose to send a final response." }}
	Send(ctx context.Context, event {{ .MethodVarName }}Event) error
		{{- if .Stream.SendAndCloseName }}
	{{ comment .Stream.SendAndCloseDesc }}
	{{ comment "The result will be sent as a JSON-RPC response with the original request ID." }}
	{{ comment "If the result has an ID field populated, that ID will be used instead of the request ID." }}
	{{ .Stream.SendAndCloseName }}(ctx context.Context, event {{ .MethodVarName }}Event) error
		{{- end }}
	{{- end }}
	{{ comment "SendError sends a JSON-RPC error response." }}
	SendError(ctx context.Context, id string, err error) error
}
{{- else }}
{{- $elemType := .Stream.SendTypeRef -}}
{{- if not $elemType }}{{- $elemType = .Stream.RecvTypeRef }}{{- end }}
{{ printf "%s allows streaming instances of %s to the client." .Stream.Interface $elemType | comment }}
type {{ .Stream.Interface }} interface {
	{{- if .Stream.SendTypeRef }}
		{{- if .IsJSONRPCWebSocket }}
		{{ comment "SendNotification sends a JSON-RPC notification (no response expected)." }}
		SendNotification(context.Context, {{ .Stream.SendTypeRef }}) error
		{{ comment "SendResponse sends a JSON-RPC response with the original request ID." }}
		SendResponse(context.Context, {{ .Stream.SendTypeRef }}) error
		{{ comment "SendError sends a JSON-RPC error response." }}
		SendError(context.Context, error) error
		{{- else }}
		{{ comment .Stream.SendDesc }}
		{{ .Stream.SendName }}({{ .Stream.SendTypeRef }}) error
                {{ .Stream.SendWithContextName }}(context.Context, {{ .Stream.SendTypeRef }}) error
		{{- end }}
	{{- end }}
	{{- if and .Stream.RecvTypeRef (not .IsJSONRPCWebSocket) }}
		{{ .Stream.RecvName }}() ({{ .Stream.RecvTypeRef }}, error)
		{{ .Stream.RecvWithContextName }}(context.Context) ({{ .Stream.RecvTypeRef }}, error)
	{{- end }}

	{{- if .IsJSONRPCWebSocket }}
	{{ comment "Close closes the stream." }}
	Close() error
	{{- else if .Stream.MustClose }}
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
{{- range .Methods }}
	{{- if .Result }}
	{{ printf "Send%sNotification sends a JSON-RPC notification for the %s method (no response expected)." .VarName .Name | comment }}
	Send{{ .VarName }}Notification(ctx context.Context, result {{ .ResultRef }}) error
	{{ printf "Send%sResponse sends a JSON-RPC response for the %s method with the given ID." .VarName .Name | comment }}
	Send{{ .VarName }}Response(ctx context.Context, id any, result {{ .ResultRef }}) error
	{{- end }}
{{- end }}
	{{ comment "SendError sends a JSON-RPC error response." }}
	SendError(ctx context.Context, id any, err error) error
	{{ printf "Recv reads JSON-RPC requests from the %s service WebSocket stream and dispatches them to the appropriate method." .Name | comment }}
	Recv(ctx context.Context) error
	{{ comment "Close closes the stream." }}
	Close() error
}
{{- end }}

{{- define "jsonrpc_sse_stream" }}
{{- $hasResults := false }}
{{- $hasErrors := false }}
{{- $resultTypes := "" }}
{{- range (dedupeByResult .Methods) }}
	{{- if .Result }}
		{{- $hasResults = true }}
		{{- if $resultTypes }}
			{{- $resultTypes = printf "%s, %s" $resultTypes .ResultRef }}
		{{- else }}
			{{- $resultTypes = .ResultRef }}
		{{- end }}
	{{- end }}
{{- end }}
{{- range .Methods }}
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

    {{- range (dedupeByResult .Methods) }}
        {{- if .Result }}
{{ printf "is%sEvent implements the Event interface." $.VarName | comment }}
func ({{ .ResultRef }}) is{{ $.VarName }}Event() {}
        {{- end }}
    {{- end }}
{{- end }}
{{- end }}
