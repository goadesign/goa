{{- /* Template for SSE streaming method implementation */ -}}
{{- /* SSE streaming behavior is determined by the action, not hardcoded modifiers */ -}}

{{- if eq .Info.Action "echo" -}}
	{{- /* Echo action: Stream back the payload as notifications */ -}}
	{{- if .HasPayload -}}
	{{- if eq .Info.Type "string" -}}
	// Echo the string value back as a notification
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Value: p,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "array" -}}
	// Echo each array item as a separate notification
	for _, item := range p.Items {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Items: []string{item},
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "object" -}}
	// Echo the object back as a notification
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "map" -}}
	// Echo the map back as a notification
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Data: p.Data,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- end -}}
	{{- else -}}
	// Echo method requires payload
	return fmt.Errorf("echo method requires payload")
	{{- end -}}

{{- else if eq .Info.Action "transform" -}}
	{{- /* Transform action: Stream transformed versions of the payload */ -}}
	{{- if .HasPayload -}}
	{{- if eq .Info.Type "string" -}}
	// Transform and stream: uppercase
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Value: strings.ToUpper(p),
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "array" -}}
	// Transform and stream: reverse the array
	reversed := make([]string, len(p.Items))
	for i, item := range p.Items {
		reversed[len(p.Items)-1-i] = item
	}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Items: reversed,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "object" -}}
	// Transform and stream: uppercase field1, double field2, negate field3
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: strings.ToUpper(p.Field1),
		Field2: p.Field2 * 2,
		Field3: !p.Field3,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "map" -}}
	// Transform and stream: prefix all keys with "transformed_"
	transformed := make(map[string]any)
	for k, v := range p.Data {
		transformed["transformed_"+k] = v
	}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Data: transformed,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- end -}}
	{{- else -}}
	// Transform method requires payload
	return fmt.Errorf("transform method requires payload")
	{{- end -}}

{{- else if eq .Info.Action "generate" -}}
	{{- /* Generate action: Stream generated values, ignoring payload */ -}}
	{{- if eq .Info.Type "string" -}}
	// Generate and stream string values
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: fmt.Sprintf("generated-%d", i),
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "array" -}}
	// Generate and stream array values
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Items: []string{fmt.Sprintf("item-%d", i)},
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "object" -}}
	// Generate and stream object values
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Field1: fmt.Sprintf("generated-%d", i),
			Field2: i * 10,
			Field3: i%2 == 0,
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	{{- else if eq .Info.Type "map" -}}
	// Generate and stream map values
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Data: map[string]any{
				"iteration": i,
				"status": fmt.Sprintf("step-%d", i),
			},
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	{{- end -}}

{{- else if eq .Info.Action "stream" -}}
	{{- /* Stream action: Stream a series of notifications based on payload */ -}}
	{{- if eq .Info.Type "string" -}}
	// Stream notifications based on the string payload
	// The payload determines how many messages to send
	count := 3 // default
	{{- if .HasPayload }}
	if p != "" {
		// Use the length of the string as a hint for count (max 10)
		count = len(p)
		if count > 10 {
			count = 10
		}
	}
	{{- end }}
	for i := 1; i <= count; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: fmt.Sprintf("Stream %d of %d", i, count),
		}
		if err := stream.Send(result); err != nil {
			return err
		}
		// Small delay to simulate streaming
		time.Sleep(10 * time.Millisecond)
	}
	{{- else if eq .Info.Type "array" -}}
	// Stream notifications for each item in the array
	{{- if .HasPayload }}
	if len(p.Items) == 0 {
		// If empty, send a single notification
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Items: []string{"empty"},
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	} else {
		// Stream each item as a separate notification
		for i, item := range p.Items {
			result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
				Items: []string{fmt.Sprintf("Processing: %s", item)},
			}
			if err := stream.Send(result); err != nil {
				return err
			}
			// Small delay between items
			if i < len(p.Items)-1 {
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
	{{- else }}
	// Default without payload
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Items: []string{"stream-1", "stream-2", "stream-3"},
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- end }}
	{{- else if eq .Info.Type "object" -}}
	// Stream notifications based on object fields
	// Use Field2 as iteration count (default 3, max 10)
	count := 3  // default
	{{- if .HasPayload }}
	count = p.Field2
	if count <= 0 {
		count = 3
	}
	if count > 10 {
		count = 10
	}
	{{- end }}
	for i := 1; i <= count; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			{{- if .HasPayload }}
			Field1: fmt.Sprintf("%s-%d", p.Field1, i),
			{{- else }}
			Field1: fmt.Sprintf("stream-%d", i),
			{{- end }}
			Field2: i,
			Field3: i == count, // true for last item
		}
		if err := stream.Send(result); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	{{- else if eq .Info.Type "map" -}}
	// Stream notifications for each key-value pair
	{{- if .HasPayload }}
	if len(p.Data) == 0 {
		// If empty, send a single notification
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Data: map[string]any{"status": "empty"},
		}
		if err := stream.Send(result); err != nil {
			return err
		}
	} else {
		// Stream each key-value pair as a separate notification
		// Sort keys for deterministic ordering
		keys := make([]string, 0, len(p.Data))
		for k := range p.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		for _, k := range keys {
			v := p.Data[k]
			result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
				Data: map[string]any{
					"key": k,
					"value": v,
				},
			}
			if err := stream.Send(result); err != nil {
				return err
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	{{- else }}
	// Default without payload - stream 3 items
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Data: map[string]any{
				"iteration": i,
				"status": fmt.Sprintf("step-%d", i),
			},
		}
		if err := stream.Send(result); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	{{- end }}
	{{- end -}}

{{- else -}}
	{{- /* Default: echo behavior for unknown actions */ -}}
	// Default behavior: echo the payload if available
	{{- if .HasPayload -}}
	{{- if eq .Info.Type "string" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Value: p,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "array" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Items: p.Items,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "object" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: p.Field1,
		Field2: p.Field2,
		Field3: p.Field3,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "map" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Data: p.Data,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- end -}}
	{{- else -}}
	// No payload to echo, send default notification
	{{- if eq .Info.Type "string" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Value: "default",
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "array" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Items: []string{"default"},
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "object" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Field1: "default",
		Field2: 0,
		Field3: false,
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- else if eq .Info.Type "map" -}}
	result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
		Data: map[string]any{"status": "default"},
	}
	if err := stream.Send(result); err != nil {
		return err
	}
	{{- end -}}
	{{- end -}}
{{- end }}

{{ if eq .Info.Modifier "error" -}}
	// Return an error after streaming.
	return &goa.ServiceError{Message: "Streaming error occurred"}
{{ else -}}
	return nil
{{- end -}}
