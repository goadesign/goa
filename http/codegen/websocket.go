package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

type (
	// WebSocketData contains the data needed to render struct type that
	// implements the server and client stream interfaces.
	WebSocketData struct {
		// VarName is the name of the struct.
		VarName string
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
		// PkgName is the service package name.
		PkgName string
		// Kind is the kind of the stream (payload, result or
		// bidirectional).
		Kind expr.StreamKind
	}
)

// initWebSocketData initializes the WebSocket related data in ed.
func initWebSocketData(ed *EndpointData, e *expr.HTTPEndpointExpr, sd *ServiceData) {
	var (
		svrSendTypeName string
		svrSendTypeRef  string
		svrRecvTypeName string
		svrRecvTypeRef  string
		svrSendDesc     string
		svrRecvDesc     string
		svrPayload      *TypeData
		cliSendDesc     string
		cliRecvDesc     string
		cliPayload      *TypeData

		md     = ed.Method
		svc    = sd.Service
		svcctx = serviceContext(sd.Service.PkgName, sd.Service.Scope)
	)
	{
		svrSendTypeName = ed.Result.Name
		svrSendTypeRef = ed.Result.Ref
		svrSendDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection.", md.ServerStream.SendName, svrSendTypeName, md.Name)
		cliRecvDesc = fmt.Sprintf("%s reads instances of %q from the %q endpoint websocket connection.", md.ClientStream.RecvName, svrSendTypeName, md.Name)
		if e.MethodExpr.Stream == expr.ClientStreamKind || e.MethodExpr.Stream == expr.BidirectionalStreamKind {
			svrRecvTypeName = sd.Scope.GoFullTypeName(e.MethodExpr.StreamingPayload, svc.PkgName)
			svrRecvTypeRef = sd.Scope.GoFullTypeRef(e.MethodExpr.StreamingPayload, svc.PkgName)
			svrPayload = buildRequestBodyType(e.StreamingBody, e.MethodExpr.StreamingPayload, e, true, sd)
			if needInit(e.MethodExpr.StreamingPayload.Type) {
				makeHTTPType(e.StreamingBody)
				body := e.StreamingBody.Type
				// generate constructor function to transform request body,
				// into the method streaming payload type
				var (
					name       string
					desc       string
					serverArgs []*InitArgData
					serverCode string
					err        error
				)
				{
					n := codegen.Goify(e.MethodExpr.Name, true)
					p := codegen.Goify(svrPayload.Name, true)
					// Raw payload object has type name prefixed with endpoint name. No need to
					// prefix the type name again.
					if strings.HasPrefix(p, n) {
						name = fmt.Sprintf("New%s", p)
					} else {
						name = fmt.Sprintf("New%s%s", n, p)
					}
					desc = fmt.Sprintf("%s builds a %s service %s endpoint payload.", name, svc.Name, e.MethodExpr.Name)
					if body != expr.Empty {
						var (
							ref    string
							svcode string
						)
						{
							ref = "body"
							if expr.IsObject(body) {
								ref = "&body"
							}
							if ut, ok := body.(expr.UserType); ok {
								if val := ut.Attribute().Validation; val != nil {
									httpctx := httpContext("", sd.Scope, true, true)
									svcode = codegen.ValidationCode(ut.Attribute(), ut, httpctx, true, expr.IsAlias(ut), "body")
								}
							}
						}
						serverArgs = []*InitArgData{{
							Ref: ref,
							AttributeData: &AttributeData{
								Name:     "payload",
								VarName:  "body",
								TypeName: sd.Scope.GoTypeName(e.StreamingBody),
								TypeRef:  sd.Scope.GoTypeRef(e.StreamingBody),
								Type:     e.StreamingBody.Type,
								Required: true,
								Example:  e.Body.Example(expr.Root.API.ExampleGenerator),
								Validate: svcode,
							},
						}}
					}
					if body != expr.Empty {
						var helpers []*codegen.TransformFunctionData
						httpctx := httpContext("", sd.Scope, true, true)
						serverCode, helpers, err = marshal(e.StreamingBody, e.MethodExpr.StreamingPayload, "body", "v", httpctx, svcctx)
						if err == nil {
							sd.ServerTransformHelpers = codegen.AppendHelpers(sd.ServerTransformHelpers, helpers)
						}
					}
					if err != nil {
						fmt.Println(err.Error()) // TBD validate DSL so errors are not possible
					}
				}
				svrPayload.Init = &InitData{
					Name:           name,
					Description:    desc,
					ServerArgs:     serverArgs,
					ReturnTypeName: svc.Scope.GoFullTypeName(e.MethodExpr.StreamingPayload, svc.PkgName),
					ReturnTypeRef:  svc.Scope.GoFullTypeRef(e.MethodExpr.StreamingPayload, svc.PkgName),
					ReturnIsStruct: expr.IsObject(e.MethodExpr.StreamingPayload.Type),
					ReturnTypePkg:  svc.PkgName,
					ServerCode:     serverCode,
				}
			}
			cliPayload = buildRequestBodyType(e.StreamingBody, e.MethodExpr.StreamingPayload, e, false, sd)
			if cliPayload != nil {
				sd.ClientTypeNames[cliPayload.Name] = false
				sd.ServerTypeNames[cliPayload.Name] = false
			}
			if e.MethodExpr.Stream == expr.ClientStreamKind {
				svrSendDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection and closes the connection.", md.ServerStream.SendName, svrSendTypeName, md.Name)
				cliRecvDesc = fmt.Sprintf("%s stops sending messages to the %q endpoint websocket connection and reads instances of %q from the connection.", md.ClientStream.RecvName, md.Name, svrSendTypeName)
			}
			svrRecvDesc = fmt.Sprintf("%s reads instances of %q from the %q endpoint websocket connection.", md.ServerStream.RecvName, svrRecvTypeName, md.Name)
			cliSendDesc = fmt.Sprintf("%s streams instances of %q to the %q endpoint websocket connection.", md.ClientStream.SendName, svrRecvTypeName, md.Name)
		}
	}
	ed.ServerWebSocket = &WebSocketData{
		VarName:           md.ServerStream.VarName,
		Interface:         fmt.Sprintf("%s.%s", svc.PkgName, md.ServerStream.Interface),
		Endpoint:          ed,
		Payload:           svrPayload,
		Response:          ed.Result.Responses[0],
		PkgName:           svc.PkgName,
		Type:              "server",
		Kind:              md.ServerStream.Kind,
		SendName:          md.ServerStream.SendName,
		SendDesc:          svrSendDesc,
		SendTypeName:      svrSendTypeName,
		SendTypeRef:       svrSendTypeRef,
		RecvName:          md.ServerStream.RecvName,
		RecvDesc:          svrRecvDesc,
		RecvTypeName:      svrRecvTypeName,
		RecvTypeRef:       svrRecvTypeRef,
		RecvTypeIsPointer: expr.IsArray(e.MethodExpr.StreamingPayload.Type) || expr.IsMap(e.MethodExpr.StreamingPayload.Type),
		MustClose:         md.ServerStream.MustClose,
	}
	ed.ClientWebSocket = &WebSocketData{
		VarName:      md.ClientStream.VarName,
		Interface:    fmt.Sprintf("%s.%s", svc.PkgName, md.ClientStream.Interface),
		Endpoint:     ed,
		Payload:      cliPayload,
		Response:     ed.Result.Responses[0],
		PkgName:      svc.PkgName,
		Type:         "client",
		Kind:         md.ClientStream.Kind,
		SendName:     md.ClientStream.SendName,
		SendDesc:     cliSendDesc,
		SendTypeName: svrRecvTypeName,
		SendTypeRef:  svrRecvTypeRef,
		RecvName:     md.ClientStream.RecvName,
		RecvDesc:     cliRecvDesc,
		RecvTypeName: svrSendTypeName,
		RecvTypeRef:  svrSendTypeRef,
		MustClose:    md.ClientStream.MustClose,
	}
}

