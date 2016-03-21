// +build !appengine

package goa

import (
	"net"
	"net/http"
	"sync"
	"time"

	"gopkg.in/tylerb/graceful.v1"
)

// GracefulServer is a goa application using a graceful shutdown server.
// When sending any of the signals listed in InterruptSignals to the process GracefulServer:
//
// * disables keepalive connections.
//
// * closes the listening socket, allowing another process to listen on that port immediately.
//
// * calls Cancel, signaling all active handlers.
type GracefulServer struct {
	*Service
	sync.Mutex
	server  *graceful.Server
	timeout time.Duration

	// CancelOnShutdown tells whether existing requests should be canceled when shutdown is
	// triggered (true) or whether to wait until the requests complete (false).
	CancelOnShutdown bool
}

// NewGraceful returns a goa application that uses a graceful shutdown server.
func NewGraceful(service *Service, cancelOnShutdown bool, timeout time.Duration) *GracefulServer {
	return &GracefulServer{Service: service, CancelOnShutdown: cancelOnShutdown, timeout: timeout}
}

// ListenAndServe starts the HTTP server and sets up a listener on the given host/port.
func (serv *GracefulServer) ListenAndServe(addr string) error {
	serv.setup(addr)
	serv.Info("started", "transport", "http", "addr", addr)
	if err := serv.server.ListenAndServe(); err != nil {
		// there may be a final "accept" error after completion of graceful shutdown
		// which can be safely ignored here.
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			return err
		}
	}
	return nil
}

// ListenAndServeTLS starts a HTTPS server and sets up a listener on the given host/port.
func (serv *GracefulServer) ListenAndServeTLS(addr, certFile, keyFile string) error {
	serv.setup(addr)
	serv.Info("started", "transport", "https", "addr", addr)
	return serv.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown initiates graceful shutdown of the running server once. Returns true on
// initial shutdown and false if already shutting down.
func (serv *GracefulServer) Shutdown() bool {
	IncrCounter([]string{"goa", "graceful", "restart"}, 1.0)
	serv.server.Stop(serv.timeout)
	if serv.CancelOnShutdown {
		serv.CancelAll()
	}
	return true
}

// setup initializes the underlying graceful server.
func (serv *GracefulServer) setup(addr string) {
	serv.server = &graceful.Server{
		Timeout: serv.timeout,
		Server:  &http.Server{Addr: addr, Handler: serv.Mux},
	}
}
