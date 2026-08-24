var (
		{{- range .Commands }}
		{{ .FlagSetVar }} = flag.NewFlagSet("{{ .Name }}", flag.ContinueOnError)
		{{ range .Subcommands }}
		{{ .FlagSetVar }} = flag.NewFlagSet("{{ .Name }}", flag.ExitOnError)
		{{- $sub := . }}
		{{- range .Flags }}
		{{ .PointerVar }} = {{ $sub.FlagSetVar }}.String("{{ .Name }}", "{{ if .Default }}{{ .Default }}{{ else if .Required }}REQUIRED{{ end }}", {{ printf "%q" .Description }})
		{{- end }}
		{{ end }}
		{{- end }}
	)
	{{ range .Commands -}}
	{{ $cmd := . -}}
	{{ .FlagSetVar }}.Usage = {{ .UsageDeclaration.Name }}
	{{ range .Subcommands -}}
	{{ .FlagSetVar }}.Usage = {{ .UsageDeclaration.Name }}
	{{ end }}
	{{ end }}
	if {{ .Variables.Error }} := flag.CommandLine.Parse(os.Args[1:]); {{ .Variables.Error }} != nil {
		return nil, nil, {{ .Variables.Error }}
	}

	if flag.NArg() < 2 { // two non flag args are required: SERVICE and ENDPOINT (aka COMMAND)
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		{{ .Variables.ServiceName }} string
		{{ .Variables.ServiceFlags }} *flag.FlagSet
	)
	{
		{{ .Variables.ServiceName }} = flag.Arg(0)
		switch {{ .Variables.ServiceName }} {
	{{- range .Commands }}
		case "{{ .Name }}":
			{{ $.Variables.ServiceFlags }} = {{ .FlagSetVar }}
	{{- end }}
		default:
			return nil, nil, fmt.Errorf("unknown service %q", {{ .Variables.ServiceName }})
		}
	}
	if {{ .Variables.Error }} := {{ .Variables.ServiceFlags }}.Parse(flag.Args()[1:]); {{ .Variables.Error }} != nil {
		return nil, nil, {{ .Variables.Error }}
	}

	var (
		{{ .Variables.MethodName }} string
		{{ .Variables.MethodFlags }} *flag.FlagSet
	)
	{
		{{ .Variables.MethodName }} = {{ .Variables.ServiceFlags }}.Arg(0)
		switch {{ .Variables.ServiceName }} {
	{{- range .Commands }}
		case "{{ .Name }}":
			switch {{ $.Variables.MethodName }} {
		{{- range .Subcommands }}
			case "{{ .Name }}":
				{{ $.Variables.MethodFlags }} = {{ .FlagSetVar }}
		{{ end }}
			}
	{{ end }}
		}
	}
	if {{ .Variables.MethodFlags }} == nil {
		return nil, nil, fmt.Errorf("unknown %q endpoint %q", {{ .Variables.ServiceName }}, {{ .Variables.MethodName }})
	}

	// Parse endpoint flags if any
	if {{ .Variables.ServiceFlags }}.NArg() > 1 {
		if {{ .Variables.Error }} := {{ .Variables.MethodFlags }}.Parse({{ .Variables.ServiceFlags }}.Args()[1:]); {{ .Variables.Error }} != nil {
			return nil, nil, {{ .Variables.Error }}
		}
	}
