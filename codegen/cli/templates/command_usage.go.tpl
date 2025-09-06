{{ printf "%sUsage displays the usage of the %s command and its subcommands." .VarName .Name | comment }}
func {{ .VarName }}Usage() {
	fmt.Fprintln(os.Stderr, `{{ printDescription .Description }}`)
	fmt.Fprintf(os.Stderr, "Usage:\n    %s [globalflags] {{ .Name }} COMMAND [flags]\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "COMMAND:")
{{- range .Subcommands }}
	fmt.Fprintln(os.Stderr, `    {{ .Name }}: {{ printDescription .Description }}`)
{{- end }}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Additional help:")
	fmt.Fprintf(os.Stderr, "    %s {{ .Name }} COMMAND --help\n", os.Args[0])
}

{{- range .Subcommands }}
func {{ .FullName }}Usage() {
	// Header with flags
	fmt.Fprintf(os.Stderr, "%s [flags] {{ $.Name }} {{ .Name }}", os.Args[0])
{{- range .Flags }}
	fmt.Fprint(os.Stderr, " -{{ .Name }} {{ .Type }}")
{{- end }}
	fmt.Fprintln(os.Stderr)

	// Description
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, `{{ printDescription .Description}}`)

	// Flags list
{{- range .Flags }}
	fmt.Fprintln(os.Stderr, `    -{{ .Name }} {{ .Type }}: {{ .Description }}`)
{{- end }}

	// Example block: pass example as parameter to avoid format parsing of % characters
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Example:")
	fmt.Fprintf(os.Stderr, "    %s %s\n", os.Args[0], `{{ .Example }}`)
}
{{ end }}
