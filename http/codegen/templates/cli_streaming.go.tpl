{{- if needStream .Services }}
	var (
    dialer *websocket.Dialer
  )
  {
    dialer = websocket.DefaultDialer
  }
	{{ end }}
