package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"goa.design/goa.v2"
	goahttp "goa.design/goa.v2/http"

	sommelierc "goa.design/goa.v2/examples/cellar/gen/http/sommelier/client"
	storagec "goa.design/goa.v2/examples/cellar/gen/http/storage/client"
)

// UsageCommands returns the set of commands and sub-commands using the format
//
//    command (subcommand1|subcommand2|...)
//
func UsageCommands() string {
	return `storage (list|show|add|remove)
sommelier pick`
}

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return os.Args[0] + ` --addr http://localhost:8080 storage list` + "\n" +
		os.Args[0] + ` --addr http://localhost:8080 storage show --id 1` + "\n"
}

// ParseEndpoint returns the endpoint and payload as specified on the command
// line.
func ParseEndpoint(scheme, host string, doer goahttp.Doer, enc func(*http.Request) goahttp.Encoder, dec func(*http.Response) goahttp.Decoder, restoreBody bool) (goa.Endpoint, interface{}, error) {
	var (
		storageFlags = flag.NewFlagSet("storage", flag.ContinueOnError)

		storageListFlags = flag.NewFlagSet("storage-list", flag.ExitOnError)

		storageShowFlags  = flag.NewFlagSet("storage-show", flag.ExitOnError)
		storageShowIDFlag = storageShowFlags.String("id", "REQUIRED", "ID of bottle")

		storageAddFlags           = flag.NewFlagSet("storage-add", flag.ExitOnError)
		storageAddNameFlag        = storageAddFlags.String("name", "REQUIRED", "Name of bottle")
		storageAddWineryFlag      = storageAddFlags.String("winery", "REQUIRED", "Winery that produces wine (as JSON string)")
		storageAddVintageFlag     = storageAddFlags.String("vintage", "REQUIRED", "Vintage of bottle")
		storageAddCompositionFlag = storageAddFlags.String("composition", "", "Composition is the list of grape varietals and associated percentage.)")
		storageAddDescriptionFlag = storageAddFlags.String("description", "", "Description of bottle")
		storageAddRatingFlag      = storageAddFlags.String("rating", "", "Rating of bottle from 1 (worst) to 5 (best)")

		storageRemoveFlags  = flag.NewFlagSet("storage-remove", flag.ExitOnError)
		storageRemoveIDFlag = storageRemoveFlags.String("id", "REQUIRED", "ID of bottle")

		sommelierFlags = flag.NewFlagSet("sommelier", flag.ContinueOnError)

		sommelierPickFlags        = flag.NewFlagSet("sommelier-pick", flag.ExitOnError)
		sommelierPickNameFlag     = sommelierPickFlags.String("name", "", "Name of bottle")
		sommelierPickVarietalFlag = sommelierPickFlags.String("varietal", "", "Varietal is the list of grape varietals and associated percentage.)")
		sommelierPickWineryFlag   = sommelierPickFlags.String("winery", "", "Winery of bottle to pick")
	)
	storageFlags.Usage = storageUsage
	storageAddFlags.Usage = storageAddUsage
	storageListFlags.Usage = storageListUsage
	storageShowFlags.Usage = storageShowUsage
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
		case "storage":
			svcf = storageFlags
		case "sommelier":
			svcf = sommelierFlags
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
		case "storage":
			switch epn {
			case "add":
				epf = storageAddFlags
			case "list":
				epf = storageListFlags
			case "show":
				epf = storageShowFlags
			case "remove":
				epf = storageRemoveFlags
			}
		case "sommelier":
			switch epn {
			case "pick":
				epf = sommelierPickFlags
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
		case "storage":
			c := storagec.NewClient(scheme, host, doer, enc, dec, restoreBody)
			switch epn {
			case "add":
				endpoint = c.Add()
				data, err = buildStorageAddPayload(*storageAddNameFlag, *storageAddWineryFlag, *storageAddVintageFlag, *storageAddCompositionFlag, *storageAddDescriptionFlag, *storageAddRatingFlag)
			case "list":
				endpoint = c.List()
				data = nil
			case "show":
				endpoint = c.Show()
				data = *storageShowIDFlag
			case "remove":
				endpoint = c.Remove()
				data = *storageRemoveIDFlag
			}
		case "sommelier":
			c := sommelierc.NewClient(scheme, host, doer, enc, dec, restoreBody)
			switch epn {
			case "pick":
				endpoint = c.Pick()
				data, err = buildSommelierPickPayload(*sommelierPickNameFlag, *sommelierPickWineryFlag, *sommelierPickVarietalFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}

func buildStorageAddPayload(nameFlag, wineryFlag, vintageFlag, compositionFlag, descriptionFlag, ratingFlag string) (*storagec.AddRequestBody, error) {
	var winery storagec.WineryRequestBody
	{
		err := json.Unmarshal([]byte(wineryFlag), &winery)
		if err != nil {
			ex := storagec.WineryRequestBody{} // ...
			js, _ := json.Marshal(ex)
			return nil, fmt.Errorf("invalid JSON for winery, example of valid JSON:\n%s", js)
		}
	}

	var composition []*storagec.ComponentRequestBody
	if compositionFlag != "" {
		err := json.Unmarshal([]byte(compositionFlag), &composition)
		if err != nil {
			ex := []*storagec.ComponentRequestBody{} // ...
			js, _ := json.Marshal(ex)
			return nil, fmt.Errorf("invalid JSON for composition, example of valid JSON:\n%s", string(js))
		}
	}

	var vintage uint32
	{
		if v, err := strconv.ParseUint(ratingFlag, 10, 32); err == nil {
			vintage = uint32(v)
		}
	}

	var rating *uint32
	if ratingFlag != "" {
		if v, err := strconv.ParseUint(ratingFlag, 10, 32); err == nil {
			val := uint32(v)
			rating = &val
		}
	}

	var description *string
	if descriptionFlag != "" {
		description = &descriptionFlag
	}

	body := &storagec.AddRequestBody{
		Name:        nameFlag,
		Winery:      &winery,
		Vintage:     vintage,
		Composition: composition,
		Description: description,
		Rating:      rating,
	}

	return body, nil
}

func buildSommelierPickPayload(nameFlag, wineryFlag, varietalFlag string) (*sommelierc.PickRequestBody, error) {
	var name *string
	if nameFlag != "" {
		name = &nameFlag
	}

	var winery *string
	if wineryFlag != "" {
		winery = &wineryFlag
	}

	var varietal []string
	if varietalFlag != "" {
		err := json.Unmarshal([]byte(varietalFlag), &varietal)
		if err != nil {
			ex := []string{"pinot noir"}
			js, _ := json.Marshal(ex)
			return nil, fmt.Errorf("invalid JSON for varietal, example of valid JSON:\n%s", js)
		}
	}

	body := &sommelierc.PickRequestBody{
		Name:     name,
		Winery:   winery,
		Varietal: varietal,
	}

	return body, nil
}

func storageUsage() {
	fmt.Fprintf(os.Stderr, `Manage storages
Usage:
    %s [globalflags] storage COMMAND [flags]

COMMAND:
    add: Add new storage
    list: List all storages
    show: Show storage by ID
    remove: Remove storage by ID

Additional help:
    %s storage COMMAND --help
`, os.Args[0], os.Args[0])
}

func storageAddUsage() {
	fmt.Fprintf(os.Stderr, `Add new storage
Usage:
    %s [flags] storage add --org-id INT --name STRING

--org-id INT: ID of organization that owns newly addd storage
--name STRING: Name of new storage
`, os.Args[0])
}

func storageListUsage() {
	fmt.Fprintf(os.Stderr, `List all storages
Usage:
    %s [flags] storage list --filter STRING

--filter STRING: Filter is the storage name prefix filter
`, os.Args[0])
}

func storageShowUsage() {
	fmt.Fprintf(os.Stderr, `Show storage by ID
Usage:
    %s [flags] storage show --id STRING

--id STRING: ID of storage
`, os.Args[0])
}

func storageRemoveUsage() {
	fmt.Fprintf(os.Stderr, `Remove storage by ID
Usage:
    %s [flags] storage remove --id STRING

--id STRING: ID of storage
`, os.Args[0])
}
