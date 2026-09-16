// This file keeps released gRPC generator entry points available to plugins
// while all rendering uses the one service plan retained by Goa.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/codegen"
)

// ClientFiles returns the planned client files. genpkg must match the package
// used to create services.
func ClientFiles(genpkg string, services *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, services)
	return clientFiles(services)
}

// ClientCLIFiles returns the planned command-line client files. genpkg must
// match the package used to create services.
func ClientCLIFiles(genpkg string, services *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, services)
	return clientCLIFiles(services)
}

// ProtoFiles returns the planned protobuf files. genpkg must match the package
// used to create services.
func ProtoFiles(genpkg string, services *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, services)
	return protoFiles(services)
}

// ServerFiles returns the planned server files. genpkg must match the package
// used to create services.
func ServerFiles(genpkg string, services *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, services)
	return serverFiles(services)
}

// ServerTypeFiles returns the planned server conversion files. genpkg must
// match the package used to create services.
func ServerTypeFiles(genpkg string, services *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, services)
	return serverTypeFiles(services)
}

// ClientTypeFiles returns the planned client conversion files. genpkg must
// match the package used to create services.
func ClientTypeFiles(genpkg string, services *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, services)
	return clientTypeFiles(services)
}

// requireGeneratedPackage rejects a package argument that does not describe
// the service data supplied by the same generation run.
func requireGeneratedPackage(genpkg string, services *ServicesData) {
	if genpkg != services.GenPkg() {
		panic(fmt.Sprintf("gRPC generation package %q does not match planned package %q", genpkg, services.GenPkg()))
	}
}
