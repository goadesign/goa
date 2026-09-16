var (
{{- range .ViewedTypes }}
	{{ printf "%s is a map indexing the attribute names of %s by view name." .Declaration.Name .TypeName | comment }}
	{{ .Declaration.Name }} = map[string][]string{
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
