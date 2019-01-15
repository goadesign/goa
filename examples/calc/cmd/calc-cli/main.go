package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"

	"goa.design/goa"
)

func main() {
	var (
		hostF = flag.String("host", "development", "Server host (valid values: development, production)")
		addrF = flag.String("url", "", "URL to service host")

		versionF = flag.String("version", "v1", "API version")
		verboseF = flag.Bool("verbose", false, "Print request and response details")
		vF       = flag.Bool("v", false, "Print request and response details")
		timeoutF = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
	var (
		addr    string
		timeout int
		debug   bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
			case "development":
				addr = "http://localhost:8000/calc"
			case "production":
				addr = "https://{version}.goa.design/calc"
				addr = strings.Replace(addr, "{version}", *versionF, -1)
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: development|production)", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host   string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}
	var (
		endpoint goa.Endpoint
		payload  interface{}
		err      error
	)
	{
		switch scheme {
		case "http", "https":
			endpoint, payload, err = doHTTP(scheme, host, timeout, debug)
		case "grpc", "grpcs":
			endpoint, payload, err = doGRPC(scheme, host, timeout, debug)
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: grpc|grpcs|http|https)", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)
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
	fmt.Fprintf(os.Stderr, `%s is a command line client for the calc API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v][-version VERSION] SERVICE ENDPOINT [flags]

    -host HOST:  server host (development). valid values: development, production
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)
    -version:    API version (v1)

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
