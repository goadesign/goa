{{ printf "%s holds both the result and the HTTP response body reader of the %q method." .ResponseStruct .Name | comment }}
type {{ .ResponseStruct }} struct {
{{- if .ResultRef }}
	{{ comment "Result is the method result." }}
	Result {{ .ResultRef }}
{{- end }}
	{{ comment "Body streams the HTTP response body." }}
	Body io.ReadCloser
}
