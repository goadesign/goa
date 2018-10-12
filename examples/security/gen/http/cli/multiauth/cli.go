// Code generated by goa v2.0.0-wip, DO NOT EDIT.
//
// multiauth HTTP client CLI support package
//
// Command:
// $ goa gen goa.design/goa/examples/security/design

package cli

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	goa "goa.design/goa"
	securedservicec "goa.design/goa/examples/security/gen/http/secured_service/client"
	goahttp "goa.design/goa/http"
)

// UsageCommands returns the set of commands and sub-commands using the format
//
//    command (subcommand1|subcommand2|...)
//
func UsageCommands() string {
	return `secured-service (signin|secure|doubly-secure|also-doubly-secure)
`
}

// UsageExamples produces an example of a valid invocation of the CLI tool.
func UsageExamples() string {
	return os.Args[0] + ` secured-service signin --username "user" --password "password"` + "\n" +
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
		securedServiceFlags = flag.NewFlagSet("secured-service", flag.ContinueOnError)

		securedServiceSigninFlags        = flag.NewFlagSet("signin", flag.ExitOnError)
		securedServiceSigninUsernameFlag = securedServiceSigninFlags.String("username", "REQUIRED", "Username used to perform signin")
		securedServiceSigninPasswordFlag = securedServiceSigninFlags.String("password", "REQUIRED", "Password used to perform signin")

		securedServiceSecureFlags     = flag.NewFlagSet("secure", flag.ExitOnError)
		securedServiceSecureFailFlag  = securedServiceSecureFlags.String("fail", "", "")
		securedServiceSecureTokenFlag = securedServiceSecureFlags.String("token", "", "")

		securedServiceDoublySecureFlags     = flag.NewFlagSet("doubly-secure", flag.ExitOnError)
		securedServiceDoublySecureKeyFlag   = securedServiceDoublySecureFlags.String("key", "", "")
		securedServiceDoublySecureTokenFlag = securedServiceDoublySecureFlags.String("token", "", "")

		securedServiceAlsoDoublySecureFlags          = flag.NewFlagSet("also-doubly-secure", flag.ExitOnError)
		securedServiceAlsoDoublySecureKeyFlag        = securedServiceAlsoDoublySecureFlags.String("key", "", "")
		securedServiceAlsoDoublySecureOauthTokenFlag = securedServiceAlsoDoublySecureFlags.String("oauth-token", "", "")
		securedServiceAlsoDoublySecureTokenFlag      = securedServiceAlsoDoublySecureFlags.String("token", "", "")
		securedServiceAlsoDoublySecureUsernameFlag   = securedServiceAlsoDoublySecureFlags.String("username", "", "Username used to perform signin")
		securedServiceAlsoDoublySecurePasswordFlag   = securedServiceAlsoDoublySecureFlags.String("password", "", "Password used to perform signin")
	)
	securedServiceFlags.Usage = securedServiceUsage
	securedServiceSigninFlags.Usage = securedServiceSigninUsage
	securedServiceSecureFlags.Usage = securedServiceSecureUsage
	securedServiceDoublySecureFlags.Usage = securedServiceDoublySecureUsage
	securedServiceAlsoDoublySecureFlags.Usage = securedServiceAlsoDoublySecureUsage

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
		case "secured-service":
			svcf = securedServiceFlags
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
		case "secured-service":
			switch epn {
			case "signin":
				epf = securedServiceSigninFlags

			case "secure":
				epf = securedServiceSecureFlags

			case "doubly-secure":
				epf = securedServiceDoublySecureFlags

			case "also-doubly-secure":
				epf = securedServiceAlsoDoublySecureFlags

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
		case "secured-service":
			c := securedservicec.NewClient(scheme, host, doer, enc, dec, restore)
			switch epn {
			case "signin":
				endpoint = c.Signin()
				data, err = securedservicec.BuildSigninPayload(*securedServiceSigninUsernameFlag, *securedServiceSigninPasswordFlag)
			case "secure":
				endpoint = c.Secure()
				data, err = securedservicec.BuildSecurePayload(*securedServiceSecureFailFlag, *securedServiceSecureTokenFlag)
			case "doubly-secure":
				endpoint = c.DoublySecure()
				data, err = securedservicec.BuildDoublySecurePayload(*securedServiceDoublySecureKeyFlag, *securedServiceDoublySecureTokenFlag)
			case "also-doubly-secure":
				endpoint = c.AlsoDoublySecure()
				data, err = securedservicec.BuildAlsoDoublySecurePayload(*securedServiceAlsoDoublySecureKeyFlag, *securedServiceAlsoDoublySecureOauthTokenFlag, *securedServiceAlsoDoublySecureTokenFlag, *securedServiceAlsoDoublySecureUsernameFlag, *securedServiceAlsoDoublySecurePasswordFlag)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}

	return endpoint, data, nil
}

// secured-serviceUsage displays the usage of the secured-service command and
// its subcommands.
func securedServiceUsage() {
	fmt.Fprintf(os.Stderr, `The secured service exposes endpoints that require valid authorization credentials.
Usage:
    %s [globalflags] secured-service COMMAND [flags]

COMMAND:
    signin: Creates a valid JWT
    secure: This action is secured with the jwt scheme
    doubly-secure: This action is secured with the jwt scheme and also requires an API key query string.
    also-doubly-secure: This action is secured with the jwt scheme and also requires an API key header.

Additional help:
    %s secured-service COMMAND --help
`, os.Args[0], os.Args[0])
}
func securedServiceSigninUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] secured-service signin -username STRING -password STRING

Creates a valid JWT
    -username STRING: Username used to perform signin
    -password STRING: Password used to perform signin

Example:
    `+os.Args[0]+` secured-service signin --username "user" --password "password"
`, os.Args[0])
}

func securedServiceSecureUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] secured-service secure -fail BOOL -token STRING

This action is secured with the jwt scheme
    -fail BOOL: 
    -token STRING: 

Example:
    `+os.Args[0]+` secured-service secure --fail true --token "Ducimus non."
`, os.Args[0])
}

func securedServiceDoublySecureUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] secured-service doubly-secure -key STRING -token STRING

This action is secured with the jwt scheme and also requires an API key query string.
    -key STRING: 
    -token STRING: 

Example:
    `+os.Args[0]+` secured-service doubly-secure --key "abcdef12345" --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
`, os.Args[0])
}

func securedServiceAlsoDoublySecureUsage() {
	fmt.Fprintf(os.Stderr, `%s [flags] secured-service also-doubly-secure -key STRING -oauth-token STRING -token STRING -username STRING -password STRING

This action is secured with the jwt scheme and also requires an API key header.
    -key STRING: 
    -oauth-token STRING: 
    -token STRING: 
    -username STRING: Username used to perform signin
    -password STRING: Password used to perform signin

Example:
    `+os.Args[0]+` secured-service also-doubly-secure --key "abcdef12345" --oauth-token "Quos eveniet." --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ" --username "user" --password "password"
`, os.Args[0])
}
