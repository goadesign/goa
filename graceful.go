// +build !appengine

package goa

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gopkg.in/tylerb/graceful.v1"
)

// GracefulApplication is a goa application using a graceful shutdown server.
// When sending any of the signals listed in InterruptSignals to the process GracefulApplication:
//
// * disables keepalive connections.
//
// * closes the listening socket, allowing another process to listen on that port immediately.
//
// * calls Cancel, signaling all active handlers.
type GracefulApplication struct {
	*Application
	sync.Mutex
	server *graceful.Server

	// Interrupted is true if the application is in the process of shutting down.
	Interrupted bool

	// CancelOnShutdown tells whether existing requests should be canceled when shutdown is
	// triggered (true) or whether to wait until the requests complete (false).
	CancelOnShutdown bool
}

// InterruptSignals is the list of signals that initiate graceful shutdown.
// Note that only SIGINT is supported on Windows so this list should be
// overridden by the caller when running on that platform.
var InterruptSignals = []os.Signal{
	os.Signal(syscall.SIGINT),
	os.Signal(syscall.SIGTERM),
	os.Signal(syscall.SIGQUIT),
}

// NewGraceful returns a goa application that uses a graceful shutdown server.
func NewGraceful(name string, cancelOnShutdown bool) Service {
	app, _ := New(name).(*Application)
	return &GracefulApplication{Application: app, CancelOnShutdown: cancelOnShutdown}
}

// ListenAndServe starts the HTTP server and sets up a listener on the given host/port.
func (gapp *GracefulApplication) ListenAndServe(addr string) error {
	gapp.setup(addr)
	Log.Info(RootContext, "listen", KV{"address", addr})
	if err := gapp.server.ListenAndServe(); err != nil {
		// there may be a final "accept" error after completion of graceful shutdown
		// which can be safely ignored here.
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			return err
		}
	}
	return nil
}

// ListenAndServeTLS starts a HTTPS server and sets up a listener on the given host/port.
func (gapp *GracefulApplication) ListenAndServeTLS(addr, certFile, keyFile string) error {
	gapp.setup(addr)
	Log.Info(RootContext, "listen ssl", KV{"address", addr})
	return gapp.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown initiates graceful shutdown of the running server once. Returns true on
// initial shutdown and false if already shutting down.
func (gapp *GracefulApplication) Shutdown() bool {
	gapp.Lock()
	defer gapp.Unlock()
	if gapp.Interrupted {
		return false
	}
	gapp.Interrupted = true
	gapp.server.Stop(0)
	if gapp.CancelOnShutdown {
		CancelAll()
	}
	return true
}

// setup initializes the interrupt handler and the underlying graceful server.
func (gapp *GracefulApplication) setup(addr string) {
	// we will trap interrupts here instead of allowing the graceful package to do
	// it for us. the graceful package has the odd behavior of stopping the
	// interrupt handler after first interrupt. this leads to the dreaded double-
	// tap because the lack of any viable custom handler means that golang's
	// default handler will kill the process on a second interrupt.
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, InterruptSignals...)

	// Start interrupt handler goroutine
	go func() {
		for signal := range interruptChannel {
			if gapp.Shutdown() {
				Log.Info(RootContext, "Received signal. Initiating graceful shutdown...", KV{"signal", signal})
			} else {
				Log.Info(RootContext, "Received signal. Already gracefully shutting down.", KV{"signal", signal})
			}
		}
	}()

	// note the use of zero timeout (i.e. no forced shutdown timeout) so requests
	// can run as long as they want. there is usually a hard limit to when the
	// response must come back (e.g. the nginx timeout) before being abandoned so
	// the handler should implement some kind of internal timeout (e.g. the go
	// context deadline) instead of relying on a shutdown timeout.
	gapp.server = &graceful.Server{
		Timeout:          0,
		Server:           &http.Server{Addr: addr, Handler: gapp.ServeMux()},
		NoSignalHandling: true,
	}
}
