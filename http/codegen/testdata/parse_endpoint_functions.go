package testdata

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
				data, err = servicemultisimple1c.BuildMethodMultiSimplePayloadPayload(*serviceMultiSimple1MethodMultiSimplePayloadBodyFlag)
			}
		case "ServiceMultiSimple2":
			c := servicemultisimple2c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodMultiSimpleNoPayload":
				endpoint = c.MethodMultiSimpleNoPayload()
				data = nil
			case "MethodMultiSimplePayload":
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

var MultiRequiredPayloadParseCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		serviceMultiRequired1Flags = flag.NewFlagSet("ServiceMultiRequired1", flag.ContinueOnError)

		serviceMultiRequired1MethodMultiRequiredPayloadFlags    = flag.NewFlagSet("MethodMultiRequiredPayload", flag.ExitOnError)
		serviceMultiRequired1MethodMultiRequiredPayloadBodyFlag = serviceMultiRequired1MethodMultiRequiredPayloadFlags.String("body", "REQUIRED", "")

		serviceMultiRequired2Flags = flag.NewFlagSet("ServiceMultiRequired2", flag.ContinueOnError)

		serviceMultiRequired2MethodMultiRequiredNoPayloadFlags = flag.NewFlagSet("MethodMultiRequiredNoPayload", flag.ExitOnError)

		serviceMultiRequired2MethodMultiRequiredPayloadFlags = flag.NewFlagSet("MethodMultiRequiredPayload", flag.ExitOnError)
		serviceMultiRequired2MethodMultiRequiredPayloadAFlag = serviceMultiRequired2MethodMultiRequiredPayloadFlags.String("a", "REQUIRED", "")
	)
	serviceMultiRequired1Flags.Usage = serviceMultiRequired1Usage
	serviceMultiRequired1MethodMultiRequiredPayloadFlags.Usage = serviceMultiRequired1MethodMultiRequiredPayloadUsage

	serviceMultiRequired2Flags.Usage = serviceMultiRequired2Usage
	serviceMultiRequired2MethodMultiRequiredNoPayloadFlags.Usage = serviceMultiRequired2MethodMultiRequiredNoPayloadUsage
	serviceMultiRequired2MethodMultiRequiredPayloadFlags.Usage = serviceMultiRequired2MethodMultiRequiredPayloadUsage

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
		case "ServiceMultiRequired1":
			svcf = serviceMultiRequired1Flags
		case "ServiceMultiRequired2":
			svcf = serviceMultiRequired2Flags
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
		case "ServiceMultiRequired1":
			switch epn {
			case "MethodMultiRequiredPayload":
				epf = serviceMultiRequired1MethodMultiRequiredPayloadFlags

			}

		case "ServiceMultiRequired2":
			switch epn {
			case "MethodMultiRequiredNoPayload":
				epf = serviceMultiRequired2MethodMultiRequiredNoPayloadFlags

			case "MethodMultiRequiredPayload":
				epf = serviceMultiRequired2MethodMultiRequiredPayloadFlags

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
		case "ServiceMultiRequired1":
			c := servicemultirequired1c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodMultiRequiredPayload":
				endpoint = c.MethodMultiRequiredPayload()
				data, err = servicemultirequired1c.BuildMethodMultiRequiredPayloadPayload(*serviceMultiRequired1MethodMultiRequiredPayloadBodyFlag)
			}
		case "ServiceMultiRequired2":
			c := servicemultirequired2c.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodMultiRequiredNoPayload":
				endpoint = c.MethodMultiRequiredNoPayload()
				data = nil
			case "MethodMultiRequiredPayload":
				endpoint = c.MethodMultiRequiredPayload()
				data, err = servicemultirequired2c.BuildMethodMultiRequiredPayloadPayload(*serviceMultiRequired2MethodMultiRequiredPayloadAFlag)
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

