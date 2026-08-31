// This file builds the values used to write WebSocket client and server files
// for streaming HTTP methods.
package codegen

import (
	"fmt"
	"path/filepath"
	"slices"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// connConfigurerData gives WebSocket code the type and constructor names for
	// either the client or server package.
	connConfigurerData struct {
		*ServiceData
		Declaration     *codegen.NameDeclaration
		InitDeclaration *codegen.NameDeclaration
	}

	// WebSocketData contains the data needed to render struct type that
	// implements the server and client stream interfaces.
	WebSocketData struct {
		// VarName is the stream implementation type name kept for existing plugins.
		//
		// Deprecated: Use VarDeclaration.Name() after planning so name collisions are handled.
		VarName string
		// VarDeclaration is the generated Go type name used by the stream implementation.
		VarDeclaration *codegen.NameDeclaration
		// Type is type of the stream (server or client).
		Type string
		// Interface is the fully qualified name of the interface that
		// the struct implements.
		Interface string
		// Endpoint is endpoint data that defines streaming
		// payload/result.
		Endpoint *EndpointData
		// Payload is the streaming payload type sent via the stream.
		Payload *TypeData
		// Response is the successful response data for the streaming
		// endpoint.
		Response *ResponseData
		// SendName is the name of the send function.
		SendName string
		// SendDesc is the description for the send function.
		SendDesc string
		// SendWithContextName is the name of the send function with context.
		SendWithContextName string
		// SendWithContextDesc is the description for the send function with context.
		SendWithContextDesc string
		// SendTypeName is the fully qualified type name sent through
		// the stream.
		SendTypeName string
		// SendTypeRef is the fully qualified type ref sent through the
		// stream.
		SendTypeRef string
		// RecvName is the name of the receive function.
		RecvName string
		// RecvDesc is the description for the recv function.
		RecvDesc string
		// RecvWithContextName is the name of the receive function with context.
		RecvWithContextName string
		// RecvWithContextDesc is the description for the recv function with context.
		RecvWithContextDesc string
		// RecvTypeName is the fully qualified type name received from
		// the stream.
		RecvTypeName string
		// RecvTypeRef is the fully qualified type ref received from the
		// stream.
		RecvTypeRef string
		// RecvTypeIsPointer is true if the type received from the stream is a
		// array or map. This is needed so that the code reading the stream can
		// use a pointer reference when needed to check whether anything was
		// read (check against the nil value) and in this case return EOF.
		RecvTypeIsPointer bool
		// MustClose indicates whether to generate the Close() function
		// for the stream.
		MustClose bool
		// SelectClientBodyByView is true when the client stream must choose a
		// response body, validator, and constructor from its selected view.
		SelectClientBodyByView bool
		// PkgName is the service package name.
		PkgName string
		// Kind is the kind of the stream (payload, result or
		// bidirectional).
		Kind expr.StreamKind
	}
)

