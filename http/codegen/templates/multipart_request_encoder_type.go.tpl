{{ printf "%s is the type to encode multipart request for the %q service %q endpoint." .FuncDeclaration.Name .ServiceName .MethodName | comment }}
type {{ .FuncDeclaration.Name }} func(*multipart.Writer, {{ .Payload.Ref }}) error
