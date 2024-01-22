{{ printf "%s implements the multipart decoder for service %q endpoint %q. The decoder must populate the argument p after encoding." .FuncName .ServiceName .MethodName | comment }}
func {{ .FuncName }}(mr *multipart.Reader, p *{{ .Payload.Ref }}) error {
	// Add multipart request decoder logic here
	return nil
}
