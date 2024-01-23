{{ printf "%s implements the multipart encoder for service %q endpoint %q." .FuncName .ServiceName .MethodName | comment }}
func {{ .FuncName }}(mw *multipart.Writer, p {{ .Payload.Ref }}) error {
	// Add multipart request encoder logic here
	return nil
}
