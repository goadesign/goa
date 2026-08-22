// This file requests the Go names written by generated HTTP client and server
// files. NewPlans calls it before names are assigned, and Link later gives the
// same records to every definition and use.
package codegen

import (
	"cmp"
	"strconv"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// httpSymbols contains every package-level name emitted for one HTTP service.
	httpSymbols struct {
		serverStruct         *codegen.NameDeclaration
		mountPoint           *codegen.NameDeclaration
		serverInit           *codegen.NameDeclaration
		mountServer          *codegen.NameDeclaration
		clientStruct         *codegen.NameDeclaration
		clientInit           *codegen.NameDeclaration
		serverConfigurer     *codegen.NameDeclaration
		serverConfigurerInit *codegen.NameDeclaration
		clientConfigurer     *codegen.NameDeclaration
		clientConfigurerInit *codegen.NameDeclaration
		appendFS             *codegen.NameDeclaration
		appendPrefix         *codegen.NameDeclaration
		endpoints            map[*expr.HTTPEndpointExpr]*httpEndpointSymbols
		fileServers          map[*expr.HTTPFileServerExpr]*codegen.NameDeclaration
	}

	// httpEndpointSymbols contains the package-level names emitted for one endpoint.
	httpEndpointSymbols struct {
		mountHandler       *codegen.NameDeclaration
		handlerInit        *codegen.NameDeclaration
		requestDecoder     *codegen.NameDeclaration
		responseEncoder    *codegen.NameDeclaration
		errorEncoder       *codegen.NameDeclaration
		discardStream      *codegen.NameDeclaration
		requestEncoder     *codegen.NameDeclaration
		responseDecoder    *codegen.NameDeclaration
		buildStreamPayload *codegen.NameDeclaration
		cliPayload         *codegen.NameDeclaration
		serverMultipart    *httpMultipartSymbols
		clientMultipart    *httpMultipartSymbols
		serverStream       *codegen.NameDeclaration
		clientStream       *codegen.NameDeclaration
		sseClientInterface *codegen.NameDeclaration
		sseClientStruct    *codegen.NameDeclaration
		sseClientInit      *codegen.NameDeclaration
		serverPaths        []*codegen.NameDeclaration
		clientPaths        []*codegen.NameDeclaration
	}

	// httpMultipartSymbols contains the type and constructor names for one side
	// of a multipart endpoint.
	httpMultipartSymbols struct {
		functionType *codegen.NameDeclaration
		constructor  *codegen.NameDeclaration
	}

	// httpSymbolID identifies one emitted declaration without encoding fields in
	// a string. The output package is supplied separately by the caller.
	httpSymbolID struct {
		transport transportKind
		role      httpSymbolRole
		service   string
		method    string
		subject   string
		index     int
	}

	// httpSymbolOrder gives colliding declarations the same result regardless of
	// the order in which design roots are passed to NewPlans.
	httpSymbolOrder httpSymbolID

	// httpSymbolRole lists each package declaration emitted outside HTTP body
	// type and constructor files.
	httpSymbolRole uint8
)

const (
	httpServerStructRole httpSymbolRole = iota + 1
	httpMountPointRole
	httpServerInitRole
	httpMountServerRole
	httpClientStructRole
	httpClientInitRole
	httpConnConfigurerRole
	httpConnConfigurerInitRole
	httpAppendFSRole
	httpAppendPrefixRole
	httpMountHandlerRole
	httpHandlerInitRole
	httpRequestDecoderRole
	httpResponseEncoderRole
	httpErrorEncoderRole
	httpDiscardStreamRole
	httpRequestEncoderRole
	httpResponseDecoderRole
	httpBuildStreamPayloadRole
	httpCLIPayloadRole
	httpMultipartTypeRole
	httpMultipartInitRole
	httpServerStreamRole
	httpClientStreamRole
	httpSSEClientInterfaceRole
	httpSSEClientStructRole
	httpSSEClientInitRole
	httpPathRole
	httpFileMountRole
)

