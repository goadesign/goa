
func main() {
	var (
		hostF = flag.String("host", {{ printf "%q" .Server.DefaultHost.Name }}, "Server host (valid values: {{ (join .Server.AvailableHosts ", ") }})")
		addrF = flag.String("url", "", "URL to service host")
	{{- range .Server.Variables }}
		{{ .VarName }}F = flag.String({{ printf "%q" .Name }}, {{ printf "%q" .DefaultValue }}, {{ printf "%q" .Description }})
	{{- end }}
	{{- if and .HasJSONRPC .HasHTTP }}
		jsonrpcF = flag.Bool("jsonrpc", false, "Force JSON-RPC transport")
		jF = flag.Bool("j", false, "Force JSON-RPC transport")
	{{- end }}

		verboseF = flag.Bool("verbose", false, "Print request and response details")
		vF = flag.Bool("v", false, "Print request and response details")
		timeoutF = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
