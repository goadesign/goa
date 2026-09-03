{{ printf "%s is the type to decode multipart request for the %q service %q endpoint." .FuncDeclaration.Name .ServiceName .MethodName | comment }}
type {{ .FuncDeclaration.Name }} func(*multipart.Reader, *{{ if .Payload.Request.ServerBody.Declaration }}{{ .Payload.Request.ServerBody.Declaration.Name }}{{ else }}{{ .Payload.Request.ServerBody.VarName }}{{ end }}) error
