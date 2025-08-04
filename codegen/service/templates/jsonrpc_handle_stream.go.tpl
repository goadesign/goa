// HandleStream manages a JSON-RPC WebSocket connection, enabling bidirectional
// communication between the server and client. It receives requests from the
// client, dispatches them to the appropriate service methods, and can send
// server-initiated messages back to the client as needed.
func (s *{{ .VarName }}srvc) HandleStream(ctx context.Context, stream {{ .PkgName }}.Stream) error {
	log.Printf(ctx, "{{ .VarName }}.HandleStream")

	// Close the stream when the function returns
	defer stream.Close()
	
	// To initiate server-side streaming, send messages to the client using
	// stream.Send(ctx, data) as needed. For example, you can launch a goroutine
	// that listens to an event source and sends updates to the client.
	//
	// go func() {
	//     // Listen to a channel, timer, or other event source
	//     for data := range yourDataChannel {
	//         if err := stream.Send(ctx, data); err != nil {
	//             log.Printf(ctx, "streaming error: %v", err)
	//             return
	//         }
	//     }
	// }()
	
	// Continuously receive JSON-RPC requests from the client and
	// automatically route them to the appropriate service methods.
	// Each request is handled according to its method name and parameters.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := stream.Recv(ctx)
			if err != nil {
				return err
			}
		}
	}
}