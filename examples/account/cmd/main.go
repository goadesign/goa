package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/goadesign/goa/middleware"

	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/app"
	"goa.design/goa.v2/log"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	var ctx context.Context
	{
		ctx = context.Background()
	}

	var s app.AccountService
	{
		s = basic.NewAccountService()
		s = middleware.LogRequest(logger)(s)
	}

	var h http.Handler
	{
		h = app.MakeAccountHTTPHandler(ctx, s, log.NewContext(logger).With("component", "HTTP"))
	}

	errc := make(chan error)
	ctx := context.Background()

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger := log.NewContext(logger).With("transport", "HTTP")
		logger.Log("addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, server.NewHTTPMux())
	}()

	// Run!
	logger.Log("exit", <-errc)
}
