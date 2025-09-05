{{ printf "%sUsage displays the usage of the %s command and its subcommands." .VarName .Name | comment }}
func {{ .VarName }}Usage() {
	fmt.Fprintf(os.Stderr, `{{ printDescription .Description }}
Usage:
    %[1]s [globalflags] {{ .Name }} COMMAND [flags]

COMMAND:
    {{- range .Subcommands }}
    {{ .Name }}: {{ printDescription .Description }}
    {{- end }}

Additional help:
    %[1]s {{ .Name }} COMMAND --help
`, os.Args[0])
}

{{- range .Subcommands }}
func {{ .FullName }}Usage() {
	fmt.Fprintf(os.Stderr, "%s [flags] {{ $.Name }} {{ .Name }}{{range .Flags }} -{{ .Name }} {{ .Type }}{{ end }}\n\n", os.Args[0])
	fmt.Fprint(os.Stderr, `{{ printDescription .Description}}
	{{- range .Flags }}
    -{{ .Name }} {{ .Type }}: {{ .Description }}
	{{- end }}

Example:
    `)
	fmt.Fprintf(os.Stderr, "%s %s\n", os.Args[0], `{{ .Example }}`)
}
{{ end }}
