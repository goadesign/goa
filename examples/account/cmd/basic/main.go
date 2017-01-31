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

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/app"
	"goa.design/goa.v2/rest"
	"goa.design/goa.v2/rest/middleware/logging"
	"goa.design/goa.v2/rest/middleware/recovering"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
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
		encode      = app.NewEncoder
		decode      = app.NewDecoder
		encodeError = rest.NewErrorEncoder(encode)
		m           = logging.New(goa.AdaptLogger(logger), true)
	)

	var (
		ah *app.AccountHTTPHandlers
	)
	{
		ah = app.NewAccountHTTPHandlers(ctx, aep, decode, encode, encodeError, m)
	}

	var mux rest.ServeMux
	{
		mux = rest.NewMux()
		app.MountAccountHTTPHandlers(mux, ah)
	}

	var handler http.Handler
	{
		handler = recovering.New(encodeError)(mux)
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
		logger.Printf("listening on %s", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, handler)
	}()

	// Run!
	logger.Print("exit", <-errc)
}
