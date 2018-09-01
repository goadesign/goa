package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	divider "goa.design/goa/examples/error"
	dividersvc "goa.design/goa/examples/error/gen/divider"
)

func main() {
	// Define command line flags, add any other flag required to configure
	// the service.
	var (
		httpAddrF = flag.String("http-listen", ":8080", "HTTP listen `address`")
		dbgF      = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	// Setup logger and goa log adapter. Replace logger with your own using
	// your log package of choice. The goa.design/middleware/logging/...
	// packages define log adapters for common log packages.
	var (
		logger *log.Logger
	)
	{
		logger = log.New(os.Stderr, "[divider] ", log.Ltime)
	}

	// Create the structs that implement the services.
	var (
		dividerSvc dividersvc.Service
	)
	{
		dividerSvc = divider.NewDivider(logger)
	}

	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
		dividerEndpoints *dividersvc.Endpoints
	)
	{
		dividerEndpoints = dividersvc.NewEndpoints(dividerSvc)
	}

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the service to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		errc <- fmt.Errorf("%s", <-c)
	}()
	httpSrvr := httpServe(*httpAddrF, dividerEndpoints, errc, logger, *dbgF)

	// Wait for signal.
	logger.Printf("exiting (%v)", <-errc)
	logger.Println("Shutting down HTTP server at " + *httpAddrF)
	httpStop(httpSrvr)
	logger.Println("exited")
}
