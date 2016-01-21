package goa_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/goadesign/goa"
	"github.com/goadesign/middleware/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			Ω(app.ServeMux).ShouldNot(BeNil())
		})
	})

	Describe("Use", func() {
		Context("with a valid middleware", func() {
			var m goa.Middleware

			BeforeEach(func() {
				m = middleware.RequestID()
			})

			JustBeforeEach(func() {
				s.Use(m)
			})

			It("adds the middleware", func() {
				ctrl := s.NewController("test")
				Ω(ctrl.MiddlewareChain()).Should(HaveLen(1))
				Ω(ctrl.MiddlewareChain()[0]).Should(BeAssignableToTypeOf(middleware.RequestID()))
			})
		})
	})

	Describe("HandleFunc", func() {
		const resName = "res"
		const actName = "act"
		var handler, unmarshaler goa.Handler
		const respStatus = 200
		var respContent = []byte("response")

		var handleFunc goa.HandleFunc
		var ctx *goa.Context

		JustBeforeEach(func() {
			ctrl := s.NewController("test")
			handleFunc = ctrl.HandleFunc(actName, handler, unmarshaler)
		})

		BeforeEach(func() {
			handler = func(c *goa.Context) error {
				ctx = c
				c.RespondBytes(respStatus, respContent)
				return nil
			}
			unmarshaler = func(c *goa.Context) error {
				ctx = c
				req := c.Request()
				if req != nil {
					var payload interface{}
					err := ctx.Service().DecodeRequest(ctx, &payload)
					Ω(err).ShouldNot(HaveOccurred())
					ctx.SetPayload(payload)
				}
				return nil
			}
		})

		It("creates a handle", func() {
			Ω(handleFunc).ShouldNot(BeNil())
		})

		Context("with a request", func() {
			var rw http.ResponseWriter
			var r *http.Request
			var p url.Values

			BeforeEach(func() {
				var err error
				r, err = http.NewRequest("GET", "/foo", nil)
				Ω(err).ShouldNot(HaveOccurred())
				rw = new(TestResponseWriter)
				p = url.Values{"id": []string{"42"}, "sort": []string{"asc"}}
			})

			JustBeforeEach(func() {
				handleFunc(rw, r, p)
			})

			It("creates a handle that handles the request", func() {
				i := ctx.Get("id")
				Ω(i).Should(Equal("42"))
				s := ctx.Get("sort")
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

			Context("with different payload types", func() {
				content := []byte(`{"hello": "world"}`)
				decodedContent := map[string]interface{}{"hello": "world"}

				BeforeEach(func() {
					r.Header.Set("Content-Type", "application/json")
					r.Body = ioutil.NopCloser(bytes.NewReader(content))
					r.ContentLength = int64(len(content))
				})

				It("should work with application/json and load properly", func() {
					Ω(ctx.RawPayload()).Should(Equal(decodedContent))
				})

				Context("with an empty Content-Type", func() {
					BeforeEach(func() {
						delete(r.Header, "Content-Type")
					})

					It("defaults to application/json and loads properly for JSON bodies", func() {
						Ω(ctx.RawPayload()).Should(Equal(decodedContent))
					})
				})

				Context("with a Content-Type of 'application/octet-stream' or any other", func() {
					BeforeEach(func() {
						r.Header.Set("Content-Type", "application/octet-stream")
					})

					It("should have a nil payload", func() {
						Ω(ctx.RawPayload()).Should(BeNil())
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
