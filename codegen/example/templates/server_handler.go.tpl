

	{{ comment "Start the servers and send errors (if any) to the error channel." }}
	switch *hostF {
{{- range $h := .Server.Hosts }}
	case {{ printf "%q" $h.Name }}:
	{{- range $u := $h.URIs }}
		{{- if $.Server.HasTransport $u.Transport.Type }}
		{
			addr := {{ printf "%q" $u.URL }}
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
						log.Fatalf(ctx, "invalid value for URL '{{ .Name }}' variable: %q (valid values: {{ join .Values "," }})\n", *{{ .VarName }}F)
					}
				{{- end }}
				addr = strings.Replace(addr, "{{ printf "{%s}" .Name }}", *{{ .VarName }}F, -1)
			{{- end }}
			u, err := url.Parse(addr)
			if err != nil {
				log.Fatalf(ctx, "invalid URL %#v: %s\n", addr, err)
			}
			if *secureF {
				u.Scheme = "{{ $u.Transport.Type }}s"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *{{ $u.Transport.Type }}PortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					log.Fatalf(ctx, "invalid URL %#v: %s\n", u.Host, err)
				}
				u.Host = net.JoinHostPort(h, *{{ $u.Transport.Type }}PortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "{{ $u.Port }}")
			}
			handle{{ toUpper $u.Transport.Name }}Server(ctx, u, {{ range $t := $.Server.Transports }}{{ if eq $t.Type $u.Transport.Type }}{{ range $s := $t.Services }}{{ range $.Services }}{{ if eq $s .Name }}{{ if .Methods }}{{ .VarName }}Endpoints, {{ end }}{{ end }}{{ end }}{{ end }}{{ end }}{{ end }}&wg, errc, *dbgF)
		}
	{{- end }}
	{{ end }}
{{- end }}
	default:
		log.Fatalf(ctx, "invalid host argument: %q (valid hosts: {{ join .Server.AvailableHosts "|" }})\n", *hostF)
	}	