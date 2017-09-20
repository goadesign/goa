package goa_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"context"

	"github.com/goadesign/goa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {
	const appName = "foo"
	var s *goa.Service

	BeforeEach(func() {
		s = goa.New(appName)
		s.Decoder.Register(goa.NewJSONDecoder, "*/*")
		s.Encoder.Register(goa.NewJSONEncoder, "*/*")
	})

	Describe("New", func() {
		It("creates a service", func() {
			Ω(s).ShouldNot(BeNil())
		})

		It("initializes the service fields", func() {
			Ω(s.Name).Should(Equal(appName))
			Ω(s.Mux).ShouldNot(BeNil())
			Ω(s.Server).ShouldNot(BeNil())
		})
	})

	Describe("NotFound", func() {
		var rw *TestResponseWriter
		var req *http.Request

		BeforeEach(func() {
			req, _ = http.NewRequest("GET", "/foo", nil)
			rw = &TestResponseWriter{ParentHeader: make(http.Header)}
		})

		JustBeforeEach(func() {
			s.Mux.ServeHTTP(rw, req)
		})

		It("handles requests with no registered handlers", func() {
			Ω(string(rw.Body)).Should(MatchRegexp(`{"id":".*","code":"not_found","status":404,"detail":"/foo"}` + "\n"))
		})

		Context("with middleware", func() {
			middlewareCalled := false

			BeforeEach(func() {
				s.Use(TMiddleware(&middlewareCalled))
				// trigger finalize
				ctrl := s.NewController("test")
				ctrl.MuxHandler("", nil, nil)
			})

			It("calls the middleware", func() {
				Ω(middlewareCalled).Should(BeTrue())
			})
		})

		Context("middleware and multiple controllers", func() {
			middlewareCalled := 0

			BeforeEach(func() {
				s.Use(CMiddleware(&middlewareCalled))
				ctrl := s.NewController("test")
				ctrl.MuxHandler("/foo", nil, nil)
				ctrl.MuxHandler("/bar", nil, nil)
			})

			It("calls the middleware once", func() {
				Ω(middlewareCalled).Should(Equal(1))
			})
		})
	})

	Describe("MethodNotAllowed", func() {
		var rw *TestResponseWriter
		var req *http.Request

		JustBeforeEach(func() {
			rw = &TestResponseWriter{ParentHeader: http.Header{}}
			s.Mux.ServeHTTP(rw, req)
		})

		BeforeEach(func() {
			req, _ = http.NewRequest("GET", "/foo", nil)
			s.Mux.Handle("POST", "/foo", func(rw http.ResponseWriter, req *http.Request, vals url.Values) {})
			s.Mux.Handle("PUT", "/foo", func(rw http.ResponseWriter, req *http.Request, vals url.Values) {})
		})

		It("handles requests with wrong method but existing endpoint", func() {
			Ω(rw.Status).Should(Equal(405))
			Ω(rw.Header().Get("Allow")).Should(Or(Equal("POST, PUT"), Equal("PUT, POST")))
			Ω(string(rw.Body)).Should(MatchRegexp(`{"id":".*","code":"method_not_allowed","status":405,"detail":".*","meta":{.*}}` + "\n"))
		})
	})

	Describe("MaxRequestBodyLength", func() {
		var rw *TestResponseWriter
		var req *http.Request
		var muxHandler goa.MuxHandler

		BeforeEach(func() {
			body := bytes.NewBuffer([]byte{'"', '2', '3', '4', '"'})
			req, _ = http.NewRequest("GET", "/foo", body)
			rw = &TestResponseWriter{ParentHeader: make(http.Header)}
			ctrl := s.NewController("test")
			ctrl.MaxRequestBodyLength = 4
			unmarshaler := func(ctx context.Context, service *goa.Service, req *http.Request) error {
				_, err := ioutil.ReadAll(req.Body)
				return err
			}
			handler := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				rw.WriteHeader(400)
				rw.Write([]byte(goa.ContextError(ctx).Error()))
				return nil
			}
			muxHandler = ctrl.MuxHandler("testMax", handler, unmarshaler)
		})

		JustBeforeEach(func() {
			muxHandler(rw, req, nil)
		})

		It("prevents reading more bytes", func() {
			Ω(string(rw.Body)).Should(MatchRegexp(`\[.*\] 413 request_too_large: request body length exceeds 4 bytes`))
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
				if err := goa.ContextError(c); err != nil {
					rw.WriteHeader(400)
					rw.Write([]byte(err.Error()))
					return nil
				}
				goa.ContextRequest(c).Request = req
				ctx = c
				rw.WriteHeader(respStatus)
				rw.Write(respContent)
				return nil
			}
			unmarshaler = func(c context.Context, service *goa.Service, req *http.Request) error {
				ctx = c
				if req != nil {
					var payload interface{}
					err := service.DecodeRequest(req, &payload)
					if err != nil {
						return err
					}
					goa.ContextRequest(ctx).Payload = payload
				}
				return nil
			}
		})

		It("creates a handler", func() {
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
				rw = &TestResponseWriter{ParentHeader: make(http.Header)}
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

			Context("with an invalid payload", func() {
				BeforeEach(func() {
					r.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("not json")))
					r.ContentLength = 8
				})

				It("triggers the error handler", func() {
					Ω(rw.(*TestResponseWriter).Status).Should(Equal(400))
					Ω(string(rw.(*TestResponseWriter).Body)).Should(ContainSubstring("failed to decode"))
				})

				Context("then a valid payload", func() {
					It("then succeeds", func() {
						var err error
						r, err = http.NewRequest("GET", "/foo2", nil)
						Ω(err).ShouldNot(HaveOccurred())
						rw = &TestResponseWriter{ParentHeader: make(http.Header)}
						muxHandler(rw, r, p)
						Ω(rw.(*TestResponseWriter).Status).Should(Equal(200))
					})
				})
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
					s.Use(TErrorHandler(&errorHandlerCalled))
				})

				Context("by returning an error", func() {
					BeforeEach(func() {
						handler = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
							return fmt.Errorf("boom")
						}
						s.WithLogger(nil)
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
						s.Decoder.Register(goa.NewJSONDecoder, "*/*")
						r.Header.Set("Content-Type", "application/octet-stream")
					})

					It("should use the default decoder", func() {
						Ω(goa.ContextRequest(ctx).Payload).Should(Equal(decodedContent))
					})
				})

				Context("with a Content-Type of 'application/octet-stream' or any other and no default decoder", func() {
					BeforeEach(func() {
						s = goa.New("test")
						s.Decoder.Register(goa.NewJSONDecoder, "application/json")
						r.Header.Set("Content-Type", "application/octet-stream")
					})

					It("should bypass decoding", func() {
						Ω(goa.ContextRequest(ctx).Payload).Should(BeNil())
					})
				})
			})
		})
	})

	Describe("FileHandler", func() {
		const publicPath = "github.com/goadesign/goa/public"

		var outDir string

		var handler goa.Handler
		const respStatus = 200
		var respContent = []byte(`{"foo":"bar"}`)

		var muxHandler goa.MuxHandler

		JustBeforeEach(func() {
			gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
			outDir = filepath.Join(gopath, "src", publicPath)
			err := os.MkdirAll(filepath.Join(outDir, "swagger"), 0777)
			Ω(err).ShouldNot(HaveOccurred())
			file, err := os.Create(filepath.Join(outDir, "swagger", "swagger.json"))
			Ω(err).ShouldNot(HaveOccurred())
			_, err = file.Write(respContent)
			Ω(err).ShouldNot(HaveOccurred())
			file.Close()

			ctrl := s.NewController("test")
			handler = ctrl.FileHandler("/swagger.json", "public/swagger/swagger.json")
			muxHandler = ctrl.MuxHandler("testAct", handler, nil)
		})

		AfterEach(func() {
			os.RemoveAll(outDir)
		})

		It("creates a handler", func() {
			Ω(muxHandler).ShouldNot(BeNil())
		})

		Context("with a request", func() {
			var rw http.ResponseWriter
			var r *http.Request
			var p url.Values

			BeforeEach(func() {
				var err error
				r, err = http.NewRequest("GET", "/swagger.json", nil)
				Ω(err).ShouldNot(HaveOccurred())
				rw = &TestResponseWriter{ParentHeader: make(http.Header)}
			})

			JustBeforeEach(func() {
				muxHandler(rw, r, p)
			})

			It("creates a handle that handles the request", func() {
				tw := rw.(*TestResponseWriter)
				Ω(tw.Status).Should(Equal(respStatus))
				Ω(tw.Body).Should(Equal(respContent))
			})
		})
	})
})

func TErrorHandler(witness *bool) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			err := h(ctx, rw, req)
			if err != nil {
				*witness = true
			}
			return nil
		}
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

func CMiddleware(witness *int) goa.Middleware {
	return func(h goa.Handler) goa.Handler {
		return func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			*witness++
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
