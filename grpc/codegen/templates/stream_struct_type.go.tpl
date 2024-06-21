{{ printf "%s implements the %s interface." .VarName .ServiceInterface | comment }}
type {{ .VarName }} struct {
	stream {{ .Interface }}
{{- if .Endpoint.Method.ViewedResult }}
	view string
{{- end }}
}
