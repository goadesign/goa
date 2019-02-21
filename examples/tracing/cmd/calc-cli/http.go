package main

import (
	"net/http"
	"time"

	"goa.design/goa"
	cli "goa.design/goa/examples/basic/gen/http/cli/calc"
	goahttp "goa.design/goa/http"
	"goa.design/goa/http/middleware"
	"goa.design/goa/http/middleware/xray"
)

func doHTTP(scheme, host string, timeout int, debug bool) (goa.Endpoint, interface{}, error) {
	var (
		doer goahttp.Doer
	)
	{
		doer = &http.Client{Timeout: time.Duration(timeout) * time.Second}
		if debug {
			doer = goahttp.NewDebugDoer(doer)
		}
		// Wrap doer with X-Ray and trace client middleware. Order is very important.
		doer = xray.WrapDoer(doer)
		doer = middleware.WrapDoer(doer)
	}

	return cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
	)
}
func httpUsageCommands() string {
	return cli.UsageCommands()
}

func httpUsageExamples() string {
	return cli.UsageExamples()
}
