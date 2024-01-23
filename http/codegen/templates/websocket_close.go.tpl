{{ printf "Close closes the %q endpoint websocket connection." .Endpoint.Method.Name | comment }}
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
