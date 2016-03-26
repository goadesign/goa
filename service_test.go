package goa_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {
	const appName = "foo"
	var s *goa.Service

	BeforeEach(func() {
		s = goa.New(appName)
		s.Decoder(goa.NewJSONDecoder, "*/*")
		s.Encoder(goa.NewJSONEncoder, "*/*")
	})

	Describe("New", func() {
		It("creates a service", func() {
			Ω(s).ShouldNot(BeNil())
		})

		It("initializes the service fields", func() {
			Ω(s.Name).Should(Equal(appName))
			Ω(s.Mux).ShouldNot(BeNil())
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
				Ω(ctrl.Middleware).Should(HaveLen(1))
				Ω(ctrl.Middleware[0]).Should(BeAssignableToTypeOf(middleware.RequestID()))
			})
		})
	})

	Describe("MuxHandler", func() {
		var handler goa.Handler
		var unmarshaler goa.Unmarshaler
		const respStatus = 200
		var respContent = []byte("response")

		var muxHandler goa.MuxHandler
		var ctx context.Context

		JustBeforeEach(func() {
			ctrl := s.NewController("test")
			muxHandler = ctrl.MuxHandler("testAct", handler, unmarshaler)
		})

		BeforeEach(func() {
			handler = func(c context.Context, rw http.ResponseWriter, req *http.Request) error {
				ctx = c
				rw.WriteHeader(respStatus)
				rw.Write(respContent)
				return nil
			}
			unmarshaler = func(c context.Context, req *http.Request) error {
				ctx = c
				if req != nil {
					var payload interface{}
					err := goa.ContextService(ctx).DecodeRequest(req, &payload)
					Ω(err).ShouldNot(HaveOccurred())
					goa.ContextRequest(ctx).Payload = payload
				}
				return nil
			}
		})

		It("creates a handle", func() {
			Ω(muxHandler).ShouldNot(BeNil())
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
				muxHandler(rw, r, p)
			})

			It("creates a handle that handles the request", func() {
				i := goa.ContextRequest(ctx).Params.Get("id")
				Ω(i).Should(Equal("42"))
				s := goa.ContextRequest(ctx).Params.Get("sort")
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
					s.ErrorHandler = TErrorHandler(&errorHandlerCalled)
				})

				Context("by returning an error", func() {
					BeforeEach(func() {
						handler = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
							return fmt.Errorf("boom")
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
					Ω(goa.ContextRequest(ctx).Payload).Should(Equal(decodedContent))
				})

				Context("with an empty Content-Type", func() {
					BeforeEach(func() {
						delete(r.Header, "Content-Type")
					})

					It("defaults to application/json and loads properly for JSON bodies", func() {
						Ω(goa.ContextRequest(ctx).Payload).Should(Equal(decodedContent))
					})
				})

				Context("with a Content-Type of 'application/octet-stream' or any other", func() {
					BeforeEach(func() {
						s.Decoder(goa.NewJSONDecoder, "*/*")
						r.Header.Set("Content-Type", "application/octet-stream")
					})

					It("should use the default decoder", func() {
						Ω(goa.ContextRequest(ctx).Payload).Should(Equal(decodedContent))
					})
				})

				Context("with a Content-Type of 'application/octet-stream' or any other and no default decoder", func() {
					BeforeEach(func() {
						s = goa.New("test")
						s.Decoder(goa.NewJSONDecoder, "application/json")
						r.Header.Set("Content-Type", "application/octet-stream")
					})

					It("should bypass decoding", func() {
						Ω(goa.ContextRequest(ctx).Payload).Should(BeNil())
					})
				})
			})
		})
	})
})

func TErrorHandler(witness *bool) goa.ErrorHandler {
	return func(ctx context.Context, rw http.ResponseWriter, req *http.Request, err error) {
		*witness = true
	}
}

func TMiddleware(witness *bool) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			*witness = true
			return h(ctx, rw, req)
		}
	}
}

func SecondMiddleware(witness1, witness2 *bool) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			if !*witness1 {
				panic("middleware called in wrong order")
			}
			*witness2 = true
			return h(ctx, rw, req)
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
