{{ printf "%s builds a %s from an error." .Declaration.Name .TypeName |  comment }}
func {{ .Declaration.Name }}(err error) {{ .TypeRef }} {
	return goa.NewServiceError(err, {{ printf "%q" .ErrName }}, {{ printf "%v" .Timeout }}, {{ printf "%v" .Temporary}}, {{ printf "%v" .Fault}})
}
