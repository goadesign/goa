{{ printf "%s holds information about the mounted endpoints." .MountPointStructDeclaration.Name | comment }}
type {{ .MountPointStructDeclaration.Name }} struct {
	{{ printf "Method is the name of the service method served by the mounted HTTP handler." | comment }}
	Method string
	{{ printf "Verb is the HTTP method used to match requests to the mounted handler." | comment }}
	Verb string
	{{ printf "Pattern is the HTTP request path pattern used to match requests to the mounted handler." | comment }}
	Pattern string
}
