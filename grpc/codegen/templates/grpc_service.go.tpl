
{{ .Description | comment }}
service {{ .Name }} {
	{{- range .Endpoints }}
	{{ if .Method.Description }}{{ .Method.Description | comment }}{{ end }}
	{{- $serverStream := or (eq .Method.StreamKind 3) (eq .Method.StreamKind 4) }}
	{{- $clientStream := or (eq .Method.StreamKind 2) (eq .Method.StreamKind 4) }}
	rpc {{ .ProtoMethodName }} ({{ if $clientStream }}stream {{ end }}{{ .Request.ProtoMessageName }}) returns ({{ if $serverStream }}stream {{ end }}{{ .Response.ProtoMessageName }}){{ if .Method.Idempotent }} {
		option idempotency_level = IDEMPOTENT;
	}{{ else }};{{ end }}
	{{- end }}
}
