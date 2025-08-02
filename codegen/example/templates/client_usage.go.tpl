

func usage() {
    var usageCommands []string
{{- range .Server.Transports }}
    {{- if and (eq .Type "http") $.HasHTTP }}
    usageCommands = append(usageCommands, {{ .Type }}UsageCommands()...)
    {{- end }}
{{- end }}
{{- if .HasJSONRPC }}
    usageCommands = append(usageCommands, jsonrpcUsageCommands()...)
{{- end }}
    sort.Strings(usageCommands)
    usageCommands = slices.Compact(usageCommands)
    fmt.Fprintf(os.Stderr, `%s is a command line client for the {{ .APIName }} API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v]{{ range .Server.Variables }}[-{{ .Name }} {{ toUpper .Name }}]{{ end }} SERVICE ENDPOINT [flags]

    -host HOST:  server host ({{ .Server.DefaultHost.Name }}). valid values: {{ (join .Server.AvailableHosts ", ") }}
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
{{- if and .HasJSONRPC .HasHTTP }}
    -jsonrpc|-j: force JSON-RPC (false)
{{- end }}
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
`, os.Args[0], os.Args[0], indent(strings.Join(usageCommands, "\n")), os.Args[0], indent({{ if and (eq .Server.DefaultTransport.Type "http") (not .HasHTTP) .HasJSONRPC }}jsonrpc{{ else }}{{ .Server.DefaultTransport.Type }}{{ end }}UsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.ReplaceAll(s, "\n", "\n    ")
}
