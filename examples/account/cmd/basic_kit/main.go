package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"

	goa "goa.design/goa.v2"
	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/services"
	genhttp "goa.design/goa.v2/examples/account/gen/transport/http"
	"goa.design/goa.v2/rest"
	"goa.design/goa.v2/rest/middleware/debugging"
	"goa.design/goa.v2/rest/middleware/logging"
	"goa.design/goa.v2/rest/middleware/tracing"
)

// Logger wraps a go-kit logger and makes it compatible with goa's Logger
// interface.
type Logger struct {
	log.Logger
}

func (l *Logger) Info(_ context.Context, keyvals ...interface{}) {
	kv := append([]interface{}{"lvl", "info"}, keyvals...)
	l.Logger.Log(kv...)
}

func (l *Logger) Error(_ context.Context, keyvals ...interface{}) {
	kv := append([]interface{}{"lvl", "error"}, keyvals...)
	l.Logger.Log(kv...)
}

func main() {
	var (
		addr = flag.String("listen", ":8080", "HTTP listen `address`")
		dbg  = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	var ctx context.Context
	ctx = context.Background()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("listen", *addr).With("caller", log.DefaultCaller)

	var goalogger goa.Logger = &Logger{logger}

	var as services.Account
	as = basic.NewAccountService()

	var aep *endpoints.Account
	aep = endpoints.NewAccount(as)

	var (
		enc = genhttp.NewEncoder
		dec = genhttp.NewDecoder
	)

	createAccountHandler := httptransport.NewServer(
		ctx,
		endpoint.Endpoint(aep.Create),
		genhttp.CreateAccountDecodeRequestKit(dec),
		genhttp.CreateAccountEncodeResponseKit(enc),
		httptransport.ServerErrorEncoder(genhttp.CreateAccountEncoderErrorKit(enc, goalogger)),
	)

	var mux rest.ServeMux
	{
		mux = rest.NewMux()
		genhttp.MountCreateAccountHandler(mux, createAccountHandler)
	}

	var handler http.Handler = mux
	{
		handler = tracing.New()(handler)
		if *dbg {
			handler = debugging.New(goalogger)(handler)
		}
		handler = logging.New(goalogger)(handler)
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