// websocketServerFile returns the file implementing the WebSocket server
// streaming implementation if any.
func websocketServerFile(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	data := HTTPServices.Get(svc.Name())
	if !hasWebSocket(data) {
		return nil
	}
	svcName := data.Service.PathName
	title := fmt.Sprintf("%s WebSocket server streaming", svc.Name())
	imports := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "sync"},
		{Path: "time"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
	}
	sections = append(sections, serverStructWSSections(data)...)
	sections = append(sections, serverWSSections(data)...)

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "http", svcName, "server", "websocket.go"),
		SectionTemplates: sections,
	}
}

// websocketClientFile returns the file implementing the WebSocket client
// streaming implementation if any.
func websocketClientFile(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	data := HTTPServices.Get(svc.Name())
	if !hasWebSocket(data) {
		return nil
	}
	svcName := data.Service.PathName
	title := fmt.Sprintf("%s WebSocket client streaming", svc.Name())
	imports := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "sync"},
		{Path: "time"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", imports),
	}
	sections = append(sections, clientStructWSSections(data)...)
	sections = append(sections, clientWSSections(data)...)

	return &codegen.File{
		Path:             filepath.Join(codegen.Gendir, "http", svcName, "client", "websocket.go"),
		SectionTemplates: sections,
	}
}

