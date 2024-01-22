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