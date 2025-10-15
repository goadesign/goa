
	var (
		endpoint goa.Endpoint
		payload any
		err error
	)
	{
		switch scheme {
	{{- range $t := .Server.Transports }}
		case "{{ $t.Type }}", "{{ $t.Type }}s":
		{{- if and (eq $t.Type  "http") $.HasJSONRPC }}
			{{- if $.HasHTTP }}
			if *jsonrpcF || *jF {
				endpoint, payload, err = doJSONRPC(scheme, host, timeout, debug)
			} else {
				endpoint, payload, err = doHTTP(scheme, host, timeout, debug)
				if err != nil && strings.HasPrefix(err.Error(), "unknown") {
					endpoint, payload, err = doJSONRPC(scheme, host, timeout, debug)
				}
			}
			{{- else }}
			endpoint, payload, err = doJSONRPC(scheme, host, timeout, debug)
			{{- end }}
		{{- else }}
			endpoint, payload, err = do{{ toUpper $t.Name }}(scheme, host, timeout, debug)
		{{- end }}
	{{- end }}
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: {{ join .Server.Schemes "|" }})\n", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}
