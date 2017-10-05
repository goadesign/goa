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
		serviceMultiNoPayload1Flags = flag.NewFlagSet("ServiceMultiNoPayload1", flag.ContinueOnError)

		serviceMultiNoPayload1MethodServiceNoPayload11Flags = flag.NewFlagSet("MethodServiceNoPayload11", flag.ExitOnError)

		serviceMultiNoPayload1MethodServiceNoPayload12Flags = flag.NewFlagSet("MethodServiceNoPayload12", flag.ExitOnError)

		serviceMultiNoPayload2Flags = flag.NewFlagSet("ServiceMultiNoPayload2", flag.ContinueOnError)

		serviceMultiNoPayload2MethodServiceNoPayload21Flags = flag.NewFlagSet("MethodServiceNoPayload21", flag.ExitOnError)

		serviceMultiNoPayload2MethodServiceNoPayload22Flags = flag.NewFlagSet("MethodServiceNoPayload22", flag.ExitOnError)
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
		case "ServiceMultiNoPayload1":
			svcf = serviceMultiNoPayload1Flags
		case "ServiceMultiNoPayload2":
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
		case "ServiceMultiNoPayload1":
			switch epn {
			case "MethodServiceNoPayload11":
				epf = serviceMultiNoPayload1MethodServiceNoPayload11Flags

			case "MethodServiceNoPayload12":
				epf = serviceMultiNoPayload1MethodServiceNoPayload12Flags

			}

		case "ServiceMultiNoPayload2":
			switch epn {
			case "MethodServiceNoPayload21":
				epf = serviceMultiNoPayload2MethodServiceNoPayload21Flags

			case "MethodServiceNoPayload22":
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
		case "ServiceMultiNoPayload1":
			c := servicemultinopayload1c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodServiceNoPayload11":
				endpoint = c.MethodServiceNoPayload11()
				data = nil
			case "MethodServiceNoPayload12":
				endpoint = c.MethodServiceNoPayload12()
				data = nil
			}
		case "ServiceMultiNoPayload2":
			c := servicemultinopayload2c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodServiceNoPayload21":
				endpoint = c.MethodServiceNoPayload21()
				data = nil
			case "MethodServiceNoPayload22":
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
		serviceMultiSimple1Flags = flag.NewFlagSet("ServiceMultiSimple1", flag.ContinueOnError)

		serviceMultiSimple1MethodMultiSimpleNoPayloadFlags = flag.NewFlagSet("MethodMultiSimpleNoPayload", flag.ExitOnError)

		serviceMultiSimple1MethodMultiSimplePayloadFlags    = flag.NewFlagSet("MethodMultiSimplePayload", flag.ExitOnError)
		serviceMultiSimple1MethodMultiSimplePayloadBodyFlag = serviceMultiSimple1MethodMultiSimplePayloadFlags.String("body", "REQUIRED", "")

		serviceMultiSimple2Flags = flag.NewFlagSet("ServiceMultiSimple2", flag.ContinueOnError)

		serviceMultiSimple2MethodMultiSimpleNoPayloadFlags = flag.NewFlagSet("MethodMultiSimpleNoPayload", flag.ExitOnError)

		serviceMultiSimple2MethodMultiSimplePayloadFlags    = flag.NewFlagSet("MethodMultiSimplePayload", flag.ExitOnError)
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
		case "ServiceMultiSimple1":
			svcf = serviceMultiSimple1Flags
		case "ServiceMultiSimple2":
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
		case "ServiceMultiSimple1":
			switch epn {
			case "MethodMultiSimpleNoPayload":
				epf = serviceMultiSimple1MethodMultiSimpleNoPayloadFlags

			case "MethodMultiSimplePayload":
				epf = serviceMultiSimple1MethodMultiSimplePayloadFlags

			}

		case "ServiceMultiSimple2":
			switch epn {
			case "MethodMultiSimpleNoPayload":
				epf = serviceMultiSimple2MethodMultiSimpleNoPayloadFlags

			case "MethodMultiSimplePayload":
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
		case "ServiceMultiSimple1":
			c := servicemultisimple1c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodMultiSimpleNoPayload":
				endpoint = c.MethodMultiSimpleNoPayload()
				data = nil
			case "MethodMultiSimplePayload":
				endpoint = c.MethodMultiSimplePayload()
				data, err = servicemultisimple1c.BuildMethodMultiSimplePayloadMethodMultiSimplePayloadPayload(*serviceMultiSimple1MethodMultiSimplePayloadBodyFlag)
			}
		case "ServiceMultiSimple2":
			c := servicemultisimple2c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodMultiSimpleNoPayload":
				endpoint = c.MethodMultiSimpleNoPayload()
				data = nil
			case "MethodMultiSimplePayload":
				endpoint = c.MethodMultiSimplePayload()
				data, err = servicemultisimple2c.BuildMethodMultiSimplePayloadMethodMultiSimplePayloadPayload(*serviceMultiSimple2MethodMultiSimplePayloadBodyFlag)
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
		serviceMultiFlags = flag.NewFlagSet("ServiceMulti", flag.ContinueOnError)

		serviceMultiMethodMultiNoPayloadFlags = flag.NewFlagSet("MethodMultiNoPayload", flag.ExitOnError)

		serviceMultiMethodMultiPayloadFlags    = flag.NewFlagSet("MethodMultiPayload", flag.ExitOnError)
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
		case "ServiceMulti":
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
		case "ServiceMulti":
			switch epn {
			case "MethodMultiNoPayload":
				epf = serviceMultiMethodMultiNoPayloadFlags

			case "MethodMultiPayload":
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
		case "ServiceMulti":
			c := servicemultic.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodMultiNoPayload":
				endpoint = c.MethodMultiNoPayload()
				data = nil
			case "MethodMultiPayload":
				endpoint = c.MethodMultiPayload()
				data, err = servicemultic.BuildMethodMultiPayloadMethodMultiPayloadPayload(*serviceMultiMethodMultiPayloadBodyFlag, *serviceMultiMethodMultiPayloadBFlag, *serviceMultiMethodMultiPayloadAFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var MultiSimpleBuildCode = `// BuildMethodMultiSimplePayloadMethodMultiSimplePayloadPayload builds the
// payload for the ServiceMultiSimple1 MethodMultiSimplePayload endpoint from
// CLI flags.
func BuildMethodMultiSimplePayloadMethodMultiSimplePayloadPayload(serviceMultiSimple1MethodMultiSimplePayloadBody string) (*servicemultisimple1.MethodMultiSimplePayloadPayload, error) {
	var err error
	var body MethodMultiSimplePayloadRequestBody
	{
		err = json.Unmarshal([]byte(serviceMultiSimple1MethodMultiSimplePayloadBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'{\n      \"a\": false\n   }'")
		}
	}
	if err != nil {
		return nil, err
	}
	v := &servicemultisimple1.MethodMultiSimplePayloadPayload{
		A: body.A,
	}

	return v, nil
}
`

var MultiBuildCode = `// BuildMethodMultiPayloadMethodMultiPayloadPayload builds the payload for the
// ServiceMulti MethodMultiPayload endpoint from CLI flags.
func BuildMethodMultiPayloadMethodMultiPayloadPayload(serviceMultiMethodMultiPayloadBody string, serviceMultiMethodMultiPayloadB string, serviceMultiMethodMultiPayloadA string) (*servicemulti.MethodMultiPayloadPayload, error) {
	var err error
	var body MethodMultiPayloadRequestBody
	{
		err = json.Unmarshal([]byte(serviceMultiMethodMultiPayloadBody), &body)
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
			a = &val
			if err != nil {
				err = fmt.Errorf("invalid value for a, must be BOOL")
			}
		}
	}
	if err != nil {
		return nil, err
	}
	v := &servicemulti.MethodMultiPayloadPayload{}
	if body.C != nil {
		v.C = marshalUserTypeRequestBodyToUserType(body.C)
	}
	v.B = b
	v.A = a

	return v, nil
}
`
