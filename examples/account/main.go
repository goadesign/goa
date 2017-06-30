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
	"goa.design/goa.v2/examples/account/gen/account"
	"goa.design/goa.v2/examples/account/gen/account/transport"
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
		logger = log.New(os.Stderr, "[account] ", log.Ltime)
		adapter = goa.AdaptStdLogger(logger)
	}

	// Create the structs that implement the services. These structs contain
	// the business logic specific to the service. The generated top-level
	// package contains dummy implementations as a way to get started
	// quickly. These files should be edited like this one and won't be
	// generated again.
	var (
		as account.Service
	)
	{
		as = NewAccountService()
	}

	// Wrap the services in endpoints that can be invoked from other
	// services potentially running in different processes.
	var (
		aep *account.Endpoints
	)
	{
		aep = account.NewEndpoints(as)
	}

	// Provide the transport specific request decoder and response encoder.
	// The goa rest package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions.
	var (
		dec = rest.RequestDecoder
		enc = rest.ResponseEncoder
	)

	// Build the service HTTP request multiplexer
	var mux rest.Muxer
	{
		mux = rest.NewMuxer()
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// http package contains code generated from the design which maps the
	// service input and output data structures to HTTP requests and
	// responses.
	var (
		ah *transport.Handlers
	)
	{
		ah = transport.NewHandlers(aep, mux, dec, enc)
	}

	// Configure the mux to serve HTTP requests to the service endpoints.
	transport.MountHandlers(mux, ah)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if *dbg {
			handler = debugging.New(mux, adapter)(handler)
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
	srv := &http.Server{Addr: *addr, Handler: handler}
	go func() {
		logger.Printf("[INFO] listening on %s", *addr)
		errc <- srv.ListenAndServe()
	}()

	// Wait for signal.
	logger.Printf("exiting: %s", <-errc)

	// Shutdown gracefully, but wait no longer than 5 seconds before
	// halting.
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)

	logger.Println("exited")
}
