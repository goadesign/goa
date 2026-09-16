{{ printf "%s lists the service endpoint gRPC clients." .ClientStructDeclaration.Name | comment }}
type {{ .ClientStructDeclaration.Name }} struct {
	grpccli {{ .ClientProtobufPkgName }}.{{ .ClientInterface }}
	opts []grpc.CallOption
}
