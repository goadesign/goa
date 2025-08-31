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
	fmt.Fprintf(os.Stderr, `%[1]s [flags] {{ $.Name }} {{ .Name }}{{range .Flags }} -{{ .Name }} {{ .Type }}{{ end }}

{{ printDescription .Description}}
	{{- range .Flags }}
    -{{ .Name }} {{ .Type }}: {{ .Description }}
	{{- end }}

Example:
    %[1]s {{ .Example }}
`, os.Args[0])
}
{{ end }}