// serverStructWSSections return section templates that generate WebSocket
// related struct type definitions for the server.
func serverStructWSSections(data *ServiceData) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "server-websocket-conn-configurer-struct",
		Source:  readTemplate("websocket_conn_configurer_struct"),
		Data:    data,
		FuncMap: map[string]any{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ServerWebSocket != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-websocket-struct-type",
				Source: readTemplate("websocket_struct_type"),
				Data:   e.ServerWebSocket,
			})
		}
	}

	return sections
}

// serverWSSections returns section templates that contain server WebSocket
// specific code for the given service.
func serverWSSections(data *ServiceData) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "server-websocket-conn-configurer-struct-init",
		Source:  readTemplate("websocket_conn_configurer_struct_init"),
		Data:    data,
		FuncMap: map[string]any{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ServerWebSocket != nil {
			if e.ServerWebSocket.SendTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-send",
					Source: readTemplate("websocket_send", "websocket_upgrade"),
					Data:   e.ServerWebSocket,
					FuncMap: map[string]any{
						"upgradeParams":    upgradeParams,
						"viewedServerBody": viewedServerBody,
					},
				})
			}
			switch e.ServerWebSocket.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "server-websocket-recv",
					Source:  readTemplate("websocket_recv", "websocket_upgrade"),
					Data:    e.ServerWebSocket,
					FuncMap: map[string]any{"upgradeParams": upgradeParams},
				})
			}
			if e.ServerWebSocket.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "server-websocket-close",
					Source:  readTemplate("websocket_close"),
					Data:    e.ServerWebSocket,
					FuncMap: map[string]any{"upgradeParams": upgradeParams},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-set-view",
					Source: readTemplate("websocket_set_view"),
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
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "client-websocket-conn-configurer-struct",
		Source:  readTemplate("websocket_conn_configurer_struct"),
		Data:    data,
		FuncMap: map[string]any{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ClientWebSocket != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-websocket-struct-type",
				Source: readTemplate("websocket_struct_type"),
				Data:   e.ClientWebSocket,
			})
		}
	}
	return sections
}

// clientWSSections returns section templates that contain client WebSocket
// specific code for the given service.
func clientWSSections(data *ServiceData) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	sections = append(sections, &codegen.SectionTemplate{
		Name:    "client-websocket-conn-configurer-struct-init",
		Source:  readTemplate("websocket_conn_configurer_struct_init"),
		Data:    data,
		FuncMap: map[string]any{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ClientWebSocket != nil {
			if e.ClientWebSocket.RecvTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "client-websocket-recv",
					Source:  readTemplate("websocket_recv", "websocket_upgrade"),
					Data:    e.ClientWebSocket,
					FuncMap: map[string]any{"upgradeParams": upgradeParams},
				})
			}
			switch e.ClientWebSocket.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-websocket-send",
					Source: readTemplate("websocket_send", "websocket_upgrade"),
					Data:   e.ClientWebSocket,
					FuncMap: map[string]any{
						"upgradeParams":    upgradeParams,
						"viewedServerBody": viewedServerBody,
					},
				})
			}
			if e.ClientWebSocket.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "client-websocket-close",
					Source:  readTemplate("websocket_close"),
					Data:    e.ClientWebSocket,
					FuncMap: map[string]any{"upgradeParams": upgradeParams},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-websocket-set-view",
					Source: readTemplate("websocket_set_view"),
					Data:   e.ClientWebSocket,
				})
			}
		}
	}
	return sections
}

// hasWebSocket returns true if at least one of the endpoints in the service
// defines a streaming payload or result.
func hasWebSocket(sd *ServiceData) bool {
	for _, e := range sd.Endpoints {
		if isWebSocketEndpoint(e) {
			return true
		}
	}
	return false
}

// isWebSocketEndpoint returns true if the endpoint defines a streaming payload
// or result.
func isWebSocketEndpoint(ed *EndpointData) bool {
	return ed.ServerWebSocket != nil || ed.ClientWebSocket != nil
}
