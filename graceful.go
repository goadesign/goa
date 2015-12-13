// +build !appengine

package goa

import (
	"fmt"
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
func NewGraceful(name string) Service {
	app, _ := New(name).(*Application)
	return &GracefulApplication{Application: app}
}

// ListenAndServe starts the HTTP server and sets up a listener on the given host/port.
func (gapp *GracefulApplication) ListenAndServe(addr string) error {
	gapp.setup(addr)
	gapp.Info("listen", "addr", addr)
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
	gapp.Info("listen ssl", "addr", addr)
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
	Cancel()
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
				gapp.Warn(fmt.Sprintf("Received %v. Initiating graceful shutdown...", signal))
			} else {
				gapp.Warn(fmt.Sprintf("Received %v. Already gracefully shutting down.", signal))
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
