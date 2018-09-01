package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	dividersvc "goa.design/goa/examples/error/gen/divider"
	dividersvcsvr "goa.design/goa/examples/error/gen/http/divider/server"
	goahttp "goa.design/goa/http"
	"goa.design/goa/http/middleware"
)

func httpServe(addr string, dividerEndpoints *dividersvc.Endpoints, errc chan error, logger *log.Logger, debug bool) *http.Server {
	// Setup goa log adapter. Replace logger with your own using your
	// log package of choice. The goa.design/middleware/logging/...
	// ackages define log adapters for common log packages.
	var (
		adapter middleware.Logger
	)
	{
		adapter = middleware.NewLogger(logger)
	}

	// Provide the transport specific request decoder and response encoder.
	// The goa http package has built-in support for JSON, XML and gob.
	// Other encodings can be used by providing the corresponding functions,
	// see goa.design/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)

	// Build the service HTTP request multiplexer and configure it to serve
	// HTTP requests to the service endpoints.
	var mux goahttp.Muxer
	{
		mux = goahttp.NewMuxer()
	}

	// Wrap the endpoints with the transport specific layers. The generated
	// server packages contains code generated from the design which maps
	// the service input and output data structures to HTTP requests and
	// responses.
	var (
		dividerServer *dividersvcsvr.Server
	)
	{
		eh := errorHandler(logger)
		dividerServer = dividersvcsvr.New(dividerEndpoints, mux, dec, enc, eh)
	}

	// Configure the mux.
	dividersvcsvr.Mount(mux, dividerServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if debug {
			handler = middleware.Debug(mux, os.Stdout)(handler)
		}
		handler = middleware.Log(adapter)(handler)
		handler = middleware.RequestID()(handler)
	}

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: addr, Handler: handler}
	go func() {
		for _, m := range dividerServer.Mounts {
			logger.Printf("method %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
		}
		logger.Printf("HTTP listening on %s", addr)
		errc <- srv.ListenAndServe()
	}()

	return srv
}

func httpStop(srv *http.Server) {
	// Shutdown gracefully with a 30s timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger *log.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}
