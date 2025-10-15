{{- /* Template for WebSocket streaming method implementation */ -}}
{{- /* WebSocket behavior is determined by the action, similar to SSE */ -}}

{{- /* Handle validation first if validate modifier is set */ -}}
{{- if eq .Info.Modifier "validate" }}
	{{- if eq .Info.Type "object" }}
	if p != nil && (p.Field1 == "" || p.Field2 < 0) {
		validationErr := &goa.ServiceError{
			Name:    "validation_error",
			Message: "validation error",
		}
		if err := stream.SendError(ctx, validationErr); err != nil {
			return err
		}
		return nil
	}
	{{- else if eq .Info.Type "string" }}
	if p != nil && p.Value == "" {
		validationErr := &goa.ServiceError{
			Name:    "validation_error", 
			Message: "validation error",
		}
		if err := stream.SendError(ctx, validationErr); err != nil {
			return err
		}
		return nil
	}
	{{- end }}
{{- end }}

{{- /* For echo with error modifier, return error immediately */ -}}
{{- if and (eq .Info.Action "echo") (eq .Info.Modifier "error") }}
	// For echo methods with error modifier, always send error response
	testErr := &goa.ServiceError{
		Name:    "test_error",
		Message: "Invalid params",
	}
	if err := stream.SendError(ctx, testErr); err != nil {
		return err
	}
	return nil
{{- else if eq .Info.Action "echo" }}
	{{- /* Echo action: Return the payload exactly as received */ -}}
	{{- if .IsBidirectional }}
	{{- if eq .Info.Type "string" }}
	// Echo back the received payload
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: p.Value,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "array" }}
	// Echo back the received array
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Items: p.Items,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "object" }}
	// Echo back the received object
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Field1: p.Field1,
			Field2: p.Field2,
			Field3: p.Field3,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "map" }}
	// Echo back the received map
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Data: p.Data,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- end }}
	return nil
	{{- else }}
	// Non-bidirectional echo - shouldn't happen for WebSocket
	return fmt.Errorf("echo action requires bidirectional streaming")
	{{- end }}

{{- else if eq .Info.Action "transform" }}
	{{- /* Transform action: Apply transformations to the payload */ -}}
	{{- if .IsBidirectional }}
	{{- if eq .Info.Type "string" }}
	// Transform and return: uppercase
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: strings.ToUpper(p.Value),
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "array" }}
	// Transform and return: reverse the array
	if p != nil {
		reversed := make([]string, len(p.Items))
		for i, item := range p.Items {
			reversed[len(p.Items)-1-i] = item
		}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Items: reversed,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "object" }}
	// Transform and return: uppercase field1, double field2, negate field3
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Field1: strings.ToUpper(p.Field1),
			Field2: p.Field2 * 2,
			Field3: !p.Field3,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "map" }}
	// Transform and return: prefix all keys with "transformed_"
	if p != nil && p.Data != nil {
		transformed := make(map[string]any)
		if data, ok := p.Data.(map[string]any); ok {
			for k, v := range data {
				transformed["transformed_"+k] = v
			}
		}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Data: transformed,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- end }}
	return nil
	{{- else }}
	return fmt.Errorf("transform action requires bidirectional streaming")
	{{- end }}

