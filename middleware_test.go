package goa_test

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
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

var _ = Describe("LogRequest", func() {
	var handler *testHandler
	var ctx *goa.Context
	var service goa.Service
	params := url.Values{"param": []string{"value"}}
	payload := map[string]interface{}{"payload": 42}

	BeforeEach(func() {
		service = goa.New("test")
		req, err := http.NewRequest("POST", "/goo", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw := new(TestResponseWriter)
		ctx = goa.NewContext(nil, service, req, rw, params)
		ctx.SetPayload(payload)
		handler = new(testHandler)
		logger := log15.New("test", "test")
		logger.SetHandler(handler)
		ctx.Logger = logger
	})

	It("logs requests", func() {
		h := func(ctx *goa.Context) error {
			ctx.Respond(200, "ok")
			return nil
		}
		lg := goa.LogRequest()(h)
		Ω(lg(ctx)).ShouldNot(HaveOccurred())
		Ω(handler.Records).Should(HaveLen(4))

		Ω(handler.Records[0].Ctx).Should(HaveLen(6))
		Ω(handler.Records[0].Ctx[4]).Should(Equal("POST"))
		Ω(handler.Records[0].Ctx[5]).Should(Equal("/goo"))

		Ω(handler.Records[1].Ctx).Should(HaveLen(6))
		Ω(handler.Records[1].Ctx[4]).Should(Equal("param"))
		Ω(handler.Records[1].Ctx[5]).Should(Equal([]string{"value"}))

		Ω(handler.Records[2].Ctx).Should(HaveLen(6))
		Ω(handler.Records[2].Ctx[4]).Should(Equal("payload"))
		Ω(handler.Records[2].Ctx[5]).Should(Equal(42))

		Ω(handler.Records[3].Ctx).Should(HaveLen(10))
		Ω(handler.Records[3].Ctx[4]).Should(Equal("status"))
		Ω(handler.Records[3].Ctx[6]).Should(Equal("bytes"))
		Ω(handler.Records[3].Ctx[5]).Should(Equal(200))
		Ω(handler.Records[3].Ctx[7]).Should(Equal(5))
		Ω(handler.Records[3].Ctx[8]).Should(Equal("time"))
	})
})

var _ = Describe("LogResponse", func() {
	var handler *testHandler
	var ctx *goa.Context
	params := url.Values{"param": []string{"value"}}
	payload := map[string]interface{}{"payload": 42}
	responseText := "some response data to be logged"

	BeforeEach(func() {
		req, err := http.NewRequest("POST", "/goo", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw := new(TestResponseWriter)
		ctx = goa.NewContext(nil, goa.New("test"), req, rw, params)
		ctx.SetPayload(payload)
		handler = new(testHandler)
		logger := log15.New("test", "test")
		logger.SetHandler(handler)
		ctx.Logger = logger
	})

	It("logs responses", func() {
		h := func(ctx *goa.Context) error {
			ctx.RespondBytes(200, []byte(responseText))
			return nil
		}
		lg := goa.LogResponse()(h)
		Ω(lg(ctx)).ShouldNot(HaveOccurred())
		Ω(handler.Records).Should(HaveLen(1))

		Ω(handler.Records[0].Ctx).Should(HaveLen(4))
		Ω(handler.Records[0].Ctx[2]).Should(Equal("raw"))
		Ω(handler.Records[0].Ctx[3]).Should(Equal(responseText))
	})
})

var _ = Describe("RequestID", func() {
	const reqID = "request id"
	var ctx *goa.Context

	BeforeEach(func() {
		req, err := http.NewRequest("GET", "/goo", nil)
		Ω(err).ShouldNot(HaveOccurred())
		req.Header.Set("X-Request-Id", reqID)
		ctx = goa.NewContext(nil, goa.New("test"), req, new(TestResponseWriter), nil)
	})

	It("sets the request ID in the context", func() {
		h := func(ctx *goa.Context) error {
			ctx.Respond(200, "ok")
			return nil
		}
		rg := goa.RequestID()(h)
		Ω(rg(ctx)).ShouldNot(HaveOccurred())
		Ω(ctx.Value(goa.ReqIDKey)).Should(Equal(reqID))
	})
})

var _ = Describe("Recover", func() {
	It("recovers", func() {
		h := func(ctx *goa.Context) error {
			panic("boom")
		}
		rg := goa.Recover()(h)
		err := rg(goa.NewContext(nil, goa.New("test"), nil, nil, nil))
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal("panic: boom"))
	})
})

var _ = Describe("Timeout", func() {
	It("sets a deadline", func() {
		h := func(ctx *goa.Context) error {
			ctx.Respond(200, "ok")
			return nil
		}
		req, err := http.NewRequest("POST", "/goo", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		t := goa.Timeout(time.Duration(1))(h)
		ctx := goa.NewContext(nil, goa.New("test"), req, nil, nil)
		err = t(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		_, ok := ctx.Deadline()
		Ω(ok).Should(BeTrue())
	})
})

var _ = Describe("RequireHeader", func() {
	var handler *testHandler
	var ctx *goa.Context
	var req *http.Request
	params := url.Values{"param": []string{"value"}}
	payload := map[string]interface{}{"payload": 42}
	headerName := "Some-Header"

	BeforeEach(func() {
		var err error
		req, err = http.NewRequest("POST", "/foo/bar", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw := new(TestResponseWriter)
		ctx = goa.NewContext(nil, goa.New("test"), req, rw, params)
		ctx.SetPayload(payload)
		handler = new(testHandler)
		logger := log15.New("test", "test")
		logger.SetHandler(handler)
		ctx.Logger = logger
	})

	It("matches a header value", func() {
		req.Header.Set(headerName, "some value")
		h := func(ctx *goa.Context) error {
			ctx.Respond(http.StatusOK, "ok")
			return nil
		}
		t := goa.RequireHeader(
			regexp.MustCompile("^/foo"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(ctx.ResponseStatus()).Should(Equal(http.StatusOK))
	})

	It("responds with failure on mismatch", func() {
		req.Header.Set(headerName, "some other value")
		h := func(ctx *goa.Context) error {
			panic("unreachable")
		}
		t := goa.RequireHeader(
			regexp.MustCompile("^/foo"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(ctx.ResponseStatus()).Should(Equal(http.StatusUnauthorized))
	})

	It("responds with failure when header is missing", func() {
		h := func(ctx *goa.Context) error {
			panic("unreachable")
		}
		t := goa.RequireHeader(
			regexp.MustCompile("^/foo"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(ctx.ResponseStatus()).Should(Equal(http.StatusUnauthorized))
	})

	It("passes through for a non-matching path", func() {
		req.Header.Set(headerName, "bogus")
		h := func(ctx *goa.Context) error {
			ctx.Respond(http.StatusOK, "ok")
			return nil
		}
		t := goa.RequireHeader(
			regexp.MustCompile("^/baz"),
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusUnauthorized)(h)
		err := t(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(ctx.ResponseStatus()).Should(Equal(http.StatusOK))
	})

	It("matches value for a nil path pattern", func() {
		req.Header.Set(headerName, "bogus")
		h := func(ctx *goa.Context) error {
			panic("unreachable")
		}
		t := goa.RequireHeader(
			nil,
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusNotFound)(h)
		err := t(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(ctx.ResponseStatus()).Should(Equal(http.StatusNotFound))
	})
})

type testHandler struct {
	Records []*log15.Record
}

func (t *testHandler) Log(r *log15.Record) error {
	t.Records = append(t.Records, r)
	return nil
}
