package goa_test

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
)

var _ = Describe("Application", func() {
	const appName = "foo"
	var s goa.Service

	BeforeEach(func() {
		s = goa.New(appName)
	})

	Describe("New", func() {
		It("creates an application", func() {
			Ω(s).ShouldNot(BeNil())
		})

		It("initializes the application fields", func() {
			Ω(s.Name()).Should(Equal(appName))
			Ω(s).Should(BeAssignableToTypeOf(&goa.Application{}))
			app, _ := s.(*goa.Application)
			Ω(app.Name()).Should(Equal(appName))
			Ω(app.Logger).ShouldNot(BeNil())
			Ω(app.Router).ShouldNot(BeNil())
		})
	})

	Describe("Use", func() {
		Context("with a valid middleware", func() {
			var m goa.Middleware

			BeforeEach(func() {
				m = goa.RequestID()
			})

			JustBeforeEach(func() {
				s.Use(m)
			})

			It("adds the middleware", func() {
				ctrl := s.NewController("test")
				Ω(ctrl.MiddlewareChain()).Should(HaveLen(1))
				Ω(ctrl.MiddlewareChain()[0]).Should(BeAssignableToTypeOf(goa.RequestID()))
			})
		})
	})

	Describe("NewHTTPRouterHandle", func() {
		const resName = "res"
		const actName = "act"
		var handler goa.Handler
		const respStatus = 200
		var respContent = []byte("response")

		var httpHandle httprouter.Handle
		var ctx *goa.Context

		JustBeforeEach(func() {
			ctrl := s.NewController("test")
			httpHandle = ctrl.NewHTTPRouterHandle(actName, handler)
		})

		BeforeEach(func() {
			handler = func(c *goa.Context) error {
				ctx = c
				c.Respond(respStatus, respContent)
				return nil
			}
		})

		It("creates a handle", func() {
			Ω(httpHandle).ShouldNot(BeNil())
		})

		Context("with a request", func() {
			var rw http.ResponseWriter
			var r *http.Request
			var p httprouter.Params

			BeforeEach(func() {
				var err error
				r, err = http.NewRequest("GET", "/foo", nil)
				Ω(err).ShouldNot(HaveOccurred())
				rw = new(TestResponseWriter)
				id := httprouter.Param{Key: "id", Value: "42"}
				query := httprouter.Param{Key: "sort", Value: "asc"}
				p = httprouter.Params{id, query}
			})

			JustBeforeEach(func() {
				httpHandle(rw, r, p)
			})

			It("creates a handle that handles the request", func() {
				i, ok := ctx.Get("id")
				Ω(ok).Should(BeTrue())
				Ω(i).Should(Equal("42"))
				s, ok := ctx.Get("sort")
				Ω(ok).Should(BeTrue())
				Ω(s).Should(Equal("asc"))
				tw := rw.(*TestResponseWriter)
				Ω(tw.Status).Should(Equal(respStatus))
				Ω(tw.Body).Should(Equal(respContent))
			})

			Context("and middleware", func() {
				middlewareCalled := false

				BeforeEach(func() {
					s.Use(TMiddleware(&middlewareCalled))
				})

				It("calls the middleware", func() {
					Ω(middlewareCalled).Should(BeTrue())
				})
			})

			Context("and a middleware chain", func() {
				middlewareCalled := false
				secondCalled := false

				BeforeEach(func() {
					s.Use(TMiddleware(&middlewareCalled))
					s.Use(SecondMiddleware(&middlewareCalled, &secondCalled))
				})

				It("calls the middleware in the right order", func() {
					Ω(middlewareCalled).Should(BeTrue())
					Ω(secondCalled).Should(BeTrue())
				})
			})

			Context("with a handler that fails", func() {
				errorHandlerCalled := false

				BeforeEach(func() {
					s.SetErrorHandler(TErrorHandler(&errorHandlerCalled))
				})

				Context("by returning an error", func() {
					BeforeEach(func() {
						handler = func(ctx *goa.Context) error {
							return fmt.Errorf("boom")
						}
					})

					It("triggers the error handler", func() {
						Ω(errorHandlerCalled).Should(BeTrue())
					})
				})

				Context("by not handling the request", func() {
					BeforeEach(func() {
						handler = func(ctx *goa.Context) error {
							return nil
						}
					})

					It("triggers the error handler", func() {
						Ω(errorHandlerCalled).Should(BeTrue())
					})
				})
			})
		})
	})
})

func TErrorHandler(witness *bool) goa.ErrorHandler {
	return func(ctx *goa.Context, err error) {
		*witness = true
	}
}

func TMiddleware(witness *bool) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx *goa.Context) error {
			*witness = true
			return h(ctx)
		}
	}
}

func SecondMiddleware(witness1, witness2 *bool) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx *goa.Context) error {
			if !*witness1 {
				panic("middleware called in wrong order")
			}
			*witness2 = true
			return h(ctx)
		}
	}
}

type TestResponseWriter struct {
	ParentHeader http.Header
	Body         []byte
	Status       int
}

func (t *TestResponseWriter) Header() http.Header {
	return t.ParentHeader
}

func (t *TestResponseWriter) Write(b []byte) (int, error) {
	t.Body = append(t.Body, b...)
	return len(b), nil
}

func (t *TestResponseWriter) WriteHeader(s int) {
	t.Status = s
}