// initWebSocketData initializes the WebSocket related data in ed.
func (sds *ServicesData) initWebSocketData(ed *EndpointData, e *expr.HTTPEndpointExpr, sd *ServiceData) {
	if !e.UsesWebSocket() {
		return
	}
	var (
		svrRecvTypeName        string
		svrRecvTypeRef         string
		svrRecvDesc            string
		svrRecvWithContextDesc string
		svrPayload             *TypeData
		cliSendDesc            string
		cliSendWithContextDesc string
		cliPayload             *TypeData
	)
	md := ed.Method
	svc := sd.Service
	svcctx := sds.serviceTypeContext(sd, "server").Enter(e.MethodExpr.StreamingPayload)
	svrSendTypeName := ed.Result.Name
	svrSendTypeRef := ed.Result.Ref
	svrSendDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection.", md.ServerStream.SendName, svrSendTypeName, md.Name)
	svrSendWithContextDesc := fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection with context.", md.ServerStream.SendWithContextName, svrSendTypeName, md.Name)
	cliRecvDesc := fmt.Sprintf("%s reads instances of %q from the %q endpoint websocket connection.", md.ClientStream.RecvName, svrSendTypeName, md.Name)
	cliRecvWithContextDesc := fmt.Sprintf("%s reads instances of %q from the %q endpoint websocket connection with context.", md.ClientStream.RecvWithContextName, svrSendTypeName, md.Name)
	if e.MethodExpr.Stream == expr.ClientStreamKind || e.MethodExpr.Stream == expr.BidirectionalStreamKind {
		streamBody := sd.bodies.streaming(e)
		streamOwner := expr.MethodStreamingPayloadExampleIdentity(e.MethodExpr)
		svrRecvTypeName = svcctx.Scope.Name(e.MethodExpr.StreamingPayload, svcctx.Pkg(e.MethodExpr.StreamingPayload), false, true)
		svrRecvTypeRef = svcctx.Scope.Ref(e.MethodExpr.StreamingPayload, svcctx.Pkg(e.MethodExpr.StreamingPayload))
		svrPayload = sds.buildRequestBodyType(streamBody, e.MethodExpr.StreamingPayload, e, wireStreamPayload, true, sd, streamOwner, streamOwner)
		if needInit(e.MethodExpr.StreamingPayload.Type) {
			body := streamBody.Type
			// generate constructor function to transform request body,
			// into the method streaming payload type
			var (
				name       string
				desc       string
				serverArgs []*InitArgData
				serverCode string
				err        error
			)
			declaration := sds.streamConstructors[e]
			if declaration == nil {
				panic(fmt.Sprintf("streaming payload constructor for %s.%s was not submitted", svc.Name, e.Name()))
			}
			name = declaration.Name()
			desc = fmt.Sprintf("%s builds a %s service %s endpoint payload.", name, svc.Name, e.MethodExpr.Name)
			if body != expr.Empty {
				ref := "body"
				if expr.IsObject(body) {
					ref = "&body"
				}
				var svcode string
				if ut, ok := body.(expr.UserType); ok {
					if val := ut.Attribute().Validation; val != nil {
						httpctx := jsonBodyContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
						svcode = codegen.ValidationCode(ut.Attribute(), ut, httpctx, true, expr.IsAlias(ut), false, "body")
					}
				}
				serverArgs = []*InitArgData{{
					Ref: ref,
					AttributeData: &AttributeData{
						Name:     "payload",
						VarName:  "body",
						TypeName: svrPayload.VarName,
						TypeRef:  svrPayload.Ref,
						Type:     streamBody.Type,
						Required: true,
						Example:  sds.Example(streamBody, streamOwner),
						Validate: svcode,
					},
				}}
			}
			if body != expr.Empty {
				var helpers []*codegen.TransformFunctionData
				httpctx := jsonBodyContext(sd.serverWireTypes, sd.serverWireTypes.scope, true, true)
				transforms := sd.transforms.requests[clientBodyConstructorKey{endpoint: e, role: wireStreamPayload}]
				serverCode, helpers, err = sd.serverWireTypes.renderTransform(transforms.serverDecode, streamBody, "body", "v", httpctx, svcctx)
				if err == nil {
					sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
				}
			}
			if err != nil {
				panic(err) // bug
			}
			svrPayload.Init = &InitData{
				Declaration:    declaration,
				Name:           name,
				Description:    desc,
				ServerArgs:     serverArgs,
				ReturnTypeName: svcctx.Scope.Name(e.MethodExpr.StreamingPayload, svcctx.Pkg(e.MethodExpr.StreamingPayload), false, true),
				ReturnTypeRef:  svcctx.Scope.Ref(e.MethodExpr.StreamingPayload, svcctx.Pkg(e.MethodExpr.StreamingPayload)),
				ReturnIsStruct: expr.IsObject(e.MethodExpr.StreamingPayload.Type),
				ReturnTypePkg:  svcctx.Pkg(e.MethodExpr.StreamingPayload),
				ServerCode:     serverCode,
			}
		}
		cliPayload = sds.buildRequestBodyType(streamBody, e.MethodExpr.StreamingPayload, e, wireStreamPayload, false, sd, streamOwner, streamOwner)
		if e.MethodExpr.Stream == expr.ClientStreamKind {
			svrSendDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection and closes the connection.", md.ServerStream.SendName, svrSendTypeName, md.Name)
			svrSendWithContextDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection with context and closes the connection.", md.ServerStream.SendWithContextName, svrSendTypeName, md.Name)
			cliRecvDesc = fmt.Sprintf("%s stops sending messages to the %q endpoint websocket connection and reads instances of %q from the connection.", md.ClientStream.RecvName, md.Name, svrSendTypeName)
			cliRecvWithContextDesc = fmt.Sprintf("%s stops sending messages to the %q endpoint websocket connection and reads instances of %q from the connection with context.", md.ClientStream.RecvWithContextName, md.Name, svrSendTypeName)
		}
		svrRecvDesc = fmt.Sprintf("%s reads instances of %q from the %q endpoint websocket connection.", md.ServerStream.RecvName, svrRecvTypeName, md.Name)
		svrRecvWithContextDesc = fmt.Sprintf("%s reads instances of %q from the %q endpoint websocket connection with context.", md.ServerStream.RecvWithContextName, svrRecvTypeName, md.Name)
		cliSendDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection.", md.ClientStream.SendName, svrRecvTypeName, md.Name)
		cliSendWithContextDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection with context.", md.ClientStream.SendWithContextName, svrRecvTypeName, md.Name)
	}
	ed.ServerWebSocket = &WebSocketData{
		VarName:             md.ServerStream.VarName,
		Interface:           fmt.Sprintf("%s.%s", svc.PkgName, md.ServerStream.Interface),
		Endpoint:            ed,
		Payload:             svrPayload,
		Response:            ed.Result.Responses[0],
		PkgName:             svc.PkgName,
		Type:                "server",
		Kind:                md.ServerStream.Kind,
		SendName:            md.ServerStream.SendName,
		SendDesc:            svrSendDesc,
		SendWithContextName: md.ServerStream.SendWithContextName,
		SendWithContextDesc: svrSendWithContextDesc,
		SendTypeName:        svrSendTypeName,
		SendTypeRef:         svrSendTypeRef,
		RecvName:            md.ServerStream.RecvName,
		RecvDesc:            svrRecvDesc,
		RecvWithContextName: md.ServerStream.RecvWithContextName,
		RecvWithContextDesc: svrRecvWithContextDesc,
		RecvTypeName:        svrRecvTypeName,
		RecvTypeRef:         svrRecvTypeRef,
		RecvTypeIsPointer:   expr.IsArray(e.MethodExpr.StreamingPayload.Type) || expr.IsMap(e.MethodExpr.StreamingPayload.Type),
		MustClose:           md.ServerStream.MustClose,
	}
	ed.ClientWebSocket = &WebSocketData{
		VarName:                md.ClientStream.VarName,
		Interface:              fmt.Sprintf("%s.%s", svc.PkgName, md.ClientStream.Interface),
		Endpoint:               ed,
		Payload:                cliPayload,
		Response:               ed.Result.Responses[0],
		PkgName:                svc.PkgName,
		Type:                   "client",
		Kind:                   md.ClientStream.Kind,
		SendName:               md.ClientStream.SendName,
		SendDesc:               cliSendDesc,
		SendWithContextName:    md.ClientStream.SendWithContextName,
		SendWithContextDesc:    cliSendWithContextDesc,
		SendTypeName:           svrRecvTypeName,
		SendTypeRef:            svrRecvTypeRef,
		RecvName:               md.ClientStream.RecvName,
		RecvDesc:               cliRecvDesc,
		RecvWithContextName:    md.ClientStream.RecvWithContextName,
		RecvWithContextDesc:    cliRecvWithContextDesc,
		RecvTypeName:           svrSendTypeName,
		RecvTypeRef:            svrSendTypeRef,
		MustClose:              md.ClientStream.MustClose,
		SelectClientBodyByView: !e.IsJSONRPC() && len(ed.Result.Responses[0].ViewedRepresentations) > 0,
	}
}

