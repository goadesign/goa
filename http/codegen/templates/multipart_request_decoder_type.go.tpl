{{ printf "%s is the type to decode multipart request for the %q service %q endpoint." .FuncDeclaration.Name .ServiceName .MethodName | comment }}
type {{ .FuncDeclaration.Name }} func(*multipart.Reader, *{{ .Payload.Ref }}) error
