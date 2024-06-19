{{ printf "%s lists the service endpoint gRPC clients." .ClientStruct | comment }}
type {{ .ClientStruct }} struct {
	grpccli {{ .PkgName }}.{{ .ClientInterface }}
	opts []grpc.CallOption
}