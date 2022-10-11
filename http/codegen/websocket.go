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
								VarName:  "body",
								TypeName: sd.Scope.GoTypeName(e.StreamingBody),
								TypeRef:  sd.Scope.GoTypeRef(e.StreamingBody),
								Type:     e.StreamingBody.Type,
								Required: true,
								Example:  e.Body.Example(expr.Root.API.Random()),
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
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "io"},
			{Path: "net/http"},
			{Path: "sync"},
			{Path: "time"},
			{Path: "github.com/gorilla/websocket"},
			codegen.GoaImport(""),
			codegen.GoaNamedImport("http", "goahttp"),
			{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
		}),
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
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "client", []*codegen.ImportSpec{
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
		}),
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
		Source:  webSocketConnConfigurerStructT,
		Data:    data,
		FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ServerWebSocket != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-websocket-struct-type",
				Source: webSocketStructTypeT,
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
		Source:  webSocketConnConfigurerStructInitT,
		Data:    data,
		FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ServerWebSocket != nil {
			if e.ServerWebSocket.SendTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-send",
					Source: webSocketSendT,
					Data:   e.ServerWebSocket,
					FuncMap: map[string]interface{}{
						"upgradeParams":    upgradeParams,
						"viewedServerBody": viewedServerBody,
					},
				})
			}
			switch e.ServerWebSocket.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "server-websocket-recv",
					Source:  webSocketRecvT,
					Data:    e.ServerWebSocket,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			if e.ServerWebSocket.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "server-websocket-close",
					Source:  webSocketCloseT,
					Data:    e.ServerWebSocket,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-websocket-set-view",
					Source: webSocketSetViewT,
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
		Source:  webSocketConnConfigurerStructT,
		Data:    data,
		FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ClientWebSocket != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-websocket-struct-type",
				Source: webSocketStructTypeT,
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
		Source:  webSocketConnConfigurerStructInitT,
		Data:    data,
		FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
	})
	for _, e := range data.Endpoints {
		if e.ClientWebSocket != nil {
			if e.ClientWebSocket.RecvTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "client-websocket-recv",
					Source:  webSocketRecvT,
					Data:    e.ClientWebSocket,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			switch e.ClientWebSocket.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-websocket-send",
					Source: webSocketSendT,
					Data:   e.ClientWebSocket,
					FuncMap: map[string]interface{}{
						"upgradeParams":    upgradeParams,
						"viewedServerBody": viewedServerBody,
					},
				})
			}
			if e.ClientWebSocket.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "client-websocket-close",
					Source:  webSocketCloseT,
					Data:    e.ClientWebSocket,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-websocket-set-view",
					Source: webSocketSetViewT,
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

