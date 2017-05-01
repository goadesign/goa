package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"goa.design/goa.v2/examples/account/gen/service"
	genhttp "goa.design/goa.v2/examples/account/gen/transport/http"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/kit"
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
		logger log.Logger
	)
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "listen", *addr)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var (
		gl goa.LogAdapter
	)
	{
		gl = &Logger{logger}
	}

	var (
		as service.Account
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

	var (
		createAccountHandler http.Handler
		listAccountHandler   http.Handler
		showAccountHandler   http.Handler
		deleteAccountHandler http.Handler
	)
	{
		createAccountHandler = httptransport.NewServer(
			endpoint.Endpoint(create(aep)),
			kit.CreateAccountDecodeRequest(dec),
			kit.CreateAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.CreateAccountEncodeError(enc, gl)),
		)
		listAccountHandler = httptransport.NewServer(
			endpoint.Endpoint(list(aep)),
			kit.ListAccountDecodeRequest(dec),
			kit.ListAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.ListAccountEncodeError(enc, gl)),
		)
		showAccountHandler = httptransport.NewServer(
			endpoint.Endpoint(show(aep)),
			kit.ShowAccountDecodeRequest(dec),
			kit.ShowAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.ShowAccountEncodeError(enc, gl)),
		)
		deleteAccountHandler = httptransport.NewServer(
			endpoint.Endpoint(delete_(aep)),
			kit.DeleteAccountDecodeRequest(dec),
			kit.DeleteAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.DeleteAccountEncodeError(enc, gl)),
		)
	}

	var (
		mux rest.Muxer
	)
	{
		mux = rest.NewMuxer()
		genhttp.MountCreateAccountHandler(mux, createAccountHandler)
		genhttp.MountListAccountHandler(mux, listAccountHandler)
		genhttp.MountShowAccountHandler(mux, showAccountHandler)
		genhttp.MountDeleteAccountHandler(mux, deleteAccountHandler)
	}

	var (
		handler http.Handler
	)
	{
		handler = tracing.New()(mux)
		if *dbg {
			handler = debugging.New(gl)(handler)
		}
		handler = logging.New(gl)(handler)
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *addr)
		errs <- http.ListenAndServe(*addr, handler)
	}()

	logger.Log("exit", <-errs)
}

// Temporary until go-kit moves to the stdlib context.

func create(aep *endpoints.Account) func(ctx context.Context, req interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return aep.Create(ctx, req)
	}
}

func list(aep *endpoints.Account) func(ctx context.Context, req interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return aep.List(ctx, req)
	}
}

func show(aep *endpoints.Account) func(ctx context.Context, req interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return aep.Show(ctx, req)
	}
}

func delete_(aep *endpoints.Account) func(ctx context.Context, req interface{}) (interface{}, error) {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return aep.Delete(ctx, req)
	}
}
