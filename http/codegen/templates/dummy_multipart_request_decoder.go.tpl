{{ printf "%s reads the multipart request body for service %q endpoint %q into body." .FuncDeclaration.Name .ServiceName .MethodName | comment }}
func {{ .FuncDeclaration.Name }}(mr *multipart.Reader, body *{{ .BodyType }}) error {
	// Add multipart request decoder logic here
	return nil
}
