
	var(
		endpoint goa.Endpoint
		payload any
		err error
	)
	{
		switch scheme {
	{{- range $t := .Server.Transports }}
		case "{{ $t.Type }}", "{{ $t.Type }}s":
			endpoint, payload, err = do{{ toUpper $t.Name }}(scheme, host, timeout, debug)
	{{- end }}
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: {{ join .Server.Schemes "|" }})\n", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}