var MultiBuildCode = `// BuildMethodMultiPayloadPayload builds the payload for the ServiceMulti
// MethodMultiPayload endpoint from CLI flags.
func BuildMethodMultiPayloadPayload(serviceMultiMethodMultiPayloadBody string, serviceMultiMethodMultiPayloadB string, serviceMultiMethodMultiPayloadA string) (*servicemulti.MethodMultiPayloadPayload, error) {
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

var QueryBoolBuildCode = `// BuildMethodQueryBoolPayload builds the payload for the ServiceQueryBool
// MethodQueryBool endpoint from CLI flags.
func BuildMethodQueryBoolPayload(serviceQueryBoolMethodQueryBoolQ string) (*servicequerybool.MethodQueryBoolPayload, error) {
	var err error
	var q *bool
	{
		if serviceQueryBoolMethodQueryBoolQ != "" {
			val, err := strconv.ParseBool(serviceQueryBoolMethodQueryBoolQ)
			q = &val
			if err != nil {
				err = fmt.Errorf("invalid value for q, must be BOOL")
			}
		}
	}
	if err != nil {
		return nil, err
	}
	payload := &servicequerybool.MethodQueryBoolPayload{
		Q: q,
	}
	return payload, nil
}
`

var BodyQueryPathObjectBuildCode = `// BuildMethodBodyQueryPathObjectPayload builds the payload for the
// ServiceBodyQueryPathObject MethodBodyQueryPathObject endpoint from CLI flags.
func BuildMethodBodyQueryPathObjectPayload(serviceBodyQueryPathObjectMethodBodyQueryPathObjectBody string, serviceBodyQueryPathObjectMethodBodyQueryPathObjectC string, serviceBodyQueryPathObjectMethodBodyQueryPathObjectB string) (*servicebodyquerypathobject.MethodBodyQueryPathObjectPayload, error) {
	var err error
	var body MethodBodyQueryPathObjectRequestBody
	{
		err = json.Unmarshal([]byte(serviceBodyQueryPathObjectMethodBodyQueryPathObjectBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'{\n      \"a\": \"Ullam aut.\"\n   }'")
		}
	}
	var c *string
	{
		if serviceBodyQueryPathObjectMethodBodyQueryPathObjectC != "" {
			c = &serviceBodyQueryPathObjectMethodBodyQueryPathObjectC
		}
	}
	var b *string
	{
		if serviceBodyQueryPathObjectMethodBodyQueryPathObjectB != "" {
			b = &serviceBodyQueryPathObjectMethodBodyQueryPathObjectB
		}
	}
	if err != nil {
		return nil, err
	}
	v := &servicebodyquerypathobject.MethodBodyQueryPathObjectPayload{
		A: body.A,
	}
	v.C = c
	v.B = b
	return v, nil
}
`

var PayloadPrimitiveTypeParseCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		serviceBodyPrimitiveBoolValidateFlags = flag.NewFlagSet("ServiceBodyPrimitiveBoolValidate", flag.ContinueOnError)

		serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidateFlags = flag.NewFlagSet("MethodBodyPrimitiveBoolValidate", flag.ExitOnError)
		serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidatePFlag = serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidateFlags.String("p", "REQUIRED", "bool is the payload type of the ServiceBodyPrimitiveBoolValidate service MethodBodyPrimitiveBoolValidate method.")
	)
	serviceBodyPrimitiveBoolValidateFlags.Usage = serviceBodyPrimitiveBoolValidateUsage
	serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidateFlags.Usage = serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidateUsage

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
		case "ServiceBodyPrimitiveBoolValidate":
			svcf = serviceBodyPrimitiveBoolValidateFlags
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
		case "ServiceBodyPrimitiveBoolValidate":
			switch epn {
			case "MethodBodyPrimitiveBoolValidate":
				epf = serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidateFlags

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
		case "ServiceBodyPrimitiveBoolValidate":
			c := servicebodyprimitiveboolvalidatec.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodBodyPrimitiveBoolValidate":
				endpoint = c.MethodBodyPrimitiveBoolValidate()
				var err error
				data, err = strconv.ParseBool(*serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidatePFlag)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid value for serviceBodyPrimitiveBoolValidateMethodBodyPrimitiveBoolValidatePFlag, must be BOOL")
				}
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var PayloadArrayPrimitiveTypeParseCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		serviceBodyPrimitiveArrayStringValidateFlags = flag.NewFlagSet("ServiceBodyPrimitiveArrayStringValidate", flag.ContinueOnError)

		serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidateFlags = flag.NewFlagSet("MethodBodyPrimitiveArrayStringValidate", flag.ExitOnError)
		serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidatePFlag = serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidateFlags.String("p", "REQUIRED", "[]string is the payload type of the ServiceBodyPrimitiveArrayStringValidate service MethodBodyPrimitiveArrayStringValidate method.")
	)
	serviceBodyPrimitiveArrayStringValidateFlags.Usage = serviceBodyPrimitiveArrayStringValidateUsage
	serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidateFlags.Usage = serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidateUsage

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
		case "ServiceBodyPrimitiveArrayStringValidate":
			svcf = serviceBodyPrimitiveArrayStringValidateFlags
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
		case "ServiceBodyPrimitiveArrayStringValidate":
			switch epn {
			case "MethodBodyPrimitiveArrayStringValidate":
				epf = serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidateFlags

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
		case "ServiceBodyPrimitiveArrayStringValidate":
			c := servicebodyprimitivearraystringvalidatec.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MethodBodyPrimitiveArrayStringValidate":
				endpoint = c.MethodBodyPrimitiveArrayStringValidate()
				var err error
				var val []string
				err = json.Unmarshal([]byte(*serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidatePFlag), &val)
				data = val
				if err != nil {
					return nil, nil, fmt.Errorf("invalid JSON for serviceBodyPrimitiveArrayStringValidateMethodBodyPrimitiveArrayStringValidatePFlag, example of valid JSON:\n%s", "'[\n      \"val\",\n      \"val\",\n      \"val\"\n   ]'")
				}
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var PayloadArrayUserTypeBuildCode = `// BuildMethodBodyInlineArrayUserPayload builds the payload for the
// ServiceBodyInlineArrayUser MethodBodyInlineArrayUser endpoint from CLI flags.
func BuildMethodBodyInlineArrayUserPayload(serviceBodyInlineArrayUserMethodBodyInlineArrayUserBody string) ([]*servicebodyinlinearrayuser.ElemType, error) {
	var err error
	var body []*ElemTypeRequestBody
	{
		err = json.Unmarshal([]byte(serviceBodyInlineArrayUserMethodBodyInlineArrayUserBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'[\n      {\n         \"a\": \"patterna\",\n         \"b\": \"patternb\"\n      },\n      {\n         \"a\": \"patterna\",\n         \"b\": \"patternb\"\n      }\n   ]'")
		}
	}
	if err != nil {
		return nil, err
	}
	v := make([]*servicebodyinlinearrayuser.ElemType, len(body))
	for i, val := range body {
		v[i] = &servicebodyinlinearrayuser.ElemType{
			A: val.A,
			B: val.B,
		}
	}
	return v, nil
}
`

var PayloadMapUserTypeBuildCode = `// BuildMethodBodyInlineMapUserPayload builds the payload for the
// ServiceBodyInlineMapUser MethodBodyInlineMapUser endpoint from CLI flags.
func BuildMethodBodyInlineMapUserPayload(serviceBodyInlineMapUserMethodBodyInlineMapUserBody string) (map[*servicebodyinlinemapuser.KeyType]*servicebodyinlinemapuser.ElemType, error) {
	var err error
	var body map[*KeyTypeRequestBody]*ElemTypeRequestBody
	{
		err = json.Unmarshal([]byte(serviceBodyInlineMapUserMethodBodyInlineMapUserBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "null")
		}
	}
	if err != nil {
		return nil, err
	}
	v := make(map[*servicebodyinlinemapuser.KeyType]*servicebodyinlinemapuser.ElemType, len(body))
	for key, val := range body {
		tk := &servicebodyinlinemapuser.KeyType{
			A: key.A,
			B: key.B,
		}
		tv := &servicebodyinlinemapuser.ElemType{
			A: val.A,
			B: val.B,
		}
		v[tk] = tv
	}
	return v, nil
}
`

var MapQueryParseCode = `// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		serviceMapQueryPrimitiveArrayFlags = flag.NewFlagSet("ServiceMapQueryPrimitiveArray", flag.ContinueOnError)

		serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayFlags = flag.NewFlagSet("MapQueryPrimitiveArray", flag.ExitOnError)
		serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayPFlag = serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayFlags.String("p", "REQUIRED", "map[string][]uint is the payload type of the ServiceMapQueryPrimitiveArray service MapQueryPrimitiveArray method.")
	)
	serviceMapQueryPrimitiveArrayFlags.Usage = serviceMapQueryPrimitiveArrayUsage
	serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayFlags.Usage = serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayUsage

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
		case "ServiceMapQueryPrimitiveArray":
			svcf = serviceMapQueryPrimitiveArrayFlags
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
		case "ServiceMapQueryPrimitiveArray":
			switch epn {
			case "MapQueryPrimitiveArray":
				epf = serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayFlags

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
		case "ServiceMapQueryPrimitiveArray":
			c := servicemapqueryprimitivearrayc.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "MapQueryPrimitiveArray":
				endpoint = c.MapQueryPrimitiveArray()
				var err error
				var val map[string][]uint
				err = json.Unmarshal([]byte(*serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayPFlag), &val)
				data = val
				if err != nil {
					return nil, nil, fmt.Errorf("invalid JSON for serviceMapQueryPrimitiveArrayMapQueryPrimitiveArrayPFlag, example of valid JSON:\n%s", "'{\n      \"Iste perspiciatis.\": [\n         567408540461384614,\n         5721637919286150856\n      ],\n      \"Itaque inventore optio.\": [\n         944964629895926327,\n         593430823343775997\n      ],\n      \"Molestias recusandae doloribus qui quia.\": [\n         6921210467234244263,\n         3742304935485895874,\n         4170793618430505438,\n         7388093990298529880\n      ]\n   }'")
				}
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}
`

var MapQueryObjectBuildCode = `// BuildMethodMapQueryObjectPayload builds the payload for the
// ServiceMapQueryObject MethodMapQueryObject endpoint from CLI flags.
func BuildMethodMapQueryObjectPayload(serviceMapQueryObjectMethodMapQueryObjectBody string, serviceMapQueryObjectMethodMapQueryObjectA string, serviceMapQueryObjectMethodMapQueryObjectC string) (*servicemapqueryobject.PayloadType, error) {
	var err error
	var body MethodMapQueryObjectRequestBody
	{
		err = json.Unmarshal([]byte(serviceMapQueryObjectMethodMapQueryObjectBody), &body)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'{\n      \"b\": \"patternb\"\n   }'")
		}
		if body.B != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("body.b", *body.B, "patternb"))
		}
		if err != nil {
			return nil, err
		}
	}
	var a string
	{
		a = serviceMapQueryObjectMethodMapQueryObjectA
	}
	var c map[int][]string
	{
		err = json.Unmarshal([]byte(serviceMapQueryObjectMethodMapQueryObjectC), &c)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for c, example of valid JSON:\n%s", "'{\n      \"1484745265794365762\": [\n         \"Similique aspernatur.\",\n         \"Error explicabo.\",\n         \"Minima cumque voluptatem et distinctio aliquam.\",\n         \"Blanditiis ut eaque.\"\n      ],\n      \"4925854623691091547\": [\n         \"Eos aut ipsam.\",\n         \"Aliquam tempora.\"\n      ],\n      \"7174751143827362498\": [\n         \"Facilis minus explicabo nemo eos vel repellat.\",\n         \"Voluptatum magni aperiam qui.\"\n      ]\n   }'")
		}
		err = goa.MergeErrors(err, goa.ValidatePattern("c.a", c.A, "patterna"))
		if c.B != nil {
			err = goa.MergeErrors(err, goa.ValidatePattern("c.b", *c.B, "patternb"))
		}
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	v := &servicemapqueryobject.PayloadType{
		B: body.B,
	}
	v.A = a
	v.C = c
	return v, nil
}
`

var QueryUInt32BuildCode = `// BuildMethodQueryUInt32Payload builds the payload for the ServiceQueryUInt32
// MethodQueryUInt32 endpoint from CLI flags.
func BuildMethodQueryUInt32Payload(serviceQueryUInt32MethodQueryUInt32Q string) (*servicequeryuint32.MethodQueryUInt32Payload, error) {
	var err error
	var q *uint32
	{
		if serviceQueryUInt32MethodQueryUInt32Q != "" {
			var v uint64
			v, err = strconv.ParseUint(serviceQueryUInt32MethodQueryUInt32Q, 10, 32)
			val := uint32(v)
			q = &val
			if err != nil {
				err = fmt.Errorf("invalid value for q, must be UINT32")
			}
		}
	}
	if err != nil {
		return nil, err
	}
	payload := &servicequeryuint32.MethodQueryUInt32Payload{
		Q: q,
	}
	return payload, nil
}
`

var QueryUIntBuildCode = `// BuildMethodQueryUIntPayload builds the payload for the ServiceQueryUInt
// MethodQueryUInt endpoint from CLI flags.
func BuildMethodQueryUIntPayload(serviceQueryUIntMethodQueryUIntQ string) (*servicequeryuint.MethodQueryUIntPayload, error) {
	var err error
	var q *uint
	{
		if serviceQueryUIntMethodQueryUIntQ != "" {
			var v uint64
			v, err = strconv.ParseUint(serviceQueryUIntMethodQueryUIntQ, 10, 64)
			val := uint(v)
			q = &val
			if err != nil {
				err = fmt.Errorf("invalid value for q, must be UINT")
			}
		}
	}
	if err != nil {
		return nil, err
	}
	payload := &servicequeryuint.MethodQueryUIntPayload{
		Q: q,
	}
	return payload, nil
}
`
