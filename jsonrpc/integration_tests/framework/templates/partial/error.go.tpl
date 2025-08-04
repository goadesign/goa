{{- /* Template for error method implementation */ -}}
	return {{ if .HasResult }}nil, {{ end }}{{ .ServicePackage }}.MakeTestError(fmt.Errorf("test error"))