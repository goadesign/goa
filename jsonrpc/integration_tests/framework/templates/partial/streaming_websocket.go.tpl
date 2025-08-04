{{- /* Template for WebSocket streaming method implementation */ -}}
{{- if and (eq .Info.Action "collect") (eq .Info.Type "array") -}}
	// For JSON-RPC WebSocket, each request comes as a separate call to this method
	// We accumulate items across requests using service-level state
	if p != nil && p.Items != nil {
		s.collectedItems = append(s.collectedItems, p.Items...)
	}
	
	// Return the accumulated items
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		ID:    p.ID,
		Items: s.collectedItems,
	}
	if err := stream.Send(ctx, result); err != nil {
		return err
	}
	
	return nil
{{- else if .IsBidirectional -}}
	{{- if eq .Info.Type "string" -}}
	// For JSON-RPC WebSocket, each request comes as a separate call
	// Echo back the received payload
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:    p.ID,
			Value: p.Value,
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	return nil
	{{- else if eq .Info.Type "object" -}}
	// For JSON-RPC WebSocket, each request comes as a separate call
	// Echo back the received payload
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:     p.ID,
			Field1: p.Field1,
			Field2: p.Field2,
			Field3: p.Field3,
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	return nil
	{{- else -}}
	// For JSON-RPC WebSocket, each request comes as a separate call
	// Echo back the received payload
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:   p.ID,
			Data: p.Data,
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	return nil
	{{- end -}}
{{- else if eq .Info.Action "broadcast" -}}
	// Broadcast messages to client
	for i := 1; i <= 3; i++ {
		{{- if eq .Info.Type "string" -}}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:    fmt.Sprintf("broadcast-%d", i),
			Value: fmt.Sprintf("Server announcement %d", i),
		}
		{{- else -}}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID: fmt.Sprintf("broadcast-%d", i),
		}
		{{- end }}
		if err := stream.Send(result); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
{{- else -}}
	// Default WebSocket implementation for JSON-RPC
	// Each request comes as a separate call
	if p != nil {
		// Process payload and send response
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	return nil
{{- end -}}