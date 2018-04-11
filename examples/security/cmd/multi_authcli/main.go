package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	goahttp "goa.design/goa/http"
	"goa.design/plugins/security/examples/multi_auth/gen/http/cli"
)

func main() {
	var (
		addr    = flag.String("url", "http://localhost:8080", "`URL` to service host")
		verbose = flag.Bool("verbose", false, "Print request and response details")
		v       = flag.Bool("v", false, "Print request and response details")
		timeout = flag.Int("timeout", 30, "Maximum number of `seconds` to wait for response")
	)
	flag.Usage = usage
	flag.Parse()

	var (
		scheme string
		host   string
		debug  bool
	)
	{
		u, err := url.Parse(*addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", *addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
		if scheme == "" {
			scheme = "http"
		}
		debug = *verbose || *v
	}

	var (
		doer goahttp.Doer
	)
	{
		doer = &http.Client{Timeout: time.Duration(*timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
	}

	endpoint, payload, err := cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
	)
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)

	if debug {
		doer.(goahttp.DebugDoer).Fprint(os.Stderr)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil && !debug {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s is a command line client for the multi_auth API.

Usage:
    %s [-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -url URL:    specify service URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
`, os.Args[0], os.Args[0], indent(cli.UsageCommands()), os.Args[0], indent(cli.UsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