{{- else if eq .Info.Action "generate" }}
	{{- /* Generate action: Return fixed values, ignoring payload */ -}}
	{{- if .IsBidirectional }}
	{{- if eq .Info.Type "string" }}
	// Generate and return fixed string
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: "generated-string",
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "array" }}
	// Generate and return fixed array
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Items: []string{"item1", "item2", "item3"},
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "object" }}
	// Generate and return fixed object
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Field1: "generated-value1",
			Field2: 42,
			Field3: true,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "map" }}
	// Generate and return fixed map
	if p != nil {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID: p.ID,
			Data: map[string]any{
				"generated": true,
				"count":     3,
				"status":    "ok",
			},
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	{{- end }}
	return nil
	{{- else }}
	// Server-initiated generation (no client request)
	{{- if eq .Info.Type "string" }}
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: fmt.Sprintf("generated-%d", i),
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	return nil
	{{- else }}
	// Generate default values
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{}
	if err := stream.SendResponse(ctx, result); err != nil {
		return err
	}
	return nil
	{{- end }}
	{{- end }}

{{- else if eq .Info.Action "stream" }}
	{{- /* Stream action: Send a series of messages based on payload */ -}}
	{{- if .IsBidirectional }}
	{{- if eq .Info.Type "string" }}
	// Stream notifications based on string payload
	if p != nil {
		count := 3 // default
		if p.Value != "" {
			count = len(p.Value)
			if count > 10 {
				count = 10
			}
		}
		
		// For error modifier, send fewer notifications
		streamCount := count
		{{- if eq .Info.Modifier "error" }}
		if streamCount > 2 {
			streamCount = 2
		}
		{{- end }}
		
		// Send notifications without ID
		for i := 1; i <= streamCount; i++ {
			notification := &{{ $.ServicePackage }}.{{ .GoName }}Result{
				Value: fmt.Sprintf("Stream %d of %d", i, count),
			}
			if err := stream.SendNotification(ctx, notification); err != nil {
				return err
			}
		}
		{{- if ne .Info.Modifier "error" }}
		// Send final response with ID
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: "completed",
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
		{{- end }}
	}
	{{- else if eq .Info.Type "array" }}
	// Stream notifications for each array item
	if p != nil {
		// Send notification for each item
		for _, item := range p.Items {
			notification := &{{ $.ServicePackage }}.{{ .GoName }}Result{
				Items: []string{fmt.Sprintf("Processing: %s", item)},
			}
			if err := stream.SendNotification(ctx, notification); err != nil {
				return err
			}
		}
		{{- if ne .Info.Modifier "error" }}
		// Send final response with ID
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Items: []string{"completed"},
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
		{{- end }}
	}
	{{- else if eq .Info.Type "object" }}
	// Stream notifications based on field2 count
	if p != nil {
		count := p.Field2
		if count <= 0 {
			count = 3
		}
		if count > 10 {
			count = 10
		}
		// Send notifications without ID
		for i := 1; i <= count; i++ {
			notification := &{{ $.ServicePackage }}.{{ .GoName }}Result{
				Field1: fmt.Sprintf("%s-%d", p.Field1, i),
				Field2: i,
				Field3: i == count,
			}
			if err := stream.SendNotification(ctx, notification); err != nil {
				return err
			}
		}
		{{- if ne .Info.Modifier "error" }}
		// Send final response with ID
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Field1: "completed",
			Field2: 100,
			Field3: true,
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
		{{- end }}
	}
	{{- else if eq .Info.Type "map" }}
	// Stream notifications for each key-value pair
	if p != nil && p.Data != nil {
		// Send notification for each pair
		if data, ok := p.Data.(map[string]any); ok {
			// Sort keys for deterministic ordering
			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			
			for _, k := range keys {
				v := data[k]
				notification := &{{ $.ServicePackage }}.{{ .GoName }}Result{
					Data: map[string]any{
						"key":   k,
						"value": v,
					},
				}
				if err := stream.SendNotification(ctx, notification); err != nil {
					return err
				}
			}
		}
		{{- if ne .Info.Modifier "error" }}
		// Send final response with ID
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Data: map[string]any{
				"status": "completed",
			},
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
		{{- end }}
	}
	{{- end }}
	{{- if ne .Info.Modifier "error" }}
	return nil
	{{- end }}
	{{- else }}
	// Server-side streaming without client payload
	return fmt.Errorf("stream action requires bidirectional streaming for WebSocket")
	{{- end }}

{{- else if eq .Info.Action "collect" }}
	{{- /* Collect action: Accumulate client messages */ -}}
	{{- if eq .Info.Type "array" }}
	// For JSON-RPC WebSocket, each request comes as a separate call to this method
	// We accumulate items across requests using service-level state
	if p != nil && p.Items != nil {
		s.collectedItems = append(s.collectedItems, p.Items...)
	}
	
	// Return the accumulated items
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Items: s.collectedItems,
	}
	if err := stream.SendResponse(ctx, result); err != nil {
		return err
	}
	return nil
	{{- else }}
	// Collect only supports array type currently
	return fmt.Errorf("collect action only supports array type")
	{{- end }}

