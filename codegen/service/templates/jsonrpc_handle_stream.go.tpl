// HandleStream manages a JSON-RPC WebSocket connection, enabling bidirectional
// communication between the server and client. It receives requests from the
// client, dispatches them to the appropriate service methods, and can send
// server-initiated messages back to the client as needed.
func (s *{{ .VarName }}srvc) HandleStream(ctx context.Context, stream {{ .PkgName }}.Stream) error {
    log.Printf(ctx, "{{ .VarName }}.HandleStream")

    // Example: In a real implementation you might read from an event source
    // and send notifications via stream.Send(ctx, event). This stub returns
    // when the context is canceled.
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        return nil
    }
}
