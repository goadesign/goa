{{ printf "%sUsage displays the usage of the %s command and its subcommands." .VarName .Name | comment }}
func {{ .VarName }}Usage() {
	fmt.Fprint(os.Stderr, `{{ printDescription .Description }}
Usage:
    `+os.Args[0]+` [globalflags] {{ .Name }} COMMAND [flags]

COMMAND:
    {{- range .Subcommands }}
    {{ .Name }}: {{ printDescription .Description }}
    {{- end }}

Additional help:
    `+os.Args[0]+` {{ .Name }} COMMAND --help
`)
}

{{- range .Subcommands }}
func {{ .FullName }}Usage() {
	fmt.Fprint(os.Stderr, os.Args[0]+` [flags] {{ $.Name }} {{ .Name }}{{range .Flags }} -{{ .Name }} {{ .Type }}{{ end }}

{{ printDescription .Description}}
	{{- range .Flags }}
    -{{ .Name }} {{ .Type }}: {{ .Description }}
	{{- end }}

Example:
    `+os.Args[0]+` {{ .Example }}
`)
}
{{ end }}
