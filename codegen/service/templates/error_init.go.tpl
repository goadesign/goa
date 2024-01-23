{{ printf "%s builds a %s from an error." .Name .TypeName |  comment }}
func {{ .Name }}(err error) {{ .TypeRef }} {
	return goa.NewServiceError(err, {{ printf "%q" .ErrName }}, {{ printf "%v" .Timeout }}, {{ printf "%v" .Temporary}}, {{ printf "%v" .Fault}})
}