const (
	// webSocketStructTypeT renders the server and client struct types that
	// implements the client and server stream interfaces. The data to render
	// input: WebSocketData
	webSocketStructTypeT = `{{ printf "%s implements the %s interface." .VarName .Interface | comment }}
type {{ .VarName }} struct {
{{- if eq .Type "server" }}
	once sync.Once
	{{ comment "upgrader is the websocket connection upgrader." }}
	upgrader goahttp.Upgrader
	{{ comment "configurer is the websocket connection configurer." }}
	configurer goahttp.ConnConfigureFunc
	{{ comment "cancel is the context cancellation function which cancels the request context when invoked." }}
	cancel context.CancelFunc
	{{ comment "w is the HTTP response writer used in upgrading the connection." }}
	w http.ResponseWriter
	{{ comment "r is the HTTP request." }}
	r *http.Request
{{- end }}
	{{ comment "conn is the underlying websocket connection." }}
	conn *websocket.Conn
	{{- if .Endpoint.Method.ViewedResult }}
		{{- if not .Endpoint.Method.ViewedResult.ViewName }}
	{{ printf "view is the view to render %s result type before sending to the websocket connection." .SendTypeName | comment }}
	view string
		{{- end }}
	{{- end }}
}
`

	// webSocketConnConfigurerStructT generates the struct type that holds the
	// websocket connection configurers for all the streaming endpoints in the
	// service.
	// input: ServiceData
	webSocketConnConfigurerStructT = `{{ printf "ConnConfigurer holds the websocket connection configurer functions for the streaming endpoints in %q service." .Service.Name | comment }}
type ConnConfigurer struct {
{{- range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
		{{ .Method.VarName }}Fn goahttp.ConnConfigureFunc
	{{- end }}
{{- end }}
}
`

	// webSocketConnConfigurerStructInitT generates the constructor function to
	// initialize the websocket connection configurer struct.
	// input: ServiceData
	webSocketConnConfigurerStructInitT = `{{ printf "NewConnConfigurer initializes the websocket connection configurer function with fn for all the streaming endpoints in %q service." .Service.Name | comment }}
func NewConnConfigurer(fn goahttp.ConnConfigureFunc) *ConnConfigurer {
	return &ConnConfigurer{
{{- range .Endpoints }}
	{{- if isWebSocketEndpoint . }}
		{{ .Method.VarName}}Fn: fn,
	{{- end }}
{{- end }}
	}
}
`

	// webSocketSendT renders the function implementing the Send method in
	// stream interface.
	// input: WebSocketData
	webSocketSendT = `{{ comment .SendDesc }}
func (s *{{ .VarName }}) {{ .SendName }}(v {{ .SendTypeRef }}) error {
{{- if eq .Type "server" }}
	{{- if eq .SendName "Send" }}
		var err error
		{{- template "websocket_upgrade" (upgradeParams .Endpoint .SendName) }}
	{{- else }} {{/* SendAndClose */}}
		defer s.conn.Close()
	{{- end }}
	{{- if .Endpoint.Method.ViewedResult }}
		{{- if .Endpoint.Method.ViewedResult.ViewName }}
			res := {{ .PkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(v, {{ printf "%q" .Endpoint.Method.ViewedResult.ViewName }})
		{{- else }}
			res := {{ .PkgName }}.{{ .Endpoint.Method.ViewedResult.Init.Name }}(v, s.view)
		{{- end }}
	{{- else }}
	res := v
	{{- end }}
	{{- $servBodyLen := len .Response.ServerBody }}
	{{- if gt $servBodyLen 0 }}
		{{- if (index .Response.ServerBody 0).Init }}
			{{- if .Endpoint.Method.ViewedResult }}
				{{- if .Endpoint.Method.ViewedResult.ViewName }}
					{{- $vsb := (viewedServerBody $.Response.ServerBody .Endpoint.Method.ViewedResult.ViewName) }}
					body := {{ $vsb.Init.Name }}({{ range $vsb.Init.ServerArgs }}{{ .Ref }}, {{ end }})
				{{- else }}
					var body interface{}
					switch s.view {
					{{- range .Endpoint.Method.ViewedResult.Views }}
						case {{ printf "%q" .Name }}{{ if eq .Name "default" }}, ""{{ end }}:
						{{- $vsb := (viewedServerBody $.Response.ServerBody .Name) }}
							body = {{ $vsb.Init.Name }}({{ range $vsb.Init.ServerArgs }}{{ .Ref }}, {{ end }})
						{{- end }}
					}
				{{- end }}
			{{- else }}
				body := {{ (index .Response.ServerBody 0).Init.Name }}({{ range (index .Response.ServerBody 0).Init.ServerArgs }}{{ .Ref }}, {{ end }})
			{{- end }}
			return s.conn.WriteJSON(body)
		{{- else }}
			return s.conn.WriteJSON(res)
		{{- end }}
	{{- else }}
		return s.conn.WriteJSON(res)
	{{- end }}
{{- else }}
	{{- if .Payload.Init }}
		body := {{ .Payload.Init.Name }}(v)
		return s.conn.WriteJSON(body)
	{{- else }}
		return s.conn.WriteJSON(v)
	{{- end }}
{{- end }}
}
` + upgradeT

	// webSocketRecvT renders the function implementing the Recv method in
	// stream interface.
	// input: WebSocketData
	webSocketRecvT = `{{ comment .RecvDesc }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	var (
		rv {{ .RecvTypeRef }}
	{{- if eq .Type "server" }}
		{{- if .RecvTypeIsPointer }}
		body {{ .Payload.VarName }}
		{{- else }}
		msg *{{ .Payload.VarName }}
		{{- end }}
	{{- else }}
		body {{ .Response.ClientBody.VarName }}
	{{- end }}
		err error
	)
{{- if eq .Type "server" }}
	{{- template "websocket_upgrade" (upgradeParams .Endpoint .RecvName) }}
	{{- if .RecvTypeIsPointer }}
	if err = s.conn.ReadJSON(&body); err != nil {
	{{- else }}
	if err = s.conn.ReadJSON(&msg); err != nil {
	{{- end }}
		return rv, err
	}
	{{- if .RecvTypeIsPointer }}
	if body == nil {
	{{- else }}
	if msg == nil {
	{{- end }}
		return rv, io.EOF
	}
	{{- if .Payload.ValidateRef }}
		{{- if not .RecvTypeIsPointer }}
	body := *msg
		{{- end }}
		{{ .Payload.ValidateRef }}
		if err != nil {
			return rv, err
		}
	{{- end }}
	{{- if .Payload.Init }}
		return {{ .Payload.Init.Name }}({{ if .RecvTypeIsPointer }}body{{ else }}msg{{ end }}), nil
	{{- else }}
		return {{ if .RecvTypeIsPointer }}body{{ else }}*msg{{ end }}, nil
	{{- end }}
{{- else }} {{/* client side code */}}
	{{- if eq .RecvName "CloseAndRecv" }}
		defer s.conn.Close()
		{{ comment "Send a nil payload to the server implying end of message" }}
		if err = s.conn.WriteJSON(nil); err != nil {
			return rv, err
		}
	{{- end }}
	err = s.conn.ReadJSON(&body)
	if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		{{- if not .MustClose }}
			s.conn.Close()
		{{- end }}
		return rv, io.EOF
	}
	if err != nil {
		return rv, err
	}
	{{- if and .Response.ClientBody.ValidateRef (not .Endpoint.Method.ViewedResult) }}
	{{ .Response.ClientBody.ValidateRef }}
	if err != nil {
		return rv, err
	}
	{{- end }}
	{{- if .Response.ResultInit }}
		res := {{ .Response.ResultInit.Name }}({{ range .Response.ResultInit.ClientArgs }}{{ .Ref }},{{ end }})
		{{- if .Endpoint.Method.ViewedResult }}{{ with .Endpoint.Method.ViewedResult }}
			vres := {{ if not .IsCollection }}&{{ end }}{{ .ViewsPkg }}.{{ .VarName }}{res, {{ if .ViewName }}{{ printf "%q" .ViewName }}{{ else }}s.view{{ end }} }
			if err := {{ .ViewsPkg }}.Validate{{ $.Endpoint.Method.Result }}(vres); err != nil {
				return rv, goahttp.ErrValidationError("{{ $.Endpoint.ServiceName }}", "{{ $.Endpoint.Method.Name }}", err)
			}
			return {{ $.PkgName }}.{{ .ResultInit.Name }}(vres){{ end }}, nil
		{{- else }}
			return res, nil
		{{- end }}
	{{- else }}
		return body, nil
	{{- end }}
{{- end }}
}
` + upgradeT

	// upgradeT renders the code to upgrade the HTTP connection to a gorilla
	// websocket connection.
	upgradeT = `{{- define "websocket_upgrade" }}
	{{ printf "Upgrade the HTTP connection to a websocket connection only once. Connection upgrade is done here so that authorization logic in the endpoint is executed before calling the actual service method which may call %s()." .Function | comment }}
	s.once.Do(func() {
	{{- if and .ViewedResult (eq .Function "Send") }}
		{{- if not .ViewedResult.ViewName }}
			respHdr := make(http.Header)
			respHdr.Add("goa-view", s.view)
		{{- end }}
	{{- end }}
		var conn *websocket.Conn
		{{- if eq .Function "Send" }}
			{{- if .ViewedResult }}
				{{- if not .ViewedResult.ViewName }}
					conn, err = s.upgrader.Upgrade(s.w, s.r, respHdr)
				{{- else }}
					conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
				{{- end }}
			{{- else }}
				conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
			{{- end }}
		{{- else }}
			conn, err = s.upgrader.Upgrade(s.w, s.r, nil)
		{{- end }}
		if err != nil {
			return
		}
		if s.configurer != nil {
			conn = s.configurer(conn, s.cancel)
		}
		s.conn = conn
	})
	if err != nil {
		return {{ if eq .Function "Recv" }}rv, {{ end }}err
	}
{{- end }}
`

	// webSocketCloseT renders the function implementing the Close method in
	// stream interface.
	// input: WebSocketData
	webSocketCloseT = `{{ printf "Close closes the %q endpoint websocket connection." .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}) Close() error {
	var err error
{{- if eq .Type "server" }}
	if s.conn == nil {
		return nil
	}
	if err = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
{{- else }} {{/* client side code */}}
	{{ comment "Send a nil payload to the server implying client closing connection." }}
  if err = s.conn.WriteJSON(nil); err != nil {
    return err
  }
{{- end }}
	return s.conn.Close()
}
` + upgradeT

	// webSocketSetViewT renders the function implementing the SetView method in
	// server stream interface.
	// input: WebSocketData
	webSocketSetViewT = `{{ printf "SetView sets the view to render the %s type before sending to the %q endpoint websocket connection." .SendTypeName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}) SetView(view string) {
	s.view = view
}
`
)
