{{ printf "Close closes the %q endpoint websocket connection." .Endpoint.Method.Name | comment }}
func (s *{{ .VarDeclaration.Name }}) Close() error {
{{- if eq .Type "server" }}
	s.closeOnce.Do(func() {
		s.closeErr = s.close()
	})
	return s.closeErr
}

{{ comment "close opens the websocket connection when needed, sends its normal close message, and closes it." }}
func (s *{{ .VarDeclaration.Name }}) close() error {
	var err error
	{{- if and .Endpoint.Method.ViewedResult (not .Endpoint.Method.ViewedResult.ViewName) }}
	view := s.view
	if view == "" {
		view = "default"
	}
	switch view {
	{{- range .Endpoint.Method.ViewedResult.Views }}
	case {{ printf "%q" .Name }}:
	{{- end }}
	default:
		return goa.InvalidEnumValueError("view", view, []any{ {{ range .Endpoint.Method.ViewedResult.Views }}{{ printf "%q" .Name }}, {{ end }} })
	}
	{{- end }}
	{{- template "partial_websocket_upgrade" (upgradeParams .Endpoint "Close") }}
	if err = s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server closing connection"),
		time.Now().Add(time.Second),
	); err != nil {
		return err
	}
	return s.conn.Close()
{{- else }} {{/* client side code */}}
	var err error
	{{ comment "Send a nil payload to the server implying client closing connection." }}
  if err = s.conn.WriteJSON(nil); err != nil {
    return err
  }
	return s.conn.Close()
{{- end }}
}