// collectHTTPSymbols requests each client and server name needed by service. It
// returns records that Link gives to definitions and calls after Goa assigns
// the names.
func collectHTTPSymbols(plan *Plan, service *expr.HTTPServiceExpr, clientPackage, serverPackage *codegen.GeneratedPackage) (*httpSymbols, error) {
	symbols := &httpSymbols{
		endpoints:   make(map[*expr.HTTPEndpointExpr]*httpEndpointSymbols),
		fileServers: make(map[*expr.HTTPFileServerExpr]*codegen.NameDeclaration),
	}
	declare := func(pkg *codegen.GeneratedPackage, kind codegen.PackageNameKind, preferred string, exported codegen.PackageNameVisibility, id httpSymbolID) (*codegen.NameDeclaration, error) {
		declaration := codegen.NewPreferredName(kind, preferred, exported, httpSymbolOrder(id))
		if err := pkg.DeclareName(declaration); err != nil {
			return nil, err
		}
		return declaration, nil
	}
	serviceID := httpSymbolID{transport: plan.transport, service: service.Name()}
	var err error
	if symbols.serverStruct, err = declare(serverPackage, codegen.NameType, "Server", codegen.ExportedName, serviceID.withRole(httpServerStructRole)); err != nil {
		return nil, err
	}
	if symbols.mountPoint, err = declare(serverPackage, codegen.NameType, "MountPoint", codegen.ExportedName, serviceID.withRole(httpMountPointRole)); err != nil {
		return nil, err
	}
	if symbols.serverInit, err = declare(serverPackage, codegen.NameFunction, "New", codegen.ExportedName, serviceID.withRole(httpServerInitRole)); err != nil {
		return nil, err
	}
	if symbols.mountServer, err = declare(serverPackage, codegen.NameFunction, "Mount", codegen.ExportedName, serviceID.withRole(httpMountServerRole)); err != nil {
		return nil, err
	}
	if symbols.clientStruct, err = declare(clientPackage, codegen.NameType, "Client", codegen.ExportedName, serviceID.withRole(httpClientStructRole)); err != nil {
		return nil, err
	}
	if symbols.clientInit, err = declare(clientPackage, codegen.NameFunction, "NewClient", codegen.ExportedName, serviceID.withRole(httpClientInitRole)); err != nil {
		return nil, err
	}

	hasWebSocket := false
	for _, endpoint := range service.HTTPEndpoints {
		if endpoint.UsesWebSocket() {
			hasWebSocket = true
			break
		}
	}
	if hasWebSocket {
		if symbols.serverConfigurer, err = declare(serverPackage, codegen.NameType, "ConnConfigurer", codegen.ExportedName, serviceID.withRole(httpConnConfigurerRole).withSubject("server")); err != nil {
			return nil, err
		}
		if symbols.serverConfigurerInit, err = declare(serverPackage, codegen.NameFunction, "NewConnConfigurer", codegen.ExportedName, serviceID.withRole(httpConnConfigurerInitRole).withSubject("server")); err != nil {
			return nil, err
		}
		if symbols.clientConfigurer, err = declare(clientPackage, codegen.NameType, "ConnConfigurer", codegen.ExportedName, serviceID.withRole(httpConnConfigurerRole).withSubject("client")); err != nil {
			return nil, err
		}
		if symbols.clientConfigurerInit, err = declare(clientPackage, codegen.NameFunction, "NewConnConfigurer", codegen.ExportedName, serviceID.withRole(httpConnConfigurerInitRole).withSubject("client")); err != nil {
			return nil, err
		}
	}
	if len(service.FileServers) > 0 {
		if symbols.appendFS, err = declare(serverPackage, codegen.NameType, "appendFS", codegen.UnexportedName, serviceID.withRole(httpAppendFSRole)); err != nil {
			return nil, err
		}
		if symbols.appendPrefix, err = declare(serverPackage, codegen.NameFunction, "appendPrefix", codegen.UnexportedName, serviceID.withRole(httpAppendPrefixRole)); err != nil {
			return nil, err
		}
	}
	for index, fileServer := range service.FileServers {
		id := serviceID.withRole(httpFileMountRole).withSubject(fileServer.FilePath).withIndex(index)
		declaration, err := declare(serverPackage, codegen.NameFunction, "Mount"+codegen.Goify(fileServer.FilePath, true), codegen.ExportedName, id)
		if err != nil {
			return nil, err
		}
		symbols.fileServers[fileServer] = declaration
	}
	for _, endpoint := range service.HTTPEndpoints {
		names, err := plan.servicePlan.HTTPMethodNames(endpoint.MethodExpr)
		if err != nil {
			return nil, err
		}
		id := serviceID.withMethod(endpoint.MethodExpr.Name)
		endpointSymbols := &httpEndpointSymbols{}
		endpointSymbols.mountHandler, err = declare(serverPackage, codegen.NameFunction, "Mount"+names.Method+"Handler", codegen.ExportedName, id.withRole(httpMountHandlerRole))
		if err != nil {
			return nil, err
		}
		endpointSymbols.handlerInit, err = declare(serverPackage, codegen.NameFunction, "New"+names.Method+"Handler", codegen.ExportedName, id.withRole(httpHandlerInitRole))
		if err != nil {
			return nil, err
		}
		if endpoint.MethodExpr.Payload.Type != expr.Empty {
			endpointSymbols.requestDecoder, err = declare(serverPackage, codegen.NameFunction, "Decode"+names.Method+"Request", codegen.ExportedName, id.withRole(httpRequestDecoderRole))
			if err != nil {
				return nil, err
			}
		}
		if endpoint.Redirect == nil && !endpoint.UsesWebSocket() && !endpoint.IsJSONRPC() {
			endpointSymbols.responseEncoder, err = declare(serverPackage, codegen.NameFunction, "Encode"+names.Method+"Response", codegen.ExportedName, id.withRole(httpResponseEncoderRole))
			if err != nil {
				return nil, err
			}
		}
		if len(endpoint.HTTPErrors) > 0 && !endpoint.IsJSONRPC() {
			endpointSymbols.errorEncoder, err = declare(serverPackage, codegen.NameFunction, "Encode"+names.Method+"Error", codegen.ExportedName, id.withRole(httpErrorEncoderRole))
			if err != nil {
				return nil, err
			}
		}
		if endpoint.MethodExpr.HasMixedResults() {
			endpointSymbols.discardStream, err = declare(serverPackage, codegen.NameType, "discard"+names.Method+"ServerStream", codegen.UnexportedName, id.withRole(httpDiscardStreamRole))
			if err != nil {
				return nil, err
			}
		}
		if clientRequestEncoderSelected(endpoint) {
			endpointSymbols.requestEncoder, err = declare(clientPackage, codegen.NameFunction, "Encode"+names.Method+"Request", codegen.ExportedName, id.withRole(httpRequestEncoderRole))
			if err != nil {
				return nil, err
			}
		}
		endpointSymbols.responseDecoder, err = declare(clientPackage, codegen.NameFunction, "Decode"+names.Method+"Response", codegen.ExportedName, id.withRole(httpResponseDecoderRole))
		if err != nil {
			return nil, err
		}
		if endpoint.SkipRequestBodyEncodeDecode {
			endpointSymbols.buildStreamPayload, err = declare(clientPackage, codegen.NameFunction, "Build"+names.Method+"StreamPayload", codegen.ExportedName, id.withRole(httpBuildStreamPayloadRole))
			if err != nil {
				return nil, err
			}
		}
		if needInit(endpoint.MethodExpr.Payload.Type) {
			endpointSymbols.cliPayload, err = declare(clientPackage, codegen.NameFunction, "Build"+names.Method+"Payload", codegen.ExportedName, id.withRole(httpCLIPayloadRole))
			if err != nil {
				return nil, err
			}
		}
		if endpoint.MultipartRequest {
			serviceName := codegen.Goify(service.Name(), true)
			endpointSymbols.serverMultipart, err = declareHTTPMultipart(declare, serverPackage, serviceName+names.Method+"DecoderFunc", "New"+serviceName+names.Method+"Decoder", id.withSubject("server"))
			if err != nil {
				return nil, err
			}
			endpointSymbols.clientMultipart, err = declareHTTPMultipart(declare, clientPackage, serviceName+names.Method+"EncoderFunc", "New"+serviceName+names.Method+"Encoder", id.withSubject("client"))
			if err != nil {
				return nil, err
			}
		}
		if endpoint.UsesWebSocket() {
			endpointSymbols.serverStream, err = declare(serverPackage, codegen.NameType, names.ServerStream, codegen.ExportedName, id.withRole(httpServerStreamRole))
			if err != nil {
				return nil, err
			}
			endpointSymbols.clientStream, err = declare(clientPackage, codegen.NameType, names.ClientStream, codegen.ExportedName, id.withRole(httpClientStreamRole))
			if err != nil {
				return nil, err
			}
		}
		if endpoint.UsesSSE() {
			endpointSymbols.serverStream, err = declare(serverPackage, codegen.NameType, names.ServerStream, codegen.ExportedName, id.withRole(httpServerStreamRole))
			if err != nil {
				return nil, err
			}
			endpointSymbols.sseClientInterface, err = declare(clientPackage, codegen.NameType, names.Method+"ClientStream", codegen.ExportedName, id.withRole(httpSSEClientInterfaceRole))
			if err != nil {
				return nil, err
			}
			endpointSymbols.sseClientStruct, err = declare(clientPackage, codegen.NameType, names.Method+"StreamImpl", codegen.ExportedName, id.withRole(httpSSEClientStructRole))
			if err != nil {
				return nil, err
			}
			endpointSymbols.sseClientInit, err = declare(clientPackage, codegen.NameFunction, "New"+names.Method+"Stream", codegen.ExportedName, id.withRole(httpSSEClientInitRole))
			if err != nil {
				return nil, err
			}
		}
		pathCount := 0
		for _, route := range endpoint.Routes {
			for range route.FullPaths() {
				suffix := ""
				if pathCount > 0 {
					suffix = strconv.Itoa(pathCount + 1)
				}
				preferred := names.Method + codegen.Goify(service.Name(), true) + "Path" + suffix
				pathID := id.withRole(httpPathRole).withIndex(pathCount)
				serverPath, err := declare(serverPackage, codegen.NameFunction, preferred, codegen.ExportedName, pathID.withSubject("server"))
				if err != nil {
					return nil, err
				}
				clientPath, err := declare(clientPackage, codegen.NameFunction, preferred, codegen.ExportedName, pathID.withSubject("client"))
				if err != nil {
					return nil, err
				}
				endpointSymbols.serverPaths = append(endpointSymbols.serverPaths, serverPath)
				endpointSymbols.clientPaths = append(endpointSymbols.clientPaths, clientPath)
				pathCount++
			}
		}
		symbols.endpoints[endpoint] = endpointSymbols
	}
	return symbols, nil
}

