{{ printf "%s instantiates gRPC client for all the %s service servers." .ClientInitDeclaration.Name .Service.Name | comment }}
func {{ .ClientInitDeclaration.Name }}(cc *grpc.ClientConn, opts ...grpc.CallOption) *{{ .ClientStructDeclaration.Name }} {
  return &{{ .ClientStructDeclaration.Name }}{
		grpccli: {{ .ClientInterfaceInit }}(cc),
		opts: opts,
	}
}
