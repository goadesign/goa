// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// calc HTTP client CLI support package
//
// Command:
// $ goa gen goa.design/goa/examples/calc/design

package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	goa "goa.design/goa"
	calcsvcc "goa.design/goa/examples/calc/gen/http/calc/client"
	goahttp "goa.design/goa/http"
)

// UsageCommands returns the set of commands and sub-commands using the format
//
//    command (subcommand1|subcommand2|...)
//
func UsageCommands() string {
	return `calc (add|added)
`
}

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return os.Args[0] + ` calc add --a 686605435966370186 --b 8228676432890045784` + "\n" +
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
		calcFlags = flag.NewFlagSet("calc", flag.ContinueOnError)

		calcAddFlags = flag.NewFlagSet("add", flag.ExitOnError)
		calcAddAFlag = calcAddFlags.String("a", "REQUIRED", "Left operand")
		calcAddBFlag = calcAddFlags.String("b", "REQUIRED", "Right operand")

		calcAddedFlags = flag.NewFlagSet("added", flag.ExitOnError)
		calcAddedPFlag = calcAddedFlags.String("p", "REQUIRED", "map[string][]int is the payload type of the calc service added method.")
	)
	calcFlags.Usage = calcUsage
	calcAddFlags.Usage = calcAddUsage
	calcAddedFlags.Usage = calcAddedUsage

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
		case "calc":
			svcf = calcFlags
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
		case "calc":
			switch epn {
			case "add":
				epf = calcAddFlags

			case "added":
				epf = calcAddedFlags

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
		case "calc":
			c := calcsvcc.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "add":
				endpoint = c.Add()
				data, err = calcsvcc.BuildAddAddPayload(*calcAddAFlag, *calcAddBFlag)
			case "added":
				endpoint = c.Added()
				var err error
				var val map[string][]int
				err = json.Unmarshal([]byte(*calcAddedPFlag), &val)
				data = val
				if err != nil {
					return nil, nil, fmt.Errorf("invalid JSON for calcAddedPFlag, example of valid JSON:\n%s", "'{\n      \"Est ut.\": [\n         3793862871819669726,\n         8399553735696626949\n      ],\n      \"Sed non natus.\": [\n         1918630006328122782,\n         4288748512599820841,\n         4212629202012168060,\n         1698882017578366363\n      ]\n   }'")
				}
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}

// calcUsage displays the usage of the calc command and its subcommands.
func calcUsage() {
	fmt.Fprintf(os.Stderr, `The calc service performs operations on numbers
Usage:
    %s [globalflags] calc COMMAND [flags]

COMMAND:
    add: Add implements add.
    added: Added implements added.

Additional help:
    %s calc COMMAND --help
`, os.Args[0], os.Args[0])
}
func calcAddUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] calc add -a INT -b INT

Add implements add.
    -a INT: Left operand
    -b INT: Right operand

Example:
    `+os.Args[0]+` calc add --a 686605435966370186 --b 8228676432890045784
`, os.Args[0])
}

func calcAddedUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] calc added -p JSON

Added implements added.
    -p JSON: map[string][]int is the payload type of the calc service added method.

Example:
    `+os.Args[0]+` calc added --p '{
      "Est ut.": [
         3793862871819669726,
         8399553735696626949
      ],
      "Sed non natus.": [
         1918630006328122782,
         4288748512599820841,
         4212629202012168060,
         1698882017578366363
      ]
   }'
`, os.Args[0])
}
