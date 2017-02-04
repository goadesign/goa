package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/app"
	"goa.design/goa.v2/rest"
	"goa.design/goa.v2/rest/middleware/debug"
	"goa.design/goa.v2/rest/middleware/logging"
	"goa.design/goa.v2/rest/middleware/tracing"
)

func main() {
	var (
		addr = flag.String("http.addr", ":8080", "HTTP listen `address`")
		dbg  = flag.Bool("app.debug", false, "Log request and response bodies")
	)
	flag.Parse()

	var logger *log.Logger
	{
		logger = log.New(os.Stderr, "[basic] ", log.Ltime)
	}

	var ctx context.Context
	{
		ctx = context.Background()
	}

	var (
		as app.AccountService
	)
	{
		as = basic.NewAccountService()
	}

	var (
		aep *app.AccountEndpoints
	)
	{
		aep = app.NewAccountEndpoints(as)
	}

	var (
		encode = rest.EncodeErrors(app.NewEncoder)
		decode = app.NewDecoder
	)

	var (
		ah *app.AccountHTTPHandlers
	)
	{
		ah = app.NewAccountHTTPHandlers(ctx, aep, decode, encode)
	}

	var mux rest.ServeMux
	{
		mux = rest.NewMux()
		app.MountAccountHTTPHandlers(mux, ah)
	}

	var handler http.Handler = mux
	{
		handler = tracing.New()(handler)
		if *dbg {
			handler = debug.NewStd(logger)(handler)
		}
		handler = logging.NewStd(logger)(handler)
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
		logger.Printf("listening on %s", *addr)
		errc <- http.ListenAndServe(*addr, handler)
	}()

	// Run!
	logger.Print("exit", <-errc)
}
