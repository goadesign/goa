package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	chattersvc "goa.design/goa/examples/streaming/gen/chatter"
	chattersvcsvr "goa.design/goa/examples/streaming/gen/http/chatter/server"
	goahttp "goa.design/goa/http"
	httpmdlwr "goa.design/goa/http/middleware"
	"goa.design/goa/middleware"
)

// handleHTTPServer starts configures and starts a HTTP server on the given
// URL. It shuts down the server if any error is received in the error channel.
func handleHTTPServer(ctx context.Context, u *url.URL, chatterEndpoints *chattersvc.Endpoints, wg *sync.WaitGroup, errc chan error, logger *log.Logger, debug bool) {

	// Setup goa log adapter.
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
		chatterServer *chattersvcsvr.Server
	)
	{
		eh := errorHandler(logger)
		upgrader := &websocket.Upgrader{}
		chatterConfigurer := chattersvcsvr.NewConnConfigurer(nil)
		chatterConfigurer.SubscribeFn = pingPonger(logger)
		chatterServer = chattersvcsvr.New(chatterEndpoints, mux, dec, enc, eh, upgrader, chatterConfigurer)
	}
	// Configure the mux.
	chattersvcsvr.Mount(mux, chatterServer)

	// Wrap the multiplexer with additional middlewares. Middlewares mounted
	// here apply to all the service endpoints.
	var handler http.Handler = mux
	{
		if debug {
			handler = httpmdlwr.Debug(mux, os.Stdout)(handler)
		}
		handler = httpmdlwr.Log(adapter)(handler)
		handler = httpmdlwr.RequestID()(handler)
	}

	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler}
	for _, m := range chatterServer.Mounts {
		logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		// Start HTTP server in a separate goroutine.
		go func() {
			logger.Printf("HTTP server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		select {
		case <-ctx.Done():
			logger.Printf("shutting down HTTP server at %q", u.Host)

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
func errorHandler(logger *log.Logger) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		id := ctx.Value(middleware.RequestIDKey).(string)
		w.Write([]byte("[" + id + "] encoding: " + err.Error()))
		logger.Printf("[%s] ERROR: %s", id, err.Error())
	}
}

// pingPonger configures the websocket connection to check the health of the
// connection between client and server. It periodically sends a ping message
// to the client and if the client does not respond with a pong within a
// specified time, it closes the websocket connection and cancels the request
// context.
//
// NOTE: This is suitable for use only in server-side streaming endpoints
// (i.e. client does NOT send any messages through the stream), because it
// reads the websocket connection for pong messages from the client. If this is
// used in any endpoint where the client streams, it will result in lost
// messages from the client which is undesirable.
func pingPonger(logger *log.Logger) goahttp.ConnConfigureFunc {
	pingInterval := 3 * time.Second
	return goahttp.ConnConfigureFunc(func(conn *websocket.Conn, cancel context.CancelFunc) *websocket.Conn {
		// errc is the channel read by ping-ponger to check if there were any
		// errors when reading messages sent by the client from the websocket.
		errc := make(chan error)

		// Start a goroutine to read messages sent by the client from the
		// websocket connection. This will pick up any pong message sent
		// by the client. Send any errors to errc.
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					logger.Printf("error reading messages from client: %v", err)
					errc <- err
					return
				}
			}
		}()

		// Start the pinger in a separate goroutine. Read any errors in the
		// error channel and stop the goroutine when error received. Close the
		// websocket connection and cancel the request when client when error
		// received.
		go func() {
			ticker := time.NewTicker(pingInterval)
			defer func() {
				ticker.Stop()
				logger.Printf("client did not respond with pong")
				// cancel the request context when timer expires
				cancel()
			}()

			// Set a read deadline to read pong messages from the client.
			// If a client fails to send a pong before the deadline any
			// further connection reads will result in an error. We exit the
			// goroutine if connection reads error out.
			conn.SetReadDeadline(time.Now().Add(pingInterval * 2))

			// set a custom pong handler
			pongFn := conn.PongHandler()
			conn.SetPongHandler(func(appData string) error {
				logger.Printf("client says pong")
				// Reset the read deadline
				conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
				return pongFn(appData)
			})

			for {
				select {
				case <-errc:
					return
				case <-ticker.C:
					// send periodic ping message
					if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(time.Second)); err != nil {
						return
					}
					logger.Printf("pinged client")
				}
			}
		}()
		return conn
	})
}
