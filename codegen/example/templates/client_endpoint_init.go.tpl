
	var (
		err error
	)
	{
		switch scheme {
	{{- range $t := .Server.Transports }}
		case "{{ $t.Type }}", "{{ $t.Type }}s":
		{{- if and (eq $t.Type  "http") $.HasJSONRPC }}
			{{- if $.HasHTTP }}
			if *jsonrpcF || *jF {
				err = doJSONRPC(context.Background(), scheme, host, timeout, debug, os.Stdout)
			} else {
				switch flag.Arg(0) {
				{{- range $.JSONRPCOnly }}
				case {{ printf "%q" .Service }}:
					switch flag.Arg(1) {
					{{- range .Endpoints }}
					case {{ printf "%q" . }}:
						err = doJSONRPC(context.Background(), scheme, host, timeout, debug, os.Stdout)
					{{- end }}
					default:
						err = doHTTP(context.Background(), scheme, host, timeout, debug, os.Stdout)
					}
				{{- end }}
				default:
					err = doHTTP(context.Background(), scheme, host, timeout, debug, os.Stdout)
				}
			}
			{{- else }}
			err = doJSONRPC(context.Background(), scheme, host, timeout, debug, os.Stdout)
			{{- end }}
		{{- else }}
			err = do{{ toUpper $t.Name }}(context.Background(), scheme, host, timeout, debug, os.Stdout)
		{{- end }}
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