{{- else if eq .Info.Action "broadcast" }}
	{{- /* Broadcast action: Server-initiated messages */ -}}
	{{- if .IsBidirectional }}
	// Broadcast method implementation - bidirectional streaming
	// When called (e.g., with "subscribe"), send test broadcast notifications
	{{- if eq .Info.Type "string" }}
	// Send test broadcasts
	for i := 1; i <= 2; i++ {
		result := &{{ .ServicePackage }}.{{ .GoName }}Result{
			Value: fmt.Sprintf("Server announcement %d", i),
		}
		if err := stream.SendNotification(ctx, result); err != nil {
			return err
		}
	}
	return nil
	{{- else if eq .Info.Type "array" }}
	// Send test array broadcasts
	for i := 1; i <= 2; i++ {
		result := &{{ .ServicePackage }}.{{ .GoName }}Result{
			Items: []string{fmt.Sprintf("broadcast-%d", i)},
		}
		if err := stream.SendNotification(ctx, result); err != nil {
			return err
		}
	}
	return nil
	{{- else if eq .Info.Type "object" }}
	// Send test object broadcasts
	for i := 1; i <= 2; i++ {
		result := &{{ .ServicePackage }}.{{ .GoName }}Result{
			Field1: fmt.Sprintf("broadcast-%d", i),
			Field2: i,
			Field3: i%2 == 0,
		}
		if err := stream.SendNotification(ctx, result); err != nil {
			return err
		}
	}
	return nil
	{{- else if eq .Info.Type "map" }}
	// Send test map broadcasts
	for i := 1; i <= 2; i++ {
		result := &{{ .ServicePackage }}.{{ .GoName }}Result{
			Data: map[string]any{
				"broadcast": i,
				"timestamp": time.Now().Unix(),
			},
		}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
	return nil
	{{- else }}
	// Send default broadcasts
	for i := 1; i <= 2; i++ {
		result := &{{ .ServicePackage }}.{{ .GoName }}Result{}
		if err := stream.SendNotification(ctx, result); err != nil {
			return err
		}
	}
	return nil
	{{- end }}
	{{- else }}
	// Broadcast action requires bidirectional streaming
	return fmt.Errorf("broadcast action requires bidirectional streaming")
	{{- end }}

{{- else }}
	{{- /* Default: echo behavior for unknown actions */ -}}
	// Default WebSocket implementation for JSON-RPC
	// Each request comes as a separate call
	if p != nil {
		// Echo payload back
		{{- if eq .Info.Type "string" }}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:    p.ID,
			Value: p.Value,
		}
		{{- else if eq .Info.Type "array" }}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:    p.ID,
			Items: p.Items,
		}
		{{- else if eq .Info.Type "object" }}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:     p.ID,
			Field1: p.Field1,
			Field2: p.Field2,
			Field3: p.Field3,
		}
		{{- else if eq .Info.Type "map" }}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			ID:   p.ID,
			Data: p.Data,
		}
		{{- else }}
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{}
		{{- end }}
		if err := stream.SendResponse(ctx, result); err != nil {
			return err
		}
	}
{{- end }}

{{- /* Handle remaining modifiers */ -}}
{{- if eq .Info.Modifier "error" }}
	{{- /* For stream action with error, send error after streaming */ -}}
	{{- if eq .Info.Action "stream" }}
	// For stream methods with error modifier, send error after streaming
	testErr := &goa.ServiceError{
		Name:    "test_error",
		Message: "Streaming error occurred",
	}
	if err := stream.SendError(ctx, testErr); err != nil {
		return err
	}
	return nil
	{{- else }}
	// Other actions with error modifier should have been handled above
	return nil
	{{- end }}
{{- else if eq .Info.Modifier "notify" }}
	// Notification - no response sent (already handled above)
	return nil
{{- else }}
	// Normal completion
	return nil
{{- end }}