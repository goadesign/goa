package main

import (
	"net/http"
	"time"

	"goa.design/goa"
	cellar "goa.design/goa/examples/cellar"
	cli "goa.design/goa/examples/cellar/gen/http/cli/cellar"
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
		}
	}

	return cli.ParseEndpoint(
		scheme,
		host,
		doer,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		debug,
		cellar.StorageMultiAddEncoderFunc,
		cellar.StorageMultiUpdateEncoderFunc,
	)
}

func httpUsageCommands() string {
	return cli.UsageCommands()
}

func httpUsageExamples() string {
	return cli.UsageExamples()
}
