

{{ printf "%s holds both the payload and the HTTP request body reader of the %q method." .RequestStruct .Name | comment }}
type {{ .RequestStruct }} struct {
{{- if .PayloadRef }}
	{{ comment "Payload is the method payload." }}
	Payload {{ .PayloadRef }}
{{- end }}
	{{ comment "Body streams the HTTP request body." }}
	Body io.ReadCloser
}
