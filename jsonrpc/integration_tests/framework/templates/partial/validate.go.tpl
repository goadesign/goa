{{- /* Template for validation method implementation */ -}}
	// This method validates the payload
	{{- if eq .Info.Type "string" }}
	if len(p) == 0 {
		return "", fmt.Errorf("validation failed: string cannot be empty")
	}
	return p + " validated", nil
	{{- else }}
	// Validation for non-string types
	return nil, fmt.Errorf("validation not implemented for type {{ .Info.Type }}")
	{{- end }}