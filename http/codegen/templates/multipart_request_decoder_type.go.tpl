{{ printf "%s is the type to decode multipart request for the %q service %q endpoint." .FuncName .ServiceName .MethodName | comment }}
type {{ .FuncName }} func(*multipart.Reader, *{{ .Payload.Ref }}) error
