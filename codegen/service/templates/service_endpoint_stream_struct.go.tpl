

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
	{{- if .IsJSONRPC }}
	{{ comment "For JSON-RPC transports, this will include SendNotification and SendResponse methods." }}
	Stream interface {
		{{ .ServerStream.Interface }}
		Send{{ .VarName }}Notification(ctx context.Context, result {{ .ServerStream.SendTypeRef }}) error
		Send{{ .VarName }}Response(ctx context.Context, result {{ .ServerStream.SendTypeRef }}) error
	}
	{{- else }}
	Stream {{ .ServerStream.Interface }}
	{{- end }}
}
