// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// divider HTTP client CLI support package
//
// Command:
// $ goa gen goa.design/goa/examples/error/design -o
// $(GOPATH)/src/goa.design/goa/examples/error

package cli

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	goa "goa.design/goa"
	dividersvcc "goa.design/goa/examples/error/gen/http/divider/client"
	goahttp "goa.design/goa/http"
)

// UsageCommands returns the set of commands and sub-commands using the format
//
//    command (subcommand1|subcommand2|...)
//
func UsageCommands() string {
	return `divider (integer-divide|divide)
`
}

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return os.Args[0] + ` divider integer-divide --a 3601367395041194197 --b 8717444617646084941` + "\n" +
		""
}

// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(
	scheme, host string,
	doer goahttp.Doer,
	enc func(*http.Request) goahttp.Encoder,
	dec func(*http.Response) goahttp.Decoder,
	restore bool,
) (goa.Endpoint, interface{}, error) {
	var (
		dividerFlags = flag.NewFlagSet("divider", flag.ContinueOnError)

		dividerIntegerDivideFlags = flag.NewFlagSet("integer-divide", flag.ExitOnError)
		dividerIntegerDivideAFlag = dividerIntegerDivideFlags.String("a", "REQUIRED", "Left operand")
		dividerIntegerDivideBFlag = dividerIntegerDivideFlags.String("b", "REQUIRED", "Right operand")

		dividerDivideFlags = flag.NewFlagSet("divide", flag.ExitOnError)
		dividerDivideAFlag = dividerDivideFlags.String("a", "REQUIRED", "Left operand")
		dividerDivideBFlag = dividerDivideFlags.String("b", "REQUIRED", "Right operand")
	)
	dividerFlags.Usage = dividerUsage
	dividerIntegerDivideFlags.Usage = dividerIntegerDivideUsage
	dividerDivideFlags.Usage = dividerDivideUsage

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
		case "divider":
			svcf = dividerFlags
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
		case "divider":
			switch epn {
			case "integer-divide":
				epf = dividerIntegerDivideFlags

			case "divide":
				epf = dividerDivideFlags

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
		case "divider":
			c := dividersvcc.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "integer-divide":
				endpoint = c.IntegerDivide()
				data, err = dividersvcc.BuildIntegerDividePayload(*dividerIntegerDivideAFlag, *dividerIntegerDivideBFlag)
			case "divide":
				endpoint = c.Divide()
				data, err = dividersvcc.BuildDividePayload(*dividerDivideAFlag, *dividerDivideBFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}

// dividerUsage displays the usage of the divider command and its subcommands.
func dividerUsage() {
	fmt.Fprintf(os.Stderr, `Service is the divider service interface.
Usage:
    %s [globalflags] divider COMMAND [flags]

COMMAND:
    integer-divide: IntegerDivide implements integer_divide.
    divide: Divide implements divide.

Additional help:
    %s divider COMMAND --help
`, os.Args[0], os.Args[0])
}
func dividerIntegerDivideUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] divider integer-divide -a INT -b INT

IntegerDivide implements integer_divide.
    -a INT: Left operand
    -b INT: Right operand

Example:
    `+os.Args[0]+` divider integer-divide --a 3601367395041194197 --b 8717444617646084941
`, os.Args[0])
}

func dividerDivideUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] divider divide -a FLOAT64 -b FLOAT64

Divide implements divide.
    -a FLOAT64: Left operand
    -b FLOAT64: Right operand

Example:
    `+os.Args[0]+` divider divide --a 0.09666497128082843 --b 0.10225959553344194
`, os.Args[0])
}
