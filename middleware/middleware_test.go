package middleware_test

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newService(logger goa.Logger) *goa.Service {
	service := goa.New("test")
	service.Encoder(goa.NewJSONEncoder, "*/*")
	service.UseLogger(logger)
	return service
}

func newContext(service *goa.Service, rw http.ResponseWriter, req *http.Request, params url.Values) context.Context {
	ctrl := service.NewController("test")
	return goa.NewContext(ctrl.Context, rw, req, params)
}

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
		var ctx context.Context
		var req *http.Request
		var rw http.ResponseWriter
		var params url.Values

		BeforeEach(func() {
			service = newService(nil)
			var err error
			req, err = http.NewRequest("GET", "/goo", nil)
			Ω(err).ShouldNot(HaveOccurred())
			rw = new(testResponseWriter)
			params = url.Values{"query": []string{"value"}}
			ctx = newContext(service, rw, req, params)
			Ω(goa.ContextResponse(ctx).Status).Should(Equal(0))
		})

		Context("using a goa handler", func() {
			BeforeEach(func() {
				var goaHandler goa.Handler = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					return goa.ContextResponse(ctx).Send(ctx, 200, "ok")
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
					return goa.ContextResponse(ctx).Send(ctx, 200, "ok")
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
				h := func(c context.Context, rw http.ResponseWriter, req *http.Request) error {
					return goa.ContextResponse(ctx).Send(ctx, 200, "ok")
				}
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(goa.ContextResponse(ctx).Status).Should(Equal(200))
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
				h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error { return nil }
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(rw.(*testResponseWriter).Status).Should(Equal(200))
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
				h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
					return nil
				}
				Ω(middleware(h)(ctx, rw, req)).ShouldNot(HaveOccurred())
				Ω(rw.(*testResponseWriter).Status).Should(Equal(200))
			})
		})

	})
})

var _ = Describe("LogRequest", func() {
	var ctx context.Context
	var rw http.ResponseWriter
	var req *http.Request
	var params url.Values
	var logger *testLogger

	payload := map[string]interface{}{"payload": 42}

	BeforeEach(func() {
		logger = new(testLogger)
		service := newService(logger)

		var err error
		req, err = http.NewRequest("POST", "/goo?param=value", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw = new(testResponseWriter)
		params = url.Values{"query": []string{"value"}}
		ctrl := service.NewController("test")
		ctx = goa.NewContext(ctrl.Context, rw, req, params)
		goa.ContextRequest(ctx).Payload = payload
	})

	It("logs requests", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			return goa.ContextResponse(ctx).Send(ctx, 200, "ok")
		}
		lg := middleware.LogRequest(true)(h)
		Ω(lg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(logger.InfoEntries).Should(HaveLen(4))

		Ω(logger.InfoEntries[0].Data).Should(HaveLen(4))
		Ω(logger.InfoEntries[0].Data[0]).Should(Equal("id"))
		Ω(logger.InfoEntries[0].Data[2]).Should(Equal("POST"))
		Ω(logger.InfoEntries[0].Data[3]).Should(Equal("/goo?param=value"))

		Ω(logger.InfoEntries[1].Data).Should(HaveLen(4))
		Ω(logger.InfoEntries[0].Data[0]).Should(Equal("id"))
		Ω(logger.InfoEntries[1].Data[2]).Should(Equal("query"))
		Ω(logger.InfoEntries[1].Data[3]).Should(Equal("value"))

		Ω(logger.InfoEntries[2].Data).Should(HaveLen(4))
		Ω(logger.InfoEntries[0].Data[0]).Should(Equal("id"))
		Ω(logger.InfoEntries[2].Data[2]).Should(Equal("payload"))
		Ω(logger.InfoEntries[2].Data[3]).Should(Equal(42))

		Ω(logger.InfoEntries[3].Data).Should(HaveLen(8))
		Ω(logger.InfoEntries[0].Data[0]).Should(Equal("id"))
		Ω(logger.InfoEntries[3].Data[2]).Should(Equal("status"))
		Ω(logger.InfoEntries[3].Data[3]).Should(Equal(200))
		Ω(logger.InfoEntries[3].Data[4]).Should(Equal("bytes"))
		Ω(logger.InfoEntries[3].Data[5]).Should(Equal(5))
		Ω(logger.InfoEntries[3].Data[6]).Should(Equal("time"))
	})
})

var _ = Describe("LogResponse", func() {
	var logger *testLogger
	var ctx context.Context
	var req *http.Request
	var rw http.ResponseWriter
	var params url.Values
	responseText := "some response data to be logged"

	BeforeEach(func() {
		logger = new(testLogger)
		service := newService(logger)

		var err error
		req, err = http.NewRequest("POST", "/goo", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw = new(testResponseWriter)
		params = url.Values{"query": []string{"value"}}
		ctx = newContext(service, rw, req, params)
	})

	It("logs responses", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			goa.ContextResponse(ctx).WriteHeader(200)
			goa.ContextResponse(ctx).Write([]byte(responseText))
			return nil
		}
		lg := middleware.LogResponse()(h)
		Ω(lg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(logger.InfoEntries).Should(HaveLen(1))

		Ω(logger.InfoEntries[0].Data).Should(HaveLen(2))
		Ω(logger.InfoEntries[0].Data[0]).Should(Equal("body"))
		Ω(logger.InfoEntries[0].Data[1]).Should(Equal(responseText))
	})
})

