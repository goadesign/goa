package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	carssvc "goa.design/goa/examples/streaming/gen/cars"
	"goa.design/goa/examples/streaming/gen/http/cli"
	goahttp "goa.design/goa/http"
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

	var (
		dialer *websocket.Dialer
	)
	{
		dialer = websocket.DefaultDialer
	}

	endpoint, payload, err := cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
		dialer,
		nil,
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
		switch s := data.(type) {
		case carssvc.ListClientStream:
			for {
				d, err := s.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println(fmt.Errorf("Error reading from stream: %s", err))
				}
				m, _ := json.MarshalIndent(d, "", "    ")
				fmt.Println(string(m))
			}
		case carssvc.AddClientStream:
			trapCtrlC()
			fmt.Println("Press Ctrl+D to stop sending payload.")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				t := scanner.Text()
				p, err := buildAddStreamingPayload(t)
				if err != nil {
					fmt.Println(fmt.Errorf("Error creating payload: %s", err))
					os.Exit(1)
				}
				if err := s.Send(p); err != nil {
					fmt.Println(fmt.Errorf("Error sending into stream: %s", err))
					os.Exit(1)
				}
			}
			d, err := s.CloseAndRecv()
			if err == io.EOF {
				os.Exit(0)
			}
			if err != nil {
				fmt.Println(fmt.Errorf("Error reading from stream: %s", err))
			}
			m, _ := json.MarshalIndent(d, "", "    ")
			fmt.Println(string(m))
		case carssvc.UpdateClientStream:
			trapCtrlC()
			fmt.Println("Press Ctrl+D to stop.")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				t := scanner.Text()
				p, err := buildUpdateStreamingPayload(t)
				if err != nil {
					fmt.Println(fmt.Errorf("Error creating payload: %s", err))
					os.Exit(1)
				}
				if err := s.Send(p); err != nil {
					fmt.Println(fmt.Errorf("Error sending into stream: %s", err))
					os.Exit(1)
				}
				var cars carssvc.StoredCarCollection
				cars, err = s.Recv()
				if err == io.EOF {
					s.Close()
					os.Exit(0)
				}
				if err != nil {
					fmt.Println(fmt.Errorf("Error reading from stream: %s", err))
				}
				m, _ := json.MarshalIndent(cars, "", "    ")
				fmt.Println(string(m))
			}
		default:
			m, _ := json.MarshalIndent(data, "", "    ")
			fmt.Println(string(m))
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `%s is a command line client for the cars API.

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

// Graceful shutdown
func trapCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for range ch {
			fmt.Println("\nexiting")
			os.Exit(0)
		}
	}()
}

// buildAddStreamingPayload builds the streaming payload for the cars add
// endpoint.
func buildAddStreamingPayload(carsAddBody string) (*carssvc.AddStreamingPayload, error) {
	var err error
	var car carssvc.AddStreamingPayload
	{
		err = json.Unmarshal([]byte(carsAddBody), &car)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body, example of valid JSON:\n%s", "'{\n      \"car\": {\n         \"body_style\": \"Laudantium qui minima voluptatibus in incidunt.\",\n         \"make\": \"Aspernatur totam.\",\n         \"model\": \"Vero odio odio id autem.\"\n      }\n   }'")
		}
	}
	return &car, nil
}

// buildUpdateStreamingPayload builds the streaming payload for the cars update
// endpoint.
func buildUpdateStreamingPayload(carsUpdateBody string) ([]*carssvc.Car, error) {
	var err error
	var cars []*carssvc.Car
	{
		err = json.Unmarshal([]byte(carsUpdateBody), &cars)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON for body")
		}
	}
	return cars, nil
}
