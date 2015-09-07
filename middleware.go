package goa

import (
	"fmt"
	"net/http"
)

type (
	// Middleware represents the canonical goa middleware signature.
	Middleware func(Handler) Handler
)

// NewMiddleware creates a goa middleware from the given argument. The allowed types for the
// argument are:
//
// - a goa middleware: Middleware, func(Handler) Handler
//
// - a goa handler: Handler, func(Context) error
//
// - a http middleware: func(http.Handler) http.Handler
//
// - or a http handler: http.Handler, func(http.ResponseWriter, *http.Request)
//
// An error is returned if the given argument is not one of the types above.
func NewMiddleware(m interface{}) (mw Middleware, err error) {
	switch m := m.(type) {
	case Middleware:
		mw = m
	case func(Handler) Handler:
		mw = m
	case Handler:
		mw = handlerToMiddleware(m)
	case func(Context) error:
		mw = handlerToMiddleware(m)
	case func(http.Handler) http.Handler:
		mw = func(h Handler) Handler {
			return func(c Context) (err error) {
				m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := c.New(w, r)
					err = h(ctx)
				})).ServeHTTP(c.ResponseWriter(), c.Request())
				return
			}
		}
	case http.Handler:
		mw = httpHandlerToMiddleware(m.ServeHTTP)
	case func(http.ResponseWriter, *http.Request):
		mw = httpHandlerToMiddleware(m)
	default:
		err = fmt.Errorf("invalid middleware %#v", m)
	}
	return
}

// handlerToMiddleware creates a middleware from a raw handler.
// The middleware calls the handler and either breaks the middleware chain if the handler returns
// an error by also returning the error or calls the next handler in the chain otherwise.
func handlerToMiddleware(m Handler) Middleware {
	return func(h Handler) Handler {
		return func(c Context) error {
			if err := m(c); err != nil {
				return err
			}
			return h(c)
		}
	}
}

// httpHandlerToMiddleware creates a middleware from a http.HandlerFunc.
// The middleware calls the ServerHTTP method exposed by the http handler and then calls the next
// middleware in the chain.
func httpHandlerToMiddleware(m http.HandlerFunc) Middleware {
	return func(h Handler) Handler {
		return func(c Context) error {
			m.ServeHTTP(c.ResponseWriter(), c.Request())
			return h(c)
		}
	}
}
