{{- if needDialer .Services }}
	var (
    dialer *websocket.Dialer
  )
  {
    dialer = websocket.DefaultDialer
  }
	{{ end }}
