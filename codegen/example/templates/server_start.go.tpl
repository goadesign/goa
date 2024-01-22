
func main() {
	{{ comment "Define command line flags, add any other flag required to configure the service." }}
	var(
		hostF = flag.String("host", {{ printf "%q" .Server.DefaultHost.Name }}, "Server host (valid values: {{ (join .Server.AvailableHosts ", ") }})")
		domainF = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
	{{- range .Server.Transports }}
	{{ .Type }}PortF = flag.String("{{ .Type }}-port", "", "{{ .Name }} port (overrides host {{ .Name }} port specified in service design)")
	{{- end }}
	{{- range .Server.Variables }}
	{{ .VarName }}F = flag.String({{ printf "%q" .Name }}, {{ printf "%q" .DefaultValue }}, "{{ .Description }}{{ if .Values }} (valid values: {{ join .Values ", " }}){{ end }}")
	{{- end }}
		secureF = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF  = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()