var _ = Describe("RequestID", func() {
	const reqID = "request id"
	var ctx context.Context
	var rw http.ResponseWriter
	var req *http.Request
	var params url.Values

	BeforeEach(func() {
		service := newService(nil)

		var err error
		req, err = http.NewRequest("GET", "/goo", nil)
		Ω(err).ShouldNot(HaveOccurred())
		req.Header.Set("X-Request-Id", reqID)
		rw = new(testResponseWriter)
		params = url.Values{"query": []string{"value"}}
		service.Encoder(goa.NewJSONEncoder, "*/*")
		ctx = newContext(service, rw, req, params)
	})

	It("sets the request ID in the context", func() {
		var newCtx context.Context
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newCtx = ctx
			return goa.ContextResponse(ctx).Send(ctx, 200, "ok")
		}
		rg := middleware.RequestID()(h)
		Ω(rg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(newCtx.Value(middleware.ReqIDKey)).Should(Equal(reqID))
	})
})

var _ = Describe("Recover", func() {
	It("recovers", func() {
		service := newService(nil)
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			panic("boom")
		}
		rg := middleware.Recover()(h)
		service.Encoder(goa.NewJSONEncoder, "*/*")
		rw := new(testResponseWriter)
		ctx := newContext(service, rw, nil, nil)
		err := rg(ctx, rw, nil)
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal("panic: boom"))
	})
})

var _ = Describe("Timeout", func() {
	It("sets a deadline", func() {
		service := newService(nil)

		req, err := http.NewRequest("POST", "/goo", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw := new(testResponseWriter)
		ctx := newContext(service, rw, req, nil)
		var newCtx context.Context
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newCtx = ctx
			return goa.ContextResponse(ctx).Send(ctx, 200, "ok")
		}
		t := middleware.Timeout(time.Duration(1))(h)
		err = t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		_, ok := newCtx.Deadline()
		Ω(ok).Should(BeTrue())
	})
})

var _ = Describe("RequireHeader", func() {
	var ctx context.Context
	var req *http.Request
	var rw http.ResponseWriter
	headerName := "Some-Header"

	BeforeEach(func() {
		var err error
		service := newService(nil)
		req, err = http.NewRequest("POST", "/foo/bar", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw = new(testResponseWriter)
		ctx = newContext(service, rw, req, nil)
	})

	It("matches a header value", func() {
		req.Header.Set(headerName, "some value")
		var newCtx context.Context
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newCtx = ctx
			return goa.ContextResponse(ctx).Send(ctx, http.StatusOK, "ok")
		}
		t := middleware.RequireHeader(
			regexp.MustCompile("^/foo"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(goa.ContextResponse(newCtx).Status).Should(Equal(http.StatusOK))
	})

	It("responds with failure on mismatch", func() {
		req.Header.Set(headerName, "some other value")
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			panic("unreachable")
		}
		t := middleware.RequireHeader(
			regexp.MustCompile("^/foo"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(goa.ContextResponse(ctx).Status).Should(Equal(http.StatusUnauthorized))
	})

	It("responds with failure when header is missing", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			panic("unreachable")
		}
		t := middleware.RequireHeader(
			regexp.MustCompile("^/foo"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(goa.ContextResponse(ctx).Status).Should(Equal(http.StatusUnauthorized))
	})

	It("passes through for a non-matching path", func() {
		var newCtx context.Context
		req.Header.Set(headerName, "bogus")
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newCtx = ctx
			return goa.ContextResponse(ctx).Send(ctx, http.StatusOK, "ok")
		}
		t := middleware.RequireHeader(
			regexp.MustCompile("^/baz"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(goa.ContextResponse(newCtx).Status).Should(Equal(http.StatusOK))
	})

	It("matches value for a nil path pattern", func() {
		req.Header.Set(headerName, "bogus")
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			panic("unreachable")
		}
		t := middleware.RequireHeader(
			nil,
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusNotFound)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(goa.ContextResponse(ctx).Status).Should(Equal(http.StatusNotFound))
	})
})

type logEntry struct {
	Msg  string
	Data []interface{}
}

type testLogger struct {
	InfoEntries  []logEntry
	ErrorEntries []logEntry
}

func (t *testLogger) Info(msg string, data ...interface{}) {
	e := logEntry{msg, data}
	t.InfoEntries = append(t.InfoEntries, e)
}

func (t *testLogger) Error(msg string, data ...interface{}) {
	e := logEntry{msg, data}
	t.ErrorEntries = append(t.ErrorEntries, e)
}

type testResponseWriter struct {
	ParentHeader http.Header
	Body         []byte
	Status       int
}

func (t *testResponseWriter) Header() http.Header {
	return t.ParentHeader
}

func (t *testResponseWriter) Write(b []byte) (int, error) {
	t.Body = append(t.Body, b...)
	return len(b), nil
}

func (t *testResponseWriter) WriteHeader(s int) {
	t.Status = s
}
