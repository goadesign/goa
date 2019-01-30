package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	chattersvc "goa.design/goa/examples/chatter/gen/chatter"
	chattersvcsvr "goa.design/goa/examples/chatter/gen/http/chatter/server"
	goahttp "goa.design/goa/http"
	httpmiddleware "goa.design/goa/http/middleware"
	"goa.design/goa/middleware"
)

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(ctx context.Context, u *url.URL, chatterEndpoints *chattersvc.Endpoints, wg *sync.WaitGroup, errc chan error, logger middleware.Logger, debug bool) {

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
		chatterServer *chattersvcsvr.Server
	)
	{
		eh := errorHandler(logger)
		upgrader := &websocket.Upgrader{}
		chatterServer = chattersvcsvr.New(chatterEndpoints, mux, dec, enc, eh, upgrader, nil)
	}
	// Configure the mux.
	chattersvcsvr.Mount(mux, chatterServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if debug {
			handler = httpmiddleware.Debug(mux, os.Stdout)(handler)
		}
		handler = httpmiddleware.Log(logger)(handler)
		handler = httpmiddleware.RequestID()(handler)
	}

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler}
	for _, m := range chatterServer.Mounts {
		logger.Log("msg", "serving HTTP", "mount", m.Method, "verb", m.Verb, "path", m.Pattern)
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start HTTP server in a separate goroutine.
		go func() {
			logger.Log("msg", "HTTP server listening", "host", u.Host)
			errc <- srv.ListenAndServe()
		}()

		select {
		case <-ctx.Done():
			logger.Log("msg", "shutting down HTTP server", "host", u.Host)

			// Shutdown gracefully with a 30s timeout.
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			srv.Shutdown(ctx)
			return
		}
	}()
}

// errorHandler returns a function that writes and logs the given error.
// The function also writes and logs the error unique ID so that it's possible
// to correlate.
func errorHandler(logger middleware.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Log("msg", "error handler", "id", id, "err", err.Error())
	}
}
