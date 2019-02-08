package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"goa.design/goa"
	cli "goa.design/goa/examples/streaming/gen/http/cli/chatter"
	goahttp "goa.design/goa/http"
)

func doHTTP(scheme, host string, timeout int, debug bool) (goa.Endpoint, interface{}, error) {
	var (
		doer goahttp.Doer
	)
	{
		doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
			doer.(goahttp.DebugDoer).Fprint(os.Stderr)
		}
	}

	var (
		dialer       *websocket.Dialer
		connConfigFn goahttp.ConnConfigureFunc
	)
	{
		dialer = websocket.DefaultDialer
	}
	return cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
		dialer,
		connConfigFn,
	)
}
func httpUsageCommands() string {
	return cli.UsageCommands()
}

func httpUsageExamples() string {
	return cli.UsageExamples()
}
