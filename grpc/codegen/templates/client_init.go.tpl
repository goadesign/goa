{{ printf "New%s instantiates gRPC client for all the %s service servers." .ClientStruct .Service.Name | comment }}
func New{{ .ClientStruct }}(cc *grpc.ClientConn, opts ...grpc.CallOption) *{{ .ClientStruct }} {
  return &{{ .ClientStruct }}{
		grpccli: {{ .ClientInterfaceInit }}(cc),
		opts: opts,
	}
}