// declareHTTPMultipart requests the type and constructor emitted for one
// multipart endpoint side.
func declareHTTPMultipart(declare func(*codegen.GeneratedPackage, codegen.PackageNameKind, string, codegen.PackageNameVisibility, httpSymbolID) (*codegen.NameDeclaration, error), pkg *codegen.GeneratedPackage, typeName, initName string, id httpSymbolID) (*httpMultipartSymbols, error) {
	functionType, err := declare(pkg, codegen.NameType, typeName, codegen.ExportedName, id.withRole(httpMultipartTypeRole))
	if err != nil {
		return nil, err
	}
	constructor, err := declare(pkg, codegen.NameFunction, initName, codegen.ExportedName, id.withRole(httpMultipartInitRole))
	if err != nil {
		return nil, err
	}
	return &httpMultipartSymbols{functionType: functionType, constructor: constructor}, nil
}

// clientRequestEncoderSelected reports whether the client codec file writes a
// request encoder for endpoint.
func clientRequestEncoderSelected(endpoint *expr.HTTPEndpointExpr) bool {
	if endpoint.IsJSONRPC() {
		return true
	}
	if endpoint.SkipRequestBodyEncodeDecode {
		return false
	}
	if endpoint.Body.Type != expr.Empty || endpoint.MapQueryParams != nil ||
		len(*expr.AsObject(endpoint.QueryParams().Type)) > 0 ||
		len(*expr.AsObject(endpoint.Headers.Type)) > 0 ||
		len(*expr.AsObject(endpoint.Cookies.Type)) > 0 {
		return true
	}
	for _, requirement := range endpoint.Requirements {
		for _, scheme := range requirement.Schemes {
			if scheme.Kind == expr.BasicAuthKind {
				return true
			}
		}
	}
	return false
}

// withRole returns id with the declaration role used by one template.
func (id httpSymbolID) withRole(role httpSymbolRole) httpSymbolID {
	id.role = role
	return id
}

// withMethod returns id with the design method that emits the declaration.
func (id httpSymbolID) withMethod(method string) httpSymbolID {
	id.method = method
	return id
}

// withSubject returns id with the route side or file path that distinguishes
// otherwise identical declarations.
func (id httpSymbolID) withSubject(subject string) httpSymbolID {
	id.subject = subject
	return id
}

// withIndex returns id with the route or file position in its design list.
func (id httpSymbolID) withIndex(index int) httpSymbolID {
	id.index = index
	return id
}

// ComparePackageName orders HTTP declarations by stable design values.
func (order httpSymbolOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	left := httpSymbolID(order)
	right := httpSymbolID(other.(httpSymbolOrder))
	for _, compared := range []int{
		cmp.Compare(left.transport, right.transport),
		cmp.Compare(left.service, right.service),
		cmp.Compare(left.method, right.method),
		cmp.Compare(left.role, right.role),
		cmp.Compare(left.subject, right.subject),
		cmp.Compare(left.index, right.index),
	} {
		if compared != 0 {
			return compared
		}
	}
	return 0
}
