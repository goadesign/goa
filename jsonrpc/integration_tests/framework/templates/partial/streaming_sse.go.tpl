{{- /* Template for SSE streaming method implementation */ -}}
{{- if and (eq .Info.Modifier "final") (eq .Info.Type "string") -}}
	// Stream progress notifications
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: fmt.Sprintf("Progress: %d%%", i*25),
		}
		if err := stream.Send(ctx, result); err != nil {
			return err
		}
	}
	// Note: Due to a Goa bug, SSE doesn't properly check the ID field in results
	// The generated code always sends notifications even when ID is set
	// For now, we'll just send 3 progress notifications
	return nil
{{- else if eq .Info.Type "string" -}}
	// Stream string results as notifications
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Value: fmt.Sprintf("message-%d", i),
		}
		if err := stream.Send(ctx, result); err != nil {
			return err
		}
	}
	return nil
{{- else if eq .Info.Type "object" -}}
	// Stream object results as notifications
	for i := 1; i <= 3; i++ {
		result := &{{ $.ServicePackage }}.{{ .GoName }}Result{
			Field1: "notification",
			Field2: i,
			Field3: i == 3,
		}
		if err := stream.Send(ctx, result); err != nil {
			return err
		}
	}
	return nil
{{- else -}}
	// Stream results as notifications
	for i := 1; i <= 3; i++ {
		if err := stream.Send(ctx, &{{ $.ServicePackage }}.{{ .GoName }}Result{}); err != nil {
			return err
		}
	}
	return nil
{{- end -}}