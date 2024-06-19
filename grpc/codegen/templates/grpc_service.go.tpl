
{{ .Description | comment }}
service {{ .Name }} {
	{{- range .Endpoints }}
	{{ if .Method.Description }}{{ .Method.Description | comment }}{{ end }}
	{{- $serverStream := or (eq .Method.StreamKind 3) (eq .Method.StreamKind 4) }}
	{{- $clientStream := or (eq .Method.StreamKind 2) (eq .Method.StreamKind 4) }}
	rpc {{ .Method.VarName }} ({{ if $clientStream }}stream {{ end }}{{ .Request.Message.VarName }}) returns ({{ if $serverStream }}stream {{ end }}{{ .Response.Message.VarName }});
	{{- end }}
}
