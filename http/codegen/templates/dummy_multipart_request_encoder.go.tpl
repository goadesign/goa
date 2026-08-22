{{ printf "%s implements the multipart encoder for service %q endpoint %q." .FuncDeclaration.Name .ServiceName .MethodName | comment }}
func {{ .FuncDeclaration.Name }}(mw *multipart.Writer, p {{ .Payload.Ref }}) error {
	// Add multipart request encoder logic here
	return nil
}
