package middleware

import (
	"net/http"
	"time"

	"github.com/goadesign/goa"

	"golang.org/x/net/context"
)

// Timeout sets a global timeout for all controller actions.
// The timeout notification is made through the context, it is the responsability of the request
// handler to handle it. For example:
//
// 	func (ctrl *Controller) DoLongRunningAction(ctx *DoLongRunningActionContext) error {
// 		action := NewLongRunning()      // setup long running action
//		c := make(chan error, 1)        // create return channel
//		go func() { c <- action.Run() } // Launch long running action goroutine
//		select {
//		case <- ctx.Done():             // timeout triggered
//			action.Cancel()         // cancel long running action
//			<-c                     // wait for Run to return.
//			return ctx.Err()        // retrieve cancel reason
//		case err := <-c:   		// action finished on time
//			return err  		// forward its return value
//		}
//	}
//
// Package golang.org/x/net/context/ctxhttp contains an implementation of an HTTP client which is
// context-aware:
//
// 	func (ctrl *Controller) HttpAction(ctx *HttpActionContext) error {
//		req, err := http.NewRequest("GET", "http://iamaslowservice.com", nil)
//		// ...
//		resp, err := ctxhttp.Do(ctx, nil, req) // returns if timeout triggers
//		// ...
// 	}
//
// Controller actions can check if a timeout is set by calling the context Deadline method.
func Timeout(timeout time.Duration) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			nctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return h(nctx, rw, req)
		}
	}
}
