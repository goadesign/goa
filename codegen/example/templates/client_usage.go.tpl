

func usage() {
  fmt.Fprintf(os.Stderr, `%s is a command line client for the {{ .APIName }} API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v]{{ range .Server.Variables }}[-{{ .Name }} {{ toUpper .Name }}]{{ end }} SERVICE ENDPOINT [flags]

    -host HOST:  server host ({{ .Server.DefaultHost.Name }}). valid values: {{ (join .Server.AvailableHosts ", ") }}
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)
	{{- range .Server.Variables }}
    -{{ .Name }}:    {{ .Description }} ({{ .DefaultValue }})
	{{- end }}

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
`, os.Args[0], os.Args[0], indent({{ .Server.DefaultTransport.Type }}UsageCommands()), os.Args[0], indent({{ .Server.DefaultTransport.Type }}UsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}