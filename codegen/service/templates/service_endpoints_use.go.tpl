

{{ printf "Use applies the given middleware to all the %q service endpoints." .Name | comment }}
func (e *{{ .VarName }}) Use(m func(goa.Endpoint) goa.Endpoint) {
{{- range .Methods }}
	e.{{ .VarName }} = m(e.{{ .VarName }})
{{- end }}
}