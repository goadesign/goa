// PLEASE NOTE: The code below is generated once as an example. It should be
// edited and handled like any other non-generated source files,

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/service"
	httptransport "goa.design/goa.v2/examples/account/gen/transport/http"
	"goa.design/goa.v2/rest"
	"goa.design/goa.v2/rest/middleware/debugging"
	"goa.design/goa.v2/rest/middleware/logging"
)

func main() {
	// Define command line flags, add any other flag required to configure
	// the service.
	var (
		addr = flag.String("listen", ":8080", "HTTP listen `address`")
		dbg  = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	// Setup logger and goa log adapter. Replace logger with your own using
	// your log package of choice. The goa.design/logging package defines
	// log adapters for common log packages. Writing adapters for other
	// log packages is very simple as well.
	var (
		logger  *log.Logger
		adapter goa.LogAdapter
	)
	{
		logger = log.New(os.Stderr, "[basic] ", log.Ltime)
		adapter = goa.AdaptStdLogger(logger)
	}

	// Create the structs that implement the services. These structs contain
	// the business logic specific to the service. The generated top-level
	// package contains dummy implementations as a way to get started
	// quickly. These files should be edited like this one and won't be
	// generated again.
	var (
		as service.Account
	)
	{
		as = basic.NewAccountService()
	}

	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
		aep *endpoints.Account
	)
	{
		aep = endpoints.NewAccount(as)
	}

	// Provide the transport specific request decoder and response encoder.
	// The goa rest package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions.
	//
	// The package goa.design/encoding provides implementations of these
	// functions for common encodings.
	var (
		dec = rest.DefaultRequestDecoder
		enc = rest.DefaultResponseEncoder
	)

	// Wrap the endints with the transport specific layers. The generated
	// http package contains code generated from the design which maps the
	// service input and output data structures to HTTP requests and
	// responses. The handlers data structure can be modified further for
	// example by taking advantage of its "Use" method to mount middlewares.
	// Middlewares can also be mounted on specific endpoints by overriding
	// the corresponding handlers struct field.
	var (
		ah *httptransport.AccountHandlers
	)
	{
		ah = httptransport.NewAccountHandlers(aep, dec, enc, adapter)
	}

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux rest.Muxer
	{
		mux = rest.NewMuxer()
		httptransport.MountAccountHandlers(mux, ah)
	}

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if *dbg {
			handler = debugging.New(adapter)(handler)
		}
		handler = logging.New(adapter)(handler)
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

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: *addr, Handler: mux}
	go func() {
		logger.Printf("[INFO] listening on %s", *addr)
		errc <- srv.ListenAndServe()
	}()

	// Wait for signal.
	logger.Print("exiting", <-errc)

	// Shutdown gracefully, but wait no longer than 5 seconds before
	// halting.
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)

	logger.Println("exited")
}
