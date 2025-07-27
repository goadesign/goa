{{ printf "Request types for %s streaming operations" .Method.Name | comment }}
{{- if and .IsPayloadStreaming .IsResultStreaming }}
type {{ .VarName }}InitRequest struct {

}

type {{ .VarName }}SendRequest struct {
	StreamID string                              `json:"streamId"`

}

type {{ .VarName }}CloseSendRequest struct {
	StreamID string `json:"streamId"`
}
{{- end }}

{{- if .IsResultStreaming }}
type {{ .VarName }}RecvRequest struct {
	StreamID string `json:"streamId"`
}

{{- if not (and .IsPayloadStreaming .IsResultStreaming) }}
type {{ .VarName }}RecvResponse struct {
	Data {{ .RecvTypeRef }} `json:"data"`
}
{{- end }}

type {{ .VarName }}CloseRequest struct {
	StreamID string `json:"streamId"`
}
{{- end }}