// websocketServerFile returns the file implementing the WebSocket server
// streaming implementation if any.
func websocketServerFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !HasWebSocket(data) {
		return nil
	}
	svcName := data.Service.PathName
	outputPath := filepath.Join(codegen.Gendir, "http", svcName, "server", "websocket.go")
	outputPackage := generatedFileOutputPackage(services, outputPath)
	data = serviceDataForOutput(data, services, outputPackage)
	title := fmt.Sprintf("%s WebSocket server streaming", svc.Name())
	structSections := serverStructWSSections(data)
	wsSections := serverWSSections(data)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(structSections)+len(wsSections))
	sections = append(sections, plannedFileHeader(title, "server", outputPath, services))
	sections = append(sections, structSections...)
	sections = append(sections, wsSections...)

	return &codegen.File{
		Path:             outputPath,
		SectionTemplates: sections,
	}
}

// websocketClientFile returns the file implementing the WebSocket client
// streaming implementation if any.
func websocketClientFile(svc *expr.HTTPServiceExpr, services *ServicesData) *codegen.File {
	data := services.Get(svc.Name())
	if !HasWebSocket(data) {
		return nil
	}
	svcName := data.Service.PathName
	outputPath := filepath.Join(codegen.Gendir, "http", svcName, "client", "websocket.go")
	outputPackage := generatedFileOutputPackage(services, outputPath)
	data = serviceDataForOutput(data, services, outputPackage)
	title := fmt.Sprintf("%s WebSocket client streaming", svc.Name())
	structSections := clientStructWSSections(data)
	wsSections := clientWSSections(data)
	sections := make([]*codegen.SectionTemplate, 0, 1+len(structSections)+len(wsSections))
	sections = append(sections, plannedFileHeader(title, "client", outputPath, services))
	sections = append(sections, structSections...)
	sections = append(sections, wsSections...)

	return &codegen.File{
		Path:             outputPath,
		SectionTemplates: sections,
	}
}

