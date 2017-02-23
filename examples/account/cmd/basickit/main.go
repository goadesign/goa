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
	genhttp "goa.design/goa.v2/examples/account/gen/transport/http"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/kit"
	"goa.design/goa.v2/examples/account/gen/services"
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
		ctx context.Context
	)
	{
		ctx = context.Background()
	}

	var (
		logger log.Logger
	)
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("listen", *addr).With("caller", log.DefaultCaller)
	}

	var (
		gl goa.Logger
	)
	{
		gl = &Logger{logger}
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
		dec rest.RequestDecoderFunc
		enc rest.ResponseEncoderFunc
	)
	{
		dec = genhttp.NewRequestDecoder
		enc = genhttp.NewResponseEncoder
	}

	var (
		createAccountHandler http.Handler
		listAccountHandler   http.Handler
		showAccountHandler   http.Handler
		deleteAccountHandler http.Handler
	)
	{
		createAccountHandler = httptransport.NewServer(
			ctx,
			endpoint.Endpoint(create(aep)),
			kit.CreateAccountDecodeRequest(dec),
			kit.CreateAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.CreateAccountEncodeError(enc, gl)),
		)
		listAccountHandler = httptransport.NewServer(
			ctx,
			endpoint.Endpoint(list(aep)),
			kit.ListAccountDecodeRequest(dec),
			kit.ListAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.ListAccountEncodeError(enc, gl)),
		)
		showAccountHandler = httptransport.NewServer(
			ctx,
			endpoint.Endpoint(show(aep)),
			kit.ShowAccountDecodeRequest(dec),
			kit.ShowAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.ShowAccountEncodeError(enc, gl)),
		)
		deleteAccountHandler = httptransport.NewServer(
			ctx,
			endpoint.Endpoint(delete_(aep)),
			kit.DeleteAccountDecodeRequest(dec),
			kit.DeleteAccountEncodeResponse(enc),
			httptransport.ServerBefore(kit.StashRequest),
			httptransport.ServerErrorEncoder(kit.DeleteAccountEncodeError(enc, gl)),
		)
	}

	var (
		mux rest.ServeMux
	)
	{
		mux = rest.NewMux()
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
