var (
{{- range .ViewedTypes }}
	{{ printf "%sMap is a map indexing the attribute names of %s by view name." .Name .Name | comment }}
	{{ .Name }}Map = map[string][]string{
	{{- range .Views }}
		"{{ .Name }}": {
			{{- range $n := .Attributes }}
				"{{ $n }}",
			{{- end }}
		},
	{{- end }}
	}
{{- end }}
)
