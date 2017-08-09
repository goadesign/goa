// Code generated with goa v2.0.0-wip, DO NOT EDIT.
//
// cellar HTTP client CLI support package
//
// Command:
// $ goa gen goa.design/goa.v2/examples/cellar/design

package cli

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	goa "goa.design/goa.v2"
	sommelierc "goa.design/goa.v2/examples/cellar/gen/http/sommelier/client"
	storagec "goa.design/goa.v2/examples/cellar/gen/http/storage/client"
	goahttp "goa.design/goa.v2/http"
)

// UsageCommands returns the set of commands and sub-commands using the format
//
//    command (subcommand1|subcommand2|...)
//
func UsageCommands() string {
	return `sommelier pick
storage (list|show|add|remove)
`
}

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return os.Args[0] + ` sommelier pick --body {
      "name": "Corrupti qui enim repellendus laboriosam accusamus.",
      "varietal": [
         "Aut reiciendis ea architecto magni tempora fugiat.",
         "Delectus vel earum doloribus.",
         "Consequuntur recusandae.",
         "Nemo omnis quae suscipit laudantium quis sapiente."
      ],
      "winery": "Sunt dolorem."
   }` + "\n" +
		os.Args[0] + ` storage list` + "\n" +
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
		sommelierFlags = flag.NewFlagSet("sommelier", flag.ContinueOnError)

		sommelierPickFlags    = flag.NewFlagSet("pick", flag.ExitOnError)
		sommelierPickBodyFlag = sommelierPickFlags.String("body", "REQUIRED", "")

		storageFlags = flag.NewFlagSet("storage", flag.ContinueOnError)

		storageListFlags = flag.NewFlagSet("list", flag.ExitOnError)

		storageShowFlags  = flag.NewFlagSet("show", flag.ExitOnError)
		storageShowIDFlag = storageShowFlags.String("id", "REQUIRED", "ID of bottle to show")

		storageAddFlags    = flag.NewFlagSet("add", flag.ExitOnError)
		storageAddBodyFlag = storageAddFlags.String("body", "REQUIRED", "")

		storageRemoveFlags  = flag.NewFlagSet("remove", flag.ExitOnError)
		storageRemoveIDFlag = storageRemoveFlags.String("id", "REQUIRED", "ID of bottle to remove")
	)
	sommelierFlags.Usage = sommelierUsage
	sommelierPickFlags.Usage = sommelierPickUsage

	storageFlags.Usage = storageUsage
	storageListFlags.Usage = storageListUsage
	storageShowFlags.Usage = storageShowUsage
	storageAddFlags.Usage = storageAddUsage
	storageRemoveFlags.Usage = storageRemoveUsage

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
		case "sommelier":
			svcf = sommelierFlags
		case "storage":
			svcf = storageFlags
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
		case "sommelier":
			switch epn {
			case "pick":
				epf = sommelierPickFlags

			}

		case "storage":
			switch epn {
			case "list":
				epf = storageListFlags

			case "show":
				epf = storageShowFlags

			case "add":
				epf = storageAddFlags

			case "remove":
				epf = storageRemoveFlags

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
		case "sommelier":
			c := sommelierc.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "pick":
				endpoint = c.Pick()
				data, err = sommelierc.BuildCriteria(*sommelierPickBodyFlag)
			}
		case "storage":
			c := storagec.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "list":
				endpoint = c.List()
				data = nil
			case "show":
				endpoint = c.Show()
				data, err = storagec.BuildShowPayload(*storageShowIDFlag)
			case "add":
				endpoint = c.Add()
				data, err = storagec.BuildBottle(*storageAddBodyFlag)
			case "remove":
				endpoint = c.Remove()
				data, err = storagec.BuildRemovePayload(*storageRemoveIDFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}

// sommelierUsage displays the usage of the sommelier command and its
// subcommands.
func sommelierUsage() {
	fmt.Fprintf(os.Stderr, `The sommelier service retrieves bottles given a set of criteria.
Usage:
    %s [globalflags] sommelier COMMAND [flags]

COMMAND:
    pick: Pick implements pick.

Additional help:
    %s sommelier COMMAND --help
`, os.Args[0], os.Args[0])
}
func sommelierPickUsage() {
	fmt.Fprintf(os.Stderr, `Pick implements pick.
Usage:
    %s [flags] sommelier pick --body JSON
body JSON: 
`, os.Args[0])
}

// storageUsage displays the usage of the storage command and its subcommands.
func storageUsage() {
	fmt.Fprintf(os.Stderr, `The storage service makes it possible to view, add or remove wine bottles.
Usage:
    %s [globalflags] storage COMMAND [flags]

COMMAND:
    list: List all stored bottles
    show: Show bottle by ID
    add: Add new bottle and return its ID.
    remove: Remove bottle from storage

Additional help:
    %s storage COMMAND --help
`, os.Args[0], os.Args[0])
}
func storageListUsage() {
	fmt.Fprintf(os.Stderr, `List all stored bottles
Usage:
    %s [flags] storage list
`, os.Args[0])
}

func storageShowUsage() {
	fmt.Fprintf(os.Stderr, `Show bottle by ID
Usage:
    %s [flags] storage show --id STRING
id STRING: ID of bottle to show
`, os.Args[0])
}

func storageAddUsage() {
	fmt.Fprintf(os.Stderr, `Add new bottle and return its ID.
Usage:
    %s [flags] storage add --body JSON
body JSON: 
`, os.Args[0])
}

func storageRemoveUsage() {
	fmt.Fprintf(os.Stderr, `Remove bottle from storage
Usage:
    %s [flags] storage remove --id STRING
id STRING: ID of bottle to remove
`, os.Args[0])
}
