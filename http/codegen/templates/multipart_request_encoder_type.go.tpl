{{ printf "%s is the type to encode multipart request for the %q service %q endpoint." .FuncName .ServiceName .MethodName | comment }}
type {{ .FuncName }} func(*multipart.Writer, {{ .Payload.Ref }}) error
