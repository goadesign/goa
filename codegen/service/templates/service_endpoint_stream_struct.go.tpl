

{{ printf "%s holds both the payload and the server stream of the %q method." .ServerStream.EndpointStruct .Name | comment }}
type {{ .ServerStream.EndpointStruct }} struct {
{{- if .PayloadRef }}
	{{ comment "Payload is the method payload." }}
	Payload {{ .PayloadRef }}
{{- end }}
{{- if .IsJSONRPC }}
	{{ comment "RequestID is the JSON-RPC request ID (available for JSON-RPC transports)." }}
	RequestID any
{{- end }}
	{{ printf "Stream is the server stream used by the %q method to send data." .Name | comment }}
	Stream {{ .ServerStream.Interface }}
}
