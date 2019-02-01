package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"goa.design/goa"
	chattersvc "goa.design/goa/examples/streaming/gen/chatter"
)

func main() {
	var (
		hostF = flag.String("host", "localhost", "Server host (valid values: localhost)")
		addrF = flag.String("url", "", "URL to service host")

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
			case "localhost":
				addr = "http://localhost:80"
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: localhost)", *hostF)
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
		case "grpc":
			endpoint, payload, err = doGRPC(scheme, host, timeout, debug)
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: grpc|http)", scheme)
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
		switch stream := data.(type) {
		case chattersvc.EchoerClientStream:
			// bidirectional streaming
			trapCtrlC(stream)
			fmt.Println("Press Ctrl+D to stop chatting.")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				p := scanner.Text()
				if err := stream.Send(p); err != nil {
					fmt.Println(fmt.Errorf("Error sending into stream: %s", err))
					os.Exit(1)
				}
				d, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println(fmt.Errorf("Error reading from stream: %s", err))
				}
				prettyPrint(d)
			}
			if err := stream.Close(); err != nil {
				fmt.Println(fmt.Errorf("Error closing stream: %s", err))
			}
		case chattersvc.ListenerClientStream:
			// payload streaming (no server response)
			trapCtrlC(stream)
			fmt.Println("Press Ctrl+D to stop chatting.")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				p := scanner.Text()
				if err := stream.Send(p); err != nil {
					fmt.Println(fmt.Errorf("Error sending into stream: %s", err))
					os.Exit(1)
				}
			}
			if err := stream.Close(); err != nil {
				fmt.Println(fmt.Errorf("Error closing stream: %s", err))
			}
		case chattersvc.SummaryClientStream:
			// payload streaming (server responds with a result type)
			trapCtrlC(stream)
			fmt.Println("Press Ctrl+D to stop chatting.")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				p := scanner.Text()
				if err := stream.Send(p); err != nil {
					fmt.Println(fmt.Errorf("Error sending into stream: %s", err))
					os.Exit(1)
				}
			}
			if p, err := stream.CloseAndRecv(); err != nil {
				fmt.Println(fmt.Errorf("Error closing stream: %s", err))
			} else {
				prettyPrint(p)
			}
		case chattersvc.HistoryClientStream:
			// result streaming
			for {
				p, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println(fmt.Errorf("Error reading from stream: %s", err))
				}
				prettyPrint(p)
			}
		default:
			prettyPrint(data)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s is a command line client for the chatter API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -host HOST:  server host (localhost). valid values: localhost
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

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

func prettyPrint(s interface{}) {
	m, _ := json.MarshalIndent(s, "", "    ")
	fmt.Println(string(m))
}

// Trap Ctrl+C to gracefully exit the client.
func trapCtrlC(stream interface{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func(stream interface{}) {
		for range ch {
			fmt.Println("\nexiting")
			if s, ok := stream.(chattersvc.EchoerClientStream); ok {
				s.Close()
			} else if s, ok := stream.(chattersvc.ListenerClientStream); ok {
				s.Close()
			} else if s, ok := stream.(chattersvc.SummaryClientStream); ok {
				s.CloseAndRecv()
			}
			os.Exit(0)
		}
	}(stream)
}
