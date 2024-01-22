
var (
		addr string
		timeout int
		debug bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
		{{- range $h := .Server.Hosts }}
			case {{ printf "%q" $h.Name }}:
				addr = {{ printf "%q" ($h.DefaultURL $.Server.DefaultTransport.Type) }}
			{{- range $h.Variables }}
				{{- if .Values }}
					var {{ .VarName }}Seen bool
					{
						for _, v := range []string{ {{ range $v := .Values }}"{{ $v }}",{{ end }} } {
							if v == *{{ .VarName }}F {
								{{ .VarName }}Seen = true
								break
							}
						}
					}
					if !{{ .VarName }}Seen {
						fmt.Fprintf(os.Stderr, "invalid value for URL '{{ .Name }}' variable: %q (valid values: {{ join .Values "," }})\n", *{{ .VarName }}F)
						os.Exit(1)
					}
				{{- end }}
				addr = strings.Replace(addr, "{{ printf "{%s}" .Name }}", *{{ .VarName }}F, -1)
			{{- end }}
		{{- end }}
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: {{ join .Server.AvailableHosts "|" }})\n", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s\n", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}	