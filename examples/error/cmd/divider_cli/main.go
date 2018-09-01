package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	var (
		addrF      = flag.String("url", "http://localhost:8080", "`URL` to service host")
		verboseF   = flag.Bool("verbose", false, "Print request and response details")
		vF         = flag.Bool("v", false, "Print request and response details")
		timeoutF   = flag.Int("timeout", 30, "Maximum number of `seconds` to wait for response")
		transportF = flag.String("transport", "http", "Transport to use for the request (Allowed values: grpc, http)")
	)
	flag.Usage = usage
	flag.Parse()

	var (
		transport string
		timeout   int
		debug     bool
	)
	{
		transport = *transportF
		timeout = *timeoutF
		debug = *verboseF || *vF
	}
	switch transport {
	case "http":
		httpDo(*addrF, timeout, debug)
	default:
		fmt.Fprintf(os.Stderr, "unknown transport %q", transport)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s is a command line client for the divider API.

Usage:
    %s [-url URL][-timeout SECONDS][-verbose|-v][-transport NAME] SERVICE ENDPOINT [flags]

    -url URL:    specify service URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)
    -transport:  specify which transport to use (allowed values: grpc, http. Default is http.)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
`, os.Args[0], os.Args[0], indent(httpUsageCommands()), os.Args[0], indent(httpUsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
