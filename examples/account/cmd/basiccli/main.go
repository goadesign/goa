package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"goa.design/goa.v2/examples/account/gen/cli"
	genhttp "goa.design/goa.v2/examples/account/gen/transport/http"
	"goa.design/goa.v2/rest"
)

func main() {
	var (
		addr    = flag.String("url", "http://localhost:8080", "`URL` to basic service host")
		debug   = flag.Bool("debug", false, "Print debug details")
		timeout = flag.Int("timeout", 30, "Maximum number of `seconds` to wait for response")
	)
	flag.Usage = usage
	flag.Parse()

	var (
		scheme string
		host   string
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
	}

	var (
		doer rest.Doer
	)
	{
		doer = http.DefaultClient
		if *debug {
			doer = rest.NewDebugDoer(doer)
		}
	}

	client := genhttp.NewAccountClient(
		scheme,
		host,
		doer,
		rest.DefaultRequestEncoder,
		rest.DefaultResponseDecoder,
	)

	data, err := cli.RunCommand(*timeout, client)

	if dd, ok := doer.(rest.DebugDoer); ok {
		dd.Fprint(os.Stderr)
	}

	if err != nil && err != flag.ErrHelp {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `basiccli is a command line client for the basic service.

Usage:
    basiccli [-url URL][-timeout TIMEOUT][-debug] SERVICE ENDPOINT [flags]

-url URL: specify service URL (http://localhost:8080)
-timeout: Maximum number of seconds to wait for response (30)
-debug:   print debug details (false)

Commands:
%s

Example:
%s
`, indent(cli.UsageCommands()), indent(cli.UsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
