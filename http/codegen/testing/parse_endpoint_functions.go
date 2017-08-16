package testing

var MultiNoPayloadParseCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		serviceMultiNoPayload1Flags = flag.NewFlagSet("serviceMultiNoPayload1", flag.ContinueOnError)

		serviceMultiNoPayload1MethodServiceNoPayload11Flags = flag.NewFlagSet("methodServiceNoPayload11", flag.ExitOnError)

		serviceMultiNoPayload1MethodServiceNoPayload12Flags = flag.NewFlagSet("methodServiceNoPayload12", flag.ExitOnError)

		serviceMultiNoPayload2Flags = flag.NewFlagSet("serviceMultiNoPayload2", flag.ContinueOnError)

		serviceMultiNoPayload2MethodServiceNoPayload21Flags = flag.NewFlagSet("methodServiceNoPayload21", flag.ExitOnError)

		serviceMultiNoPayload2MethodServiceNoPayload22Flags = flag.NewFlagSet("methodServiceNoPayload22", flag.ExitOnError)
	)
	serviceMultiNoPayload1Flags.Usage = serviceMultiNoPayload1Usage
	serviceMultiNoPayload1MethodServiceNoPayload11Flags.Usage = serviceMultiNoPayload1MethodServiceNoPayload11Usage
	serviceMultiNoPayload1MethodServiceNoPayload12Flags.Usage = serviceMultiNoPayload1MethodServiceNoPayload12Usage

	serviceMultiNoPayload2Flags.Usage = serviceMultiNoPayload2Usage
	serviceMultiNoPayload2MethodServiceNoPayload21Flags.Usage = serviceMultiNoPayload2MethodServiceNoPayload21Usage
	serviceMultiNoPayload2MethodServiceNoPayload22Flags.Usage = serviceMultiNoPayload2MethodServiceNoPayload22Usage

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
			c := servicemultinopayload1c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "methodServiceNoPayload11":
				endpoint = c.MethodServiceNoPayload11()
				data = nil
			case "methodServiceNoPayload12":
				endpoint = c.MethodServiceNoPayload12()
				data = nil
			}
		case "serviceMultiNoPayload2":
			c := servicemultinopayload2c.NewClient(scheme, host, doer, enc, dec, restore)
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

var MultiSimpleParseCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		serviceMultiSimple1Flags = flag.NewFlagSet("serviceMultiSimple1", flag.ContinueOnError)

		serviceMultiSimple1MethodMultiSimpleNoPayloadFlags = flag.NewFlagSet("methodMultiSimpleNoPayload", flag.ExitOnError)

		serviceMultiSimple1MethodMultiSimplePayloadFlags    = flag.NewFlagSet("methodMultiSimplePayload", flag.ExitOnError)
		serviceMultiSimple1MethodMultiSimplePayloadBodyFlag = serviceMultiSimple1MethodMultiSimplePayloadFlags.String("body", "REQUIRED", "")

		serviceMultiSimple2Flags = flag.NewFlagSet("serviceMultiSimple2", flag.ContinueOnError)

		serviceMultiSimple2MethodMultiSimpleNoPayloadFlags = flag.NewFlagSet("methodMultiSimpleNoPayload", flag.ExitOnError)

		serviceMultiSimple2MethodMultiSimplePayloadFlags    = flag.NewFlagSet("methodMultiSimplePayload", flag.ExitOnError)
		serviceMultiSimple2MethodMultiSimplePayloadBodyFlag = serviceMultiSimple2MethodMultiSimplePayloadFlags.String("body", "REQUIRED", "")
	)
	serviceMultiSimple1Flags.Usage = serviceMultiSimple1Usage
	serviceMultiSimple1MethodMultiSimpleNoPayloadFlags.Usage = serviceMultiSimple1MethodMultiSimpleNoPayloadUsage
	serviceMultiSimple1MethodMultiSimplePayloadFlags.Usage = serviceMultiSimple1MethodMultiSimplePayloadUsage

	serviceMultiSimple2Flags.Usage = serviceMultiSimple2Usage
	serviceMultiSimple2MethodMultiSimpleNoPayloadFlags.Usage = serviceMultiSimple2MethodMultiSimpleNoPayloadUsage
	serviceMultiSimple2MethodMultiSimplePayloadFlags.Usage = serviceMultiSimple2MethodMultiSimplePayloadUsage

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
			c := servicemultisimple1c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "methodMultiSimpleNoPayload":
				endpoint = c.MethodMultiSimpleNoPayload()
				data = nil
			case "methodMultiSimplePayload":
				endpoint = c.MethodMultiSimplePayload()
				data, err = servicemultisimple1c.BuildMethodMultiSimplePayloadPayload(*serviceMultiSimple1MethodMultiSimplePayloadBodyFlag)
			}
		case "serviceMultiSimple2":
			c := servicemultisimple2c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "methodMultiSimpleNoPayload":
				endpoint = c.MethodMultiSimpleNoPayload()
				data = nil
			case "methodMultiSimplePayload":
				endpoint = c.MethodMultiSimplePayload()
				data, err = servicemultisimple2c.BuildMethodMultiSimplePayloadPayload(*serviceMultiSimple2MethodMultiSimplePayloadBodyFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var MultiParseCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		serviceMultiFlags = flag.NewFlagSet("serviceMulti", flag.ContinueOnError)

		serviceMultiMethodMultiNoPayloadFlags = flag.NewFlagSet("methodMultiNoPayload", flag.ExitOnError)

		serviceMultiMethodMultiPayloadFlags    = flag.NewFlagSet("methodMultiPayload", flag.ExitOnError)
		serviceMultiMethodMultiPayloadBodyFlag = serviceMultiMethodMultiPayloadFlags.String("body", "REQUIRED", "")
		serviceMultiMethodMultiPayloadBFlag    = serviceMultiMethodMultiPayloadFlags.String("b", "", "")
		serviceMultiMethodMultiPayloadAFlag    = serviceMultiMethodMultiPayloadFlags.String("a", "", "")
	)
	serviceMultiFlags.Usage = serviceMultiUsage
	serviceMultiMethodMultiNoPayloadFlags.Usage = serviceMultiMethodMultiNoPayloadUsage
	serviceMultiMethodMultiPayloadFlags.Usage = serviceMultiMethodMultiPayloadUsage

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
			c := servicemultic.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "methodMultiNoPayload":
				endpoint = c.MethodMultiNoPayload()
				data = nil
			case "methodMultiPayload":
				endpoint = c.MethodMultiPayload()
				data, err = servicemultic.BuildMethodMultiPayloadPayload(*serviceMultiMethodMultiPayloadBodyFlag, *serviceMultiMethodMultiPayloadBFlag, *serviceMultiMethodMultiPayloadAFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var MultiSimpleBuildCode = `// BuildMethodMultiSimplePayloadPayload builds the payload for the
// ServiceMultiSimple1 MethodMultiSimplePayload endpoint from CLI flags.
func BuildMethodMultiSimplePayloadPayload(serviceMultiSimple1MethodMultiSimplePayloadBody string) (*servicemultisimple1.MethodMultiSimplePayloadPayload, error) {
	var body MethodMultiSimplePayloadRequestBody
	{
		err := json.Unmarshal([]byte(serviceMultiSimple1MethodMultiSimplePayloadBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'{\n      \"a\": false\n   }'")
		}
	}
	v := &servicemultisimple1.MethodMultiSimplePayloadPayload{
		A: body.A,
	}

	return v, nil
}
`

var MultiBuildCode = `// BuildMethodMultiPayloadPayload builds the payload for the ServiceMulti
// MethodMultiPayload endpoint from CLI flags.
func BuildMethodMultiPayloadPayload(serviceMultiMethodMultiPayloadBody string, serviceMultiMethodMultiPayloadB string, serviceMultiMethodMultiPayloadA string) (*servicemulti.MethodMultiPayloadPayload, error) {
	var body MethodMultiPayloadRequestBody
	{
		err := json.Unmarshal([]byte(serviceMultiMethodMultiPayloadBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'{\n      \"c\": {\n         \"att\": false,\n         \"att10\": \"Aspernatur quo error explicabo pariatur.\",\n         \"att11\": \"Q3VtcXVlIHZvbHVwdGF0ZW0u\",\n         \"att12\": \"Distinctio aliquam nihil blanditiis ut.\",\n         \"att13\": [\n            \"Nihil excepturi deserunt quasi omnis sed.\",\n            \"Sit maiores aperiam autem non ea rem.\"\n         ],\n         \"att14\": {\n            \"Excepturi totam.\": \"Ut aut facilis vel ipsam.\",\n            \"Minima et aut non sunt consequuntur.\": \"Et consequuntur porro quasi.\",\n            \"Quis voluptates quaerat et temporibus facere.\": \"Ipsam eaque sunt maxime suscipit.\"\n         },\n         \"att15\": {\n            \"inline\": \"Ea alias repellat nobis veritatis.\"\n         },\n         \"att2\": 3504438334001971349,\n         \"att3\": 2005839040,\n         \"att4\": 5845720715558772393,\n         \"att5\": 2900634008447043830,\n         \"att6\": 1865618013,\n         \"att7\": 1484745265794365762,\n         \"att8\": 0.11815318,\n         \"att9\": 0.30907290919538355\n      }\n   }'")
		}
	}
	var b *string
	{
		if serviceMultiMethodMultiPayloadB != "" {
			b = &serviceMultiMethodMultiPayloadB
		}
	}
	var a *bool
	{
		if serviceMultiMethodMultiPayloadA != "" {
			val, err := strconv.ParseBool(serviceMultiMethodMultiPayloadA)
			if err != nil {
				return nil, fmt.Errorf("invalid value for a, must be BOOL")
			}
			a = &val
		}
	}
	v := &servicemulti.MethodMultiPayloadPayload{}
	if body.C != nil {
		v.C = userTypeRequestBodyToUserType(body.C)
	}
	v.B = b
	v.A = a

	return v, nil
}
`
