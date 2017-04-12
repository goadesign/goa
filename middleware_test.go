package goa_test

import (
	"fmt"
	"net/http"

	"context"

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
		var service *goa.Service
		var req *http.Request
		var rw http.ResponseWriter
		var ctx context.Context

		BeforeEach(func() {
			service = goa.New("test")
			ctrl := service.NewController("foo")
			var err error
			req, err = http.NewRequest("GET", "/goo", nil)
			Ω(err).ShouldNot(HaveOccurred())
			rw = new(TestResponseWriter)
			ctx = goa.NewContext(ctrl.Context, rw, req, nil)
			Ω(goa.ContextResponse(ctx).Status).Should(Equal(0))
		})

		Context("using a goa handler", func() {
			BeforeEach(func() {
				var goaHandler goa.Handler = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					service.Send(ctx, 200, "ok")
					return nil
				}
				input = goaHandler
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error { return nil }
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(goa.ContextResponse(ctx).Status).Should(Equal(200))
			})
		})

		Context("using a goa handler func", func() {
			BeforeEach(func() {
				input = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					service.Send(ctx, 200, "ok")
					return nil
				}
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error { return nil }
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(goa.ContextResponse(ctx).Status).Should(Equal(200))
			})
		})

		Context("using a http middleware func", func() {
			BeforeEach(func() {
				input = func(h http.Handler) http.Handler { return h }
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					service.Send(ctx, 200, "ok")
					return nil
				}
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(goa.ContextResponse(ctx).Status).Should(Equal(200))
			})
		})

		Context("using a http handler", func() {
			BeforeEach(func() {
				input = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Write([]byte("ok"))
				})
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					return nil
				}
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(rw.(*TestResponseWriter).Status).Should(Equal(200))
			})
		})

		Context("using a http handler func", func() {
			BeforeEach(func() {
				input = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Write([]byte("ok"))
				}
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					return nil
				}
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(rw.(*TestResponseWriter).Status).Should(Equal(200))
			})
		})

	})
})
