package service

// data: Data
const dummyAuthFuncsT = `{{ range .Schemes }}
{{ printf "%sAuth implements the authorization logic for service %q for the %q security scheme." .Type $.Name .SchemeName | comment }}
func (s *{{ $.VarName }}srvc) {{ .Type }}Auth(ctx context.Context, {{ if eq .Type "Basic" }}user, pass{{ else if eq .Type "APIKey" }}key{{ else }}token{{ end }} string, scheme *security.{{ .Type }}Scheme) (context.Context, error) {
	//
	// TBD: add authorization logic.
	//
	// In case of authorization failure this function should return
	// one of the generated error structs, e.g.:
	//
	//    return ctx, myservice.MakeUnauthorizedError("invalid token")
	//
	// Alternatively this function may return an instance of
	// goa.ServiceError with a Name field value that matches one of
	// the design error names, e.g:
	//
	//    return ctx, goa.PermanentError("unauthorized", "invalid token")
	//
	return ctx, fmt.Errorf("not implemented")
}
{{- end }}
`
