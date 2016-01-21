package goa_test

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/goadesign/goa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewMiddleware", func() {
	var input interface{}
	var middleware goa.Middleware
	var mErr error

	JustBeforeEach(func() {
		middleware, mErr = goa.NewMiddleware(input)
	})

	Context("using a goa Middleware", func() {
		var goaMiddleware goa.Middleware

		BeforeEach(func() {
			goaMiddleware = func(h goa.Handler) goa.Handler { return h }
			input = goaMiddleware
		})

		It("returns the middleware", func() {
			Ω(fmt.Sprintf("%#v", middleware)).Should(Equal(fmt.Sprintf("%#v", goaMiddleware)))
			Ω(mErr).ShouldNot(HaveOccurred())
		})
	})

	Context("using a goa middleware func", func() {
		var goaMiddlewareFunc func(goa.Handler) goa.Handler

		BeforeEach(func() {
			goaMiddlewareFunc = func(h goa.Handler) goa.Handler { return h }
			input = goaMiddlewareFunc
		})

		It("returns the middleware", func() {
			Ω(fmt.Sprintf("%#v", middleware)).Should(Equal(fmt.Sprintf("%#v", goa.Middleware(goaMiddlewareFunc))))
			Ω(mErr).ShouldNot(HaveOccurred())
		})
	})

	Context("with a context", func() {
		var service goa.Service
		var ctx *goa.Context

		BeforeEach(func() {
			service = goa.New("test")
			req, err := http.NewRequest("GET", "/goo", nil)
			Ω(err).ShouldNot(HaveOccurred())
			rw := new(TestResponseWriter)
			params := url.Values{"foo": []string{"bar"}}
			ctx = goa.NewContext(nil, service, req, rw, params)
			Ω(ctx.ResponseStatus()).Should(Equal(0))
		})

		Context("using a goa handler", func() {
			BeforeEach(func() {
				var goaHandler goa.Handler = func(ctx *goa.Context) error {
					ctx.Respond(200, "ok")
					return nil
				}
				input = goaHandler
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx *goa.Context) error { return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

		Context("using a goa handler func", func() {
			BeforeEach(func() {
				input = func(ctx *goa.Context) error {
					ctx.Respond(200, "ok")
					return nil
				}
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx *goa.Context) error { return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

		Context("using a http middleware func", func() {
			BeforeEach(func() {
				input = func(h http.Handler) http.Handler { return h }
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(c *goa.Context) error { c.Respond(200, "ok"); return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

		Context("using a http handler", func() {
			BeforeEach(func() {
				var httpHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("ok"))
					w.WriteHeader(200)
				})
				input = httpHandler
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx *goa.Context) error { return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

		Context("using a http handler func", func() {
			BeforeEach(func() {
				input = func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("ok"))
					w.WriteHeader(200)
				}
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx *goa.Context) error { return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

	})
})
