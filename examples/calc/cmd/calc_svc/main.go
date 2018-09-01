package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	calc "goa.design/goa/examples/calc"
	calcsvc "goa.design/goa/examples/calc/gen/calc"
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
		logger = log.New(os.Stderr, "[calc] ", log.Ltime)
	}

	// Create the structs that implement the services.
	var (
		calcSvc calcsvc.Service
	)
	{
		calcSvc = calc.NewCalc(logger)
	}

	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
		calcEndpoints *calcsvc.Endpoints
	)
	{
		calcEndpoints = calcsvc.NewEndpoints(calcSvc)
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
	httpSrvr := httpServe(*httpAddrF, calcEndpoints, errc, logger, *dbgF)

	// Wait for signal.
	logger.Printf("exiting (%v)", <-errc)
	logger.Println("Shutting down HTTP server at " + *httpAddrF)
	httpStop(httpSrvr)
	logger.Println("exited")
}
