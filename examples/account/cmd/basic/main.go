package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/services"
	httptransport "goa.design/goa.v2/examples/account/gen/transport/http"
	"goa.design/goa.v2/rest"
	"goa.design/goa.v2/rest/middleware/debugging"
	"goa.design/goa.v2/rest/middleware/logging"
	"goa.design/goa.v2/rest/middleware/tracing"
)

func main() {
	var (
		addr = flag.String("listen", ":8080", "HTTP listen `address`")
		dbg  = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	var (
		logger *log.Logger
	)
	{
		logger = log.New(os.Stderr, "[basic] ", log.Ltime)
	}

	var (
		as services.Account
	)
	{
		as = basic.NewAccountService()
	}

	var (
		aep *endpoints.Account
	)
	{
		aep = endpoints.NewAccount(as)
	}

	var (
		enc = httptransport.NewEncoder
		dec = httptransport.NewDecoder
	)

	var (
		ah *httptransport.AccountHandlers
	)
	{
		ah = httptransport.NewAccountHandlers(aep, dec, enc, goa.AdaptStdLogger(logger))
	}

	var mux rest.ServeMux
	{
		mux = rest.NewMux()
		httptransport.MountAccountHandlers(mux, ah)
	}

	var handler http.Handler = mux
	{
		handler = tracing.New()(handler)
		if *dbg {
			handler = debugging.New(goa.AdaptStdLogger(logger))(handler)
		}
		handler = logging.New(goa.AdaptStdLogger(logger))(handler)
	}

	errc := make(chan error)

	// Setup interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Start HTTP listener
	go func() {
		logger.Printf("[INFO] listening on %s", *addr)
		errc <- http.ListenAndServe(*addr, handler)
	}()

	// Run!
	logger.Print("exit", <-errc)
}
