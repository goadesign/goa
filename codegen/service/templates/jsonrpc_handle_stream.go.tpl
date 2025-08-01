// HandleStream handles the JSON-RPC WebSocket connection.
func (s *{{ .VarName }}srvc) HandleStream(ctx context.Context, stream {{ .PkgName }}.Stream) error {
	log.Printf(ctx, "{{ .VarName }}.HandleStream")
	defer stream.Close()
	
	// TODO: For server streaming methods with no payload, you may want to 
	// initiate streaming upon connection. For example:
	//
	// go func() {
	//     // Listen to a channel, timer, or other event source
	//     for data := range yourDataChannel {
	//         if err := stream.SendYourMethod(ctx, data); err != nil {
	//             log.Printf(ctx, "streaming error: %v", err)
	//             return
	//         }
	//     }
	// }()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Recv automatically dispatches JSON-RPC requests to your service methods
			// and sends responses back through the WebSocket connection
			err := stream.Recv(ctx)
			if err != nil {
				return err
			}
		}
	}
}