// serverStructWSSections return section templates that generate WebSocket
// related struct type definitions for the server.
func serverStructWSSections(data *ServiceData) []*codegen.SectionTemplate {
	configurer := &connConfigurerData{data, data.ServerConnConfigurerDeclaration, data.ServerConnConfigurerInitDeclaration}
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "server-websocket-conn-configurer-struct",
		Source:  httpTemplates.Read(websocketConnConfigurerStructT),
		Data:    configurer,
		FuncMap: map[string]any{"isWebSocketEndpoint": IsWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ServerWebSocket != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-websocket-struct-type",
				Source: httpTemplates.Read(websocketStructTypeT),
				Data:   e.ServerWebSocket,
			})
		}
	}

	return sections
}

// serverWSSections returns section templates that contain server WebSocket
// specific code for the given service.
func serverWSSections(data *ServiceData) []*codegen.SectionTemplate {
	configurer := &connConfigurerData{data, data.ServerConnConfigurerDeclaration, data.ServerConnConfigurerInitDeclaration}
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "server-websocket-conn-configurer-struct-init",
		Source:  httpTemplates.Read(websocketConnConfigurerStructInitT),
		Data:    configurer,
		FuncMap: map[string]any{"isWebSocketEndpoint": IsWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ServerWebSocket != nil {
			if e.ServerWebSocket.SendTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-send",
					Source: httpTemplates.Read(websocketSendT, websocketUpgradeP),
					Data:   e.ServerWebSocket,
					FuncMap: map[string]any{
						"upgradeParams":      upgradeParams,
						"viewedServerBody":   viewedServerBody,
						"isClientStreamKind": isClientStreamKind,
					},
				})
			}
			switch e.ServerWebSocket.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-recv",
					Source: httpTemplates.Read(websocketRecvT, websocketUpgradeP),
					Data:   e.ServerWebSocket,
					FuncMap: map[string]any{
						"upgradeParams":      upgradeParams,
						"isClientStreamKind": isClientStreamKind,
					},
				})
			}
			if e.ServerWebSocket.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-close",
					Source: httpTemplates.Read(websocketCloseT, websocketUpgradeP),
					Data:   e.ServerWebSocket,
					FuncMap: map[string]any{
						"upgradeParams":      upgradeParams,
						"isClientStreamKind": isClientStreamKind,
					},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-set-view",
					Source: httpTemplates.Read(websocketSetViewT),
					Data:   e.ServerWebSocket,
				})
			}
		}
	}
	return sections
}

