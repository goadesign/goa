{{ printf "Close asks the %s client to close normally, closes the WebSocket, and returns errors from either operation." .Service.Name | comment }}
func (s *{{ websocketServerStreamName }}) Close() error {
	controlErr := s.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second),
	)
	if controlErr != nil {
		controlErr = fmt.Errorf("write normal WebSocket close message: %w", controlErr)
	}
	closeErr := s.conn.Close()
	if closeErr != nil {
		closeErr = fmt.Errorf("close WebSocket connection: %w", closeErr)
	}
	return errors.Join(controlErr, closeErr)
}
