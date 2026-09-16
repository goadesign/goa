// This file keeps released HTTP generator entry points available to plugins
// while all rendering uses the one transport plan retained by Goa.
package codegen

import (
	"fmt"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// ClientFiles returns the planned client files. genpkg must match the package
// used to create data.
func ClientFiles(genpkg string, data *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, data)
	return clientFiles(data)
}

// ServerFiles returns the planned server files. genpkg must match the package
// used to create data.
func ServerFiles(genpkg string, data *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, data)
	return serverFiles(data)
}

// ServerTypeFiles returns the planned server type files. genpkg must match the
// package used to create data.
func ServerTypeFiles(genpkg string, data *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, data)
	return serverTypeFiles(data)
}

// ClientTypeFiles returns the planned client type files. genpkg must match the
// package used to create data.
func ClientTypeFiles(genpkg string, data *ServicesData) []*codegen.File {
	requireGeneratedPackage(genpkg, data)
	return clientTypeFiles(data)
}

// PathFiles returns the planned request path files.
func PathFiles(data *ServicesData) []*codegen.File {
	return pathFiles(data)
}

// ClientEncodeDecodeFile returns the planned client encoder and decoder file
// for service. genpkg must match the package used to create data.
func ClientEncodeDecodeFile(genpkg string, service *expr.HTTPServiceExpr, data *ServicesData) *codegen.File {
	requireGeneratedPackage(genpkg, data)
	return clientEncodeDecodeFile(service, data)
}

// ServerEncodeDecodeFile returns the planned server encoder and decoder file
// for service. genpkg must match the package used to create data.
func ServerEncodeDecodeFile(genpkg string, service *expr.HTTPServiceExpr, data *ServicesData) *codegen.File {
	requireGeneratedPackage(genpkg, data)
	return serverEncodeDecodeFile(service, data)
}

// WebsocketClientFile returns the planned WebSocket client file for service.
// genpkg must match the package used to create data.
func WebsocketClientFile(genpkg string, service *expr.HTTPServiceExpr, data *ServicesData) *codegen.File {
	requireGeneratedPackage(genpkg, data)
	return websocketClientFile(service, data)
}

// requireGeneratedPackage rejects a package argument that does not describe
// the HTTP data supplied by the same generation run.
func requireGeneratedPackage(genpkg string, data *ServicesData) {
	if genpkg != data.GenPkg() {
		panic(fmt.Sprintf("HTTP generation package %q does not match planned package %q", genpkg, data.GenPkg()))
	}
}