// clientStructWSSections return section templates that generate WebSocket
// related struct type definitions for the client.
func clientStructWSSections(data *ServiceData) []*codegen.SectionTemplate {
	configurer := &connConfigurerData{data, data.ClientConnConfigurerDeclaration, data.ClientConnConfigurerInitDeclaration}
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "client-websocket-conn-configurer-struct",
		Source:  httpTemplates.Read(websocketConnConfigurerStructT),
		Data:    configurer,
		FuncMap: map[string]any{"isWebSocketEndpoint": IsWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ClientWebSocket != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-websocket-struct-type",
				Source: httpTemplates.Read(websocketStructTypeT),
				Data:   e.ClientWebSocket,
			})
		}
	}
	return sections
}

// clientWSSections returns section templates that contain client WebSocket
// specific code for the given service.
func clientWSSections(data *ServiceData) []*codegen.SectionTemplate {
	configurer := &connConfigurerData{data, data.ClientConnConfigurerDeclaration, data.ClientConnConfigurerInitDeclaration}
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "client-websocket-conn-configurer-struct-init",
		Source:  httpTemplates.Read(websocketConnConfigurerStructInitT),
		Data:    configurer,
		FuncMap: map[string]any{"isWebSocketEndpoint": IsWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ClientWebSocket != nil {
			if e.ClientWebSocket.RecvTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-websocket-recv",
					Source: httpTemplates.Read(websocketRecvT, websocketUpgradeP),
					Data:   e.ClientWebSocket,
					FuncMap: map[string]any{
						"upgradeParams":      upgradeParams,
						"isClientStreamKind": isClientStreamKind,
					},
				})
			}
			switch e.ClientWebSocket.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-websocket-send",
					Source: httpTemplates.Read(websocketSendT, websocketUpgradeP),
					Data:   e.ClientWebSocket,
					FuncMap: map[string]any{
						"upgradeParams":      upgradeParams,
						"viewedServerBody":   viewedServerBody,
						"isClientStreamKind": isClientStreamKind,
					},
				})
			}
			if e.ClientWebSocket.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "client-websocket-close",
					Source:  httpTemplates.Read(websocketCloseT),
					Data:    e.ClientWebSocket,
					FuncMap: map[string]any{"upgradeParams": upgradeParams},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-websocket-set-view",
					Source: httpTemplates.Read(websocketSetViewT),
					Data:   e.ClientWebSocket,
				})
			}
		}
	}
	return sections
}

// HasWebSocket returns true if at least one of the endpoints in the service
// defines a streaming payload or result.
func HasWebSocket(sd *ServiceData) bool {
	return slices.ContainsFunc(sd.Endpoints, IsWebSocketEndpoint)
}

// isClientStreamKind reports whether the client finishes sending before it
// receives the server's single result.
func isClientStreamKind(kind expr.StreamKind) bool {
	return kind == expr.ClientStreamKind
}

// isServerStreamKind reports whether the client only receives stream values.
func isServerStreamKind(kind expr.StreamKind) bool {
	return kind == expr.ServerStreamKind
}

// IsWebSocketEndpoint returns true if the endpoint defines a streaming payload
// or result.
func IsWebSocketEndpoint(ed *EndpointData) bool {
	return ed.ServerWebSocket != nil || ed.ClientWebSocket != nil
}
