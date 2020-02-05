package codegen

import (
	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/expr"
)

// serverStructWSSections return section templates that generate WebSocket
// related struct type definitions for the server.
func serverStructWSSections(data *ServiceData) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	if hasWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "server-stream-conn-configurer-struct",
			Source:  webSocketConnConfigurerStructT,
			Data:    data,
			FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
		})
	}
	for _, e := range data.Endpoints {
		if e.ServerStream != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "server-stream-struct-type",
				Source: webSocketStructTypeT,
				Data:   e.ServerStream,
			})
		}
	}

	return sections
}

// serverWSSections returns section templates that contain server WebSocket
// specific code for the given service.
func serverWSSections(data *ServiceData) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	if hasWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "server-stream-conn-configurer-struct-init",
			Source:  webSocketConnConfigurerStructInitT,
			Data:    data,
			FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
		})
	}
	for _, e := range data.Endpoints {
		if e.ServerStream != nil {
			if e.ServerStream.SendTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-stream-send",
					Source: webSocketSendT,
					Data:   e.ServerStream,
					FuncMap: map[string]interface{}{
						"upgradeParams":    upgradeParams,
						"viewedServerBody": viewedServerBody,
					},
				})
			}
			switch e.ServerStream.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "server-stream-recv",
					Source:  webSocketRecvT,
					Data:    e.ServerStream,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			if e.ServerStream.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "server-stream-close",
					Source:  webSocketCloseT,
					Data:    e.ServerStream,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "server-stream-set-view",
					Source: webSocketSetViewT,
					Data:   e.ServerStream,
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
	if hasWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "client-stream-conn-configurer-struct",
			Source:  webSocketConnConfigurerStructT,
			Data:    data,
			FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
		})
	}
	for _, e := range data.Endpoints {
		if e.ClientStream != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "client-stream-struct-type",
				Source: webSocketStructTypeT,
				Data:   e.ClientStream,
			})
		}
	}
	return sections
}

// clientWSSections returns section templates that contain client WebSocket
// specific code for the given service.
func clientWSSections(data *ServiceData) []*codegen.SectionTemplate {
	var sections []*codegen.SectionTemplate
	if hasWebSocket(data) {
		sections = append(sections, &codegen.SectionTemplate{
			Name:    "client-stream-conn-configurer-struct-init",
			Source:  webSocketConnConfigurerStructInitT,
			Data:    data,
			FuncMap: map[string]interface{}{"isWebSocketEndpoint": isWebSocketEndpoint},
		})
	}
	for _, e := range data.Endpoints {
		if e.ClientStream != nil {
			if e.ClientStream.RecvTypeRef != "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "client-stream-recv",
					Source:  webSocketRecvT,
					Data:    e.ClientStream,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			switch e.ClientStream.Kind {
			case expr.ClientStreamKind, expr.BidirectionalStreamKind:
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-send",
					Source: webSocketSendT,
					Data:   e.ClientStream,
					FuncMap: map[string]interface{}{
						"upgradeParams":    upgradeParams,
						"viewedServerBody": viewedServerBody,
					},
				})
			}
			if e.ClientStream.MustClose {
				sections = append(sections, &codegen.SectionTemplate{
					Name:    "client-stream-close",
					Source:  webSocketCloseT,
					Data:    e.ClientStream,
					FuncMap: map[string]interface{}{"upgradeParams": upgradeParams},
				})
			}
			if e.Method.ViewedResult != nil && e.Method.ViewedResult.ViewName == "" {
				sections = append(sections, &codegen.SectionTemplate{
					Name:   "client-stream-set-view",
					Source: webSocketSetViewT,
					Data:   e.ClientStream,
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
	return ed.ServerStream != nil || ed.ClientStream != nil
}

const (
	// webSocketStructTypeT renders the server and client struct types that
	// implements the client and server stream interfaces. The data to render
	// input: StreamData
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
	// input: StreamData
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
	// input: StreamData
	webSocketRecvT = `{{ comment .RecvDesc }}
func (s *{{ .VarName }}) {{ .RecvName }}() ({{ .RecvTypeRef }}, error) {
	var (
		rv {{ .RecvTypeRef }}
	{{- if eq .Type "server" }}
		msg *{{ .Payload.Ref }}
	{{- else }}
		body {{ .Response.ClientBody.VarName }}
	{{- end }}
		err error
	)
{{- if eq .Type "server" }}
	{{- template "websocket_upgrade" (upgradeParams .Endpoint .RecvName) }}
	if err = s.conn.ReadJSON(&msg); err != nil {
		return rv, err
	}
	if msg == nil {
		return rv, io.EOF
	}
	body := *msg
	{{- if .Payload.ValidateRef }}
		{{ .Payload.ValidateRef }}
		if err != nil {
			return rv, err
		}
	{{- end }}
	{{- if .Payload.Init }}
		return {{ .Payload.Init.Name }}(body), nil
	{{- else }}
		return body, nil
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
	// input: StreamData
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
	// input: StreamData
	webSocketSetViewT = `{{ printf "SetView sets the view to render the %s type before sending to the %q endpoint websocket connection." .SendTypeName .Endpoint.Method.Name | comment }}
func (s *{{ .VarName }}) SetView(view string) {
	s.view = view
}
`
)
