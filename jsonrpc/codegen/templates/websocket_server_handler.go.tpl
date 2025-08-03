// ServeHTTP handles WebSocket JSON-RPC requests.
func (s *{{ .ServerStruct }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.errhandler(r.Context(), w, fmt.Errorf("failed to upgrade to WebSocket: %w", err))
		cancel()
		return
	}
	if s.configfn != nil {
		conn = s.configfn(conn, cancel)
	}
	defer conn.Close()

	stream := &{{ lowerInitial .Service.StructName }}Stream{
	{{- range .Endpoints }}
		{{ lowerInitial .Method.VarName }}: s.{{ lowerInitial .Method.VarName }},
		{{- if and .Method.ServerStream (or (eq .Method.ServerStream.Kind 3) (eq .Method.ServerStream.Kind 4)) }}
		{{ lowerInitial .Method.VarName }}Endpoint: s.{{ lowerInitial .Method.VarName }}Endpoint,
		{{- end }}
	{{- end }}
		r: r,
		w: w,
		conn: conn,
		cancel: cancel,
	}
	s.StreamHandler(ctx, stream)
} 
