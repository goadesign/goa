// UsageCommands returns the set of commands and sub-commands using the format
//
//    command (subcommand1|subcommand2|...)
//
func UsageCommands() []string {
        return []string{
{{- range . }}
                "{{ . }}",
{{- end }}
        }
}
