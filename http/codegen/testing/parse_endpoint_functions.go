package testing

var MultiNoPayloadCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(scheme, host string, doer goahttp.Doer, enc func(*http.Request) goahttp.Encoder, dec func(*http.Response) goahttp.Decoder) (goa.Endpoint, interface{}, error) {
	var (
		serviceMultiNoPayload1Flags = flag.NewFlagSet("serviceMultiNoPayload1", flag.ContinueOnError)

		serviceMultiNoPayload1MethodServiceNoPayload11Flags = flag.NewFlagSet("methodServiceNoPayload11", flag.ExitOnError)

		serviceMultiNoPayload1MethodServiceNoPayload12Flags = flag.NewFlagSet("methodServiceNoPayload12", flag.ExitOnError)

		serviceMultiNoPayload2Flags = flag.NewFlagSet("serviceMultiNoPayload2", flag.ContinueOnError)

		serviceMultiNoPayload2MethodServiceNoPayload21Flags = flag.NewFlagSet("methodServiceNoPayload21", flag.ExitOnError)

		serviceMultiNoPayload2MethodServiceNoPayload22Flags = flag.NewFlagSet("methodServiceNoPayload22", flag.ExitOnError)
	)
	serviceMultiNoPayload1Flags.Usage = serviceMultiNoPayload1Usage
	methodServiceNoPayload11Flags.Usage = serviceMultiNoPayload1MethodServiceNoPayload11Usage
	methodServiceNoPayload12Flags.Usage = serviceMultiNoPayload1MethodServiceNoPayload12Usage

	serviceMultiNoPayload2Flags.Usage = serviceMultiNoPayload2Usage
	methodServiceNoPayload21Flags.Usage = serviceMultiNoPayload2MethodServiceNoPayload21Usage
	methodServiceNoPayload22Flags.Usage = serviceMultiNoPayload2MethodServiceNoPayload22Usage

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, nil, err
	}

	if len(os.Args) < flag.NFlag()+3 {
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		svcn string
		svcf *flag.FlagSet
	)
	{
		svcn = os.Args[1+flag.NFlag()]
		switch svcn {
		case "serviceMultiNoPayload1":
			svcf = serviceMultiNoPayload1Flags
		case "serviceMultiNoPayload2":
			svcf = serviceMultiNoPayload2Flags
		default:
			return nil, nil, fmt.Errorf("unknown service %q", svcn)
		}
	}
	if err := svcf.Parse(os.Args[2+flag.NFlag():]); err != nil {
		return nil, nil, err
	}

	var (
		epn string
		epf *flag.FlagSet
	)
	{
		epn = os.Args[2+flag.NFlag()+svcf.NFlag()]
		switch svcn {
		case "serviceMultiNoPayload1":
			switch epn {
			case "methodServiceNoPayload11":
				epf = serviceMultiNoPayload1MethodServiceNoPayload11Flags

			case "methodServiceNoPayload12":
				epf = serviceMultiNoPayload1MethodServiceNoPayload12Flags

			}

		case "serviceMultiNoPayload2":
			switch epn {
			case "methodServiceNoPayload21":
				epf = serviceMultiNoPayload2MethodServiceNoPayload21Flags

			case "methodServiceNoPayload22":
				epf = serviceMultiNoPayload2MethodServiceNoPayload22Flags

			}

		}
	}
	if epf == nil {
		return nil, nil, fmt.Errorf("unknown %q endpoint %q", svcn, epn)
	}

	// Parse endpoint flags if any
	if len(os.Args) > 2+flag.NFlag()+svcf.NFlag() {
		if err := epf.Parse(os.Args[3+flag.NFlag()+svcf.NFlag():]); err != nil {
			return nil, nil, err
		}
	}

	var (
		data     interface{}
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
		case "serviceMultiNoPayload1":
			c := storagec.NewClient(scheme, host, doer, enc, dec)
			switch epn {
			case "methodServiceNoPayload11":
				endpoint = c.MethodServiceNoPayload11()
				data = nil
			case "methodServiceNoPayload12":
				endpoint = c.MethodServiceNoPayload12()
				data = nil
			}
		case "serviceMultiNoPayload2":
			c := storagec.NewClient(scheme, host, doer, enc, dec)
			switch epn {
			case "methodServiceNoPayload21":
				endpoint = c.MethodServiceNoPayload21()
				data = nil
			case "methodServiceNoPayload22":
				endpoint = c.MethodServiceNoPayload22()
				data = nil
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var MultiSimpleCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(scheme, host string, doer goahttp.Doer, enc func(*http.Request) goahttp.Encoder, dec func(*http.Response) goahttp.Decoder) (goa.Endpoint, interface{}, error) {
	var (
		serviceMultiSimple1Flags = flag.NewFlagSet("serviceMultiSimple1", flag.ContinueOnError)

		serviceMultiSimple1MethodMultiSimpleNoPayloadFlags = flag.NewFlagSet("methodMultiSimpleNoPayload", flag.ExitOnError)

		serviceMultiSimple1MethodMultiSimplePayloadFlags    = flag.NewFlagSet("methodMultiSimplePayload", flag.ExitOnError)
		serviceMultiSimple1MethodMultiSimplePayloadBodyFlag = serviceMultiSimple1MethodMultiSimplePayloadFlags.String("body", "", "")

		serviceMultiSimple2Flags = flag.NewFlagSet("serviceMultiSimple2", flag.ContinueOnError)

		serviceMultiSimple2MethodMultiSimpleNoPayloadFlags = flag.NewFlagSet("methodMultiSimpleNoPayload", flag.ExitOnError)

		serviceMultiSimple2MethodMultiSimplePayloadFlags    = flag.NewFlagSet("methodMultiSimplePayload", flag.ExitOnError)
		serviceMultiSimple2MethodMultiSimplePayloadBodyFlag = serviceMultiSimple2MethodMultiSimplePayloadFlags.String("body", "", "")
	)
	serviceMultiSimple1Flags.Usage = serviceMultiSimple1Usage
	methodMultiSimpleNoPayloadFlags.Usage = serviceMultiSimple1MethodMultiSimpleNoPayloadUsage
	methodMultiSimplePayloadFlags.Usage = serviceMultiSimple1MethodMultiSimplePayloadUsage

	serviceMultiSimple2Flags.Usage = serviceMultiSimple2Usage
	methodMultiSimpleNoPayloadFlags.Usage = serviceMultiSimple2MethodMultiSimpleNoPayloadUsage
	methodMultiSimplePayloadFlags.Usage = serviceMultiSimple2MethodMultiSimplePayloadUsage

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, nil, err
	}

	if len(os.Args) < flag.NFlag()+3 {
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		svcn string
		svcf *flag.FlagSet
	)
	{
		svcn = os.Args[1+flag.NFlag()]
		switch svcn {
		case "serviceMultiSimple1":
			svcf = serviceMultiSimple1Flags
		case "serviceMultiSimple2":
			svcf = serviceMultiSimple2Flags
		default:
			return nil, nil, fmt.Errorf("unknown service %q", svcn)
		}
	}
	if err := svcf.Parse(os.Args[2+flag.NFlag():]); err != nil {
		return nil, nil, err
	}

	var (
		epn string
		epf *flag.FlagSet
	)
	{
		epn = os.Args[2+flag.NFlag()+svcf.NFlag()]
		switch svcn {
		case "serviceMultiSimple1":
			switch epn {
			case "methodMultiSimpleNoPayload":
				epf = serviceMultiSimple1MethodMultiSimpleNoPayloadFlags

			case "methodMultiSimplePayload":
				epf = serviceMultiSimple1MethodMultiSimplePayloadFlags

			}

		case "serviceMultiSimple2":
			switch epn {
			case "methodMultiSimpleNoPayload":
				epf = serviceMultiSimple2MethodMultiSimpleNoPayloadFlags

			case "methodMultiSimplePayload":
				epf = serviceMultiSimple2MethodMultiSimplePayloadFlags

			}

		}
	}
	if epf == nil {
		return nil, nil, fmt.Errorf("unknown %q endpoint %q", svcn, epn)
	}

	// Parse endpoint flags if any
	if len(os.Args) > 2+flag.NFlag()+svcf.NFlag() {
		if err := epf.Parse(os.Args[3+flag.NFlag()+svcf.NFlag():]); err != nil {
			return nil, nil, err
		}
	}

	var (
		data     interface{}
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
		case "serviceMultiSimple1":
			c := storagec.NewClient(scheme, host, doer, enc, dec)
			switch epn {
			case "methodMultiSimpleNoPayload":
				endpoint = c.MethodMultiSimpleNoPayload()
				data = nil
			case "methodMultiSimplePayload":
				endpoint = c.MethodMultiSimplePayload()
				data, err = buildMethodMultiSimplePayloadPayload(*serviceMultiSimple1MethodMultiSimplePayloadBodyFlag)
			}
		case "serviceMultiSimple2":
			c := storagec.NewClient(scheme, host, doer, enc, dec)
			switch epn {
			case "methodMultiSimpleNoPayload":
				endpoint = c.MethodMultiSimpleNoPayload()
				data = nil
			case "methodMultiSimplePayload":
				endpoint = c.MethodMultiSimplePayload()
				data, err = buildMethodMultiSimplePayloadPayload(*serviceMultiSimple2MethodMultiSimplePayloadBodyFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var MultiCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(scheme, host string, doer goahttp.Doer, enc func(*http.Request) goahttp.Encoder, dec func(*http.Response) goahttp.Decoder) (goa.Endpoint, interface{}, error) {
	var (
		serviceMultiFlags = flag.NewFlagSet("serviceMulti", flag.ContinueOnError)

		serviceMultiMethodMultiNoPayloadFlags = flag.NewFlagSet("methodMultiNoPayload", flag.ExitOnError)

		serviceMultiMethodMultiPayloadFlags    = flag.NewFlagSet("methodMultiPayload", flag.ExitOnError)
		serviceMultiMethodMultiPayloadBodyFlag = serviceMultiMethodMultiPayloadFlags.String("body", "", "")
		serviceMultiMethodMultiPayloadBFlag    = serviceMultiMethodMultiPayloadFlags.String("b", "", "")
		serviceMultiMethodMultiPayloadAFlag    = serviceMultiMethodMultiPayloadFlags.String("a", "", "")
	)
	serviceMultiFlags.Usage = serviceMultiUsage
	methodMultiNoPayloadFlags.Usage = serviceMultiMethodMultiNoPayloadUsage
	methodMultiPayloadFlags.Usage = serviceMultiMethodMultiPayloadUsage

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, nil, err
	}

	if len(os.Args) < flag.NFlag()+3 {
		return nil, nil, fmt.Errorf("not enough arguments")
	}

	var (
		svcn string
		svcf *flag.FlagSet
	)
	{
		svcn = os.Args[1+flag.NFlag()]
		switch svcn {
		case "serviceMulti":
			svcf = serviceMultiFlags
		default:
			return nil, nil, fmt.Errorf("unknown service %q", svcn)
		}
	}
	if err := svcf.Parse(os.Args[2+flag.NFlag():]); err != nil {
		return nil, nil, err
	}

	var (
		epn string
		epf *flag.FlagSet
	)
	{
		epn = os.Args[2+flag.NFlag()+svcf.NFlag()]
		switch svcn {
		case "serviceMulti":
			switch epn {
			case "methodMultiNoPayload":
				epf = serviceMultiMethodMultiNoPayloadFlags

			case "methodMultiPayload":
				epf = serviceMultiMethodMultiPayloadFlags

			}

		}
	}
	if epf == nil {
		return nil, nil, fmt.Errorf("unknown %q endpoint %q", svcn, epn)
	}

	// Parse endpoint flags if any
	if len(os.Args) > 2+flag.NFlag()+svcf.NFlag() {
		if err := epf.Parse(os.Args[3+flag.NFlag()+svcf.NFlag():]); err != nil {
			return nil, nil, err
		}
	}

	var (
		data     interface{}
		endpoint goa.Endpoint
		err      error
	)
	{
		switch svcn {
		case "serviceMulti":
			c := storagec.NewClient(scheme, host, doer, enc, dec)
			switch epn {
			case "methodMultiNoPayload":
				endpoint = c.MethodMultiNoPayload()
				data = nil
			case "methodMultiPayload":
				endpoint = c.MethodMultiPayload()
				data, err = buildMethodMultiPayloadPayload(*serviceMultiMethodMultiPayloadBodyFlag, *serviceMultiMethodMultiPayloadBFlag, *serviceMultiMethodMultiPayloadAFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`
