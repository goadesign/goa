package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"

	"goa.design/goa.v2/examples/account"
	"goa.design/goa.v2/examples/account/gen/endpoints"
	"goa.design/goa.v2/examples/account/gen/services"
	"goa.design/goa.v2/examples/account/gen/transport"
	"goa.design/goa.v2/rest"
	"goa.design/goa.v2/rest/middleware/debug"
	"goa.design/goa.v2/rest/middleware/logging"
)

func main() {
	var (
		listen = flag.String("listen", ":8080", "HTTP listen `address`")
	)
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.NewContext(logger).With("listen", *listen).With("caller", log.DefaultCaller)

	as := basic.NewAccountService()

	aep := endpoints.NewAccount(as)

	encode := transport.NewHTTPEncoder
	decode := transport.NewHTTPDecoder
	eencode := transport.NewErrorHTTPEncoder

	ah := transport.NewAccountHTTPHandlers(ctx, aep, decode, encode, eencode)

	createAccountHandler := httptransport.NewServer(
		ctx,
		aep.Create,
		genkit.CreateAccountDecoderFunc(decoder rest.DecoderFunc),
		encodeResponse,
	)

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}
