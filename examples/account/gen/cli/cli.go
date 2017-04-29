package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"goa.design/goa.v2/examples/account/gen/service"
	genhttp "goa.design/goa.v2/examples/account/gen/transport/http"

	"goa.design/goa.v2"
)

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return `basiccli --addr http://localhost:8080 account list --filter foo
basiccli --addr http://localhost:8080 account show --id 1`
}

// UsageCommands returns a description listing the commands supported by the CLI
// tool.
func UsageCommands() string {
	return `basiccli account (create|list|show|delete)`
}

// RunCommand parses the command line and runs the appropriate command using the
// given clients. It returns the decoded result.
func RunCommand(timeout int, accountClient *genhttp.AccountClient) (interface{}, error) {
	var (
		accountFlags = flag.NewFlagSet("account", flag.ContinueOnError)

		accountCreateFlags     = flag.NewFlagSet("account-create", flag.ExitOnError)
		accountCreateNameFlag  = accountCreateFlags.String("name", "REQUIRED", "account name")
		accountCreateOrgIDFlag = accountCreateFlags.Uint("org-id", 0, "ID of organization that owns newly created account")

		accountListFlags      = flag.NewFlagSet("account-list", flag.ExitOnError)
		accountListFilterFlag = accountListFlags.String("filter", "", "Filter is the account name prefix filter")
		accountListOrgIDFlag  = accountListFlags.Uint("org-id", 0, "ID of organization that owns account")

		accountShowFlags     = flag.NewFlagSet("account-show", flag.ExitOnError)
		accountShowIDFlag    = accountShowFlags.String("id", "REQUIRED", "ID of account")
		accountShowOrgIDFlag = accountShowFlags.Uint("org-id", 0, "ID of organization that owns account")

		accountDeleteFlags     = flag.NewFlagSet("account-delete", flag.ExitOnError)
		accountDeleteIDFlag    = accountDeleteFlags.String("id", "REQUIRED", "ID of account")
		accountDeleteOrgIDFlag = accountDeleteFlags.Uint("org-id", 0, "ID of organization that owns account")
	)
	accountFlags.Usage = accountUsage
	accountCreateFlags.Usage = accountCreateUsage
	accountListFlags.Usage = accountListUsage
	accountShowFlags.Usage = accountShowUsage
	accountDeleteFlags.Usage = accountDeleteUsage

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if len(os.Args) < flag.NFlag()+3 {
		return nil, fmt.Errorf("not enough arguments")
	}

	var (
		service      string
		serviceFlags *flag.FlagSet
	)
	{
		service = os.Args[1+flag.NFlag()]
		switch service {
		case "account":
			serviceFlags = accountFlags
		default:
			return nil, fmt.Errorf("unknown service %#v", service)
		}
	}

	var (
		endpoint      string
		endpointFlags *flag.FlagSet
	)
	{
		if err := serviceFlags.Parse(os.Args[2+flag.NFlag():]); err != nil {
			return nil, err
		}
		endpoint = os.Args[2+flag.NFlag()+serviceFlags.NFlag()]
		switch endpoint {
		case "create":
			endpointFlags = accountCreateFlags
		case "list":
			endpointFlags = accountListFlags
		case "show":
			endpointFlags = accountShowFlags
		case "delete":
			endpointFlags = accountDeleteFlags
		default:
			return nil, fmt.Errorf("unknown %s endpoint %#v", service, endpoint)
		}

		if len(os.Args) > 2+flag.NFlag()+serviceFlags.NFlag() {
			if err := endpointFlags.Parse(os.Args[3+flag.NFlag()+serviceFlags.NFlag():]); err != nil {
				return nil, err
			}
		}
	}

	var (
		data interface{}
		err  error
	)
	{
		ctx, _ := context.WithDeadline(context.Background(),
			time.Now().Add(time.Duration(timeout)*time.Second))
		switch service {
		case "account":
			switch endpoint {
			case "create":
				data, err = runAccountCreate(ctx, accountClient.Create(), *accountCreateNameFlag, *accountCreateOrgIDFlag)
			case "list":
				data, err = runAccountList(ctx, accountClient.List(), accountListFilterFlag, *accountListOrgIDFlag)
			case "show":
				data, err = runAccountShow(ctx, accountClient.Show(), *accountShowIDFlag, *accountShowOrgIDFlag)
			case "delete":
				data, err = runAccountDelete(ctx, accountClient.Delete(), *accountDeleteIDFlag, *accountDeleteOrgIDFlag)
			}
		}
	}

	return data, err
}

func runAccountCreate(ctx context.Context, endpoint goa.Endpoint, name string, orgID uint) (interface{}, error) {
	payload := service.CreateAccount{
		Name:  name,
		OrgID: orgID,
	}
	return endpoint(ctx, &payload)
}

func runAccountList(ctx context.Context, endpoint goa.Endpoint, filter *string, orgID uint) (interface{}, error) {
	payload := service.ListAccount{
		Filter: filter,
		OrgID:  orgID,
	}

	return endpoint(ctx, &payload)
}

func runAccountShow(ctx context.Context, endpoint goa.Endpoint, id string, orgID uint) (interface{}, error) {
	payload := service.ShowAccountPayload{
		ID:    id,
		OrgID: orgID,
	}

	return endpoint(ctx, &payload)
}

func runAccountDelete(ctx context.Context, endpoint goa.Endpoint, id string, orgID uint) (interface{}, error) {
	payload := service.DeleteAccountPayload{
		ID:    id,
		OrgID: orgID,
	}

	return endpoint(ctx, &payload)
}

func accountUsage() {
	fmt.Fprintf(os.Stderr, `Manage accounts
Usage:
    %s [globalflags] account COMMAND [flags]

COMMAND:
    create: Create new account
    list: List all accounts
    show: Show account by ID
    delete: Delete account by ID

Additional help:
    %s account COMMAND --help
`, os.Args[0], os.Args[0])
}

func accountCreateUsage() {
	fmt.Fprintf(os.Stderr, `Create new account
Usage:
    %s [flags] account create --org-id INT --name STRING

--org-id INT: ID of organization that owns newly created account
--name STRING: Name of new account
`, os.Args[0])
}

func accountListUsage() {
	fmt.Fprintf(os.Stderr, `List all accounts
Usage:
    %s [flags] account list --filter STRING

--filter STRING: Filter is the account name prefix filter
`, os.Args[0])
}

func accountShowUsage() {
	fmt.Fprintf(os.Stderr, `Show account by ID
Usage:
    %s [flags] account show --id STRING

--id STRING: ID of account
`, os.Args[0])
}

func accountDeleteUsage() {
	fmt.Fprintf(os.Stderr, `Delete account by ID
Usage:
    %s [flags] account delete --id STRING

--id STRING: ID of account
`, os.Args[0])
}
