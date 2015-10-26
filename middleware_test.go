package goa_test

import (
	"fmt"
	"net/http"
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
		var ctx *goa.Context

		BeforeEach(func() {
			req, err := http.NewRequest("GET", "/goo", nil)
			Ω(err).ShouldNot(HaveOccurred())
			rw := new(TestResponseWriter)
			params := map[string]string{"foo": "bar"}
			query := map[string][]string{"filter": []string{"one"}}
			ctx = goa.NewContext(nil, req, rw, params, query, nil)
			Ω(ctx.ResponseWritten()).Should(BeFalse())
		})

		Context("using a goa handler", func() {
			BeforeEach(func() {
				var goaHandler goa.Handler = func(ctx *goa.Context) error {
					ctx.JSON(200, "ok")
					return nil
				}
				input = goaHandler
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx *goa.Context) error { return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseWritten()).Should(BeTrue())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

		Context("using a goa handler func", func() {
			BeforeEach(func() {
				input = func(ctx *goa.Context) error {
					ctx.JSON(200, "ok")
					return nil
				}
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(ctx *goa.Context) error { return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseWritten()).Should(BeTrue())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

		Context("using a http middleware func", func() {
			BeforeEach(func() {
				input = func(h http.Handler) http.Handler { return h }
			})

			It("wraps it in a middleware", func() {
				Ω(mErr).ShouldNot(HaveOccurred())
				h := func(c *goa.Context) error { c.JSON(200, "ok"); return nil }
				Ω(middleware(h)(ctx)).ShouldNot(HaveOccurred())
				Ω(ctx.ResponseWritten()).Should(BeTrue())
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
				Ω(ctx.ResponseWritten()).Should(BeTrue())
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
				Ω(ctx.ResponseWritten()).Should(BeTrue())
				Ω(ctx.ResponseStatus()).Should(Equal(200))
			})
		})

	})
})

var _ = Describe("LogRequest", func() {
	var handler *testHandler
	var ctx *goa.Context
	params := map[string]string{"param": "value"}
	query := map[string][]string{"query": []string{"qvalue"}}
	payload := map[string]interface{}{"payload": 42}

	BeforeEach(func() {
		req, err := http.NewRequest("POST", "/goo", strings.NewReader(`{"payload":42}`))
		Ω(err).ShouldNot(HaveOccurred())
		rw := new(TestResponseWriter)
		ctx = goa.NewContext(nil, req, rw, params, query, payload)
		handler = new(testHandler)
		logger := log15.New("test", "test")
		logger.SetHandler(handler)
		ctx.Logger = logger
	})

	It("logs requests", func() {
		h := func(ctx *goa.Context) error {
			ctx.JSON(200, "ok")
			return nil
		}
		lg := goa.LogRequest()(h)
		Ω(lg(ctx)).ShouldNot(HaveOccurred())
		Ω(handler.Records).Should(HaveLen(5))

		Ω(handler.Records[0].Ctx).Should(HaveLen(6))
		Ω(handler.Records[0].Ctx[4]).Should(Equal("POST"))
		Ω(handler.Records[0].Ctx[5]).Should(Equal("/goo"))

		Ω(handler.Records[1].Ctx).Should(HaveLen(6))
		Ω(handler.Records[1].Ctx[4]).Should(Equal("param"))
		Ω(handler.Records[1].Ctx[5]).Should(Equal("value"))

		Ω(handler.Records[2].Ctx).Should(HaveLen(6))
		Ω(handler.Records[2].Ctx[4]).Should(Equal("query"))
		Ω(handler.Records[2].Ctx[5]).Should(Equal([]string{"qvalue"}))

		Ω(handler.Records[3].Ctx).Should(HaveLen(6))
		Ω(handler.Records[3].Ctx[4]).Should(Equal("payload"))
		Ω(handler.Records[3].Ctx[5]).Should(Equal(42))

		Ω(handler.Records[4].Ctx).Should(HaveLen(10))
		Ω(handler.Records[4].Ctx[4]).Should(Equal("status"))
		Ω(handler.Records[4].Ctx[5]).Should(Equal(200))
		Ω(handler.Records[4].Ctx[6]).Should(Equal("bytes"))
		Ω(handler.Records[4].Ctx[7]).Should(Equal(4))
		Ω(handler.Records[4].Ctx[8]).Should(Equal("time"))
	})
})

var _ = Describe("RequestID", func() {
	const reqID = "request id"
	var ctx *goa.Context

	BeforeEach(func() {
		req, err := http.NewRequest("GET", "/goo", nil)
		Ω(err).ShouldNot(HaveOccurred())
		req.Header.Set("X-Request-Id", reqID)
		ctx = goa.NewContext(nil, req, new(TestResponseWriter), nil, nil, nil)
	})

	It("sets the request ID in the context", func() {
		h := func(ctx *goa.Context) error {
			ctx.JSON(200, "ok")
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
		err := rg(goa.NewContext(nil, nil, nil, nil, nil, nil))
		Ω(err).Should(HaveOccurred())
		Ω(err.Error()).Should(Equal("panic: boom"))
	})
})

var _ = Describe("Timeout", func() {
	It("sets a deadline", func() {
		h := func(ctx *goa.Context) error {
			ctx.JSON(200, "ok")
			return nil
		}
		t := goa.Timeout(time.Duration(1))(h)
		ctx := goa.NewContext(nil, nil, nil, nil, nil, nil)
		err := t(ctx)
		Ω(err).ShouldNot(HaveOccurred())
		_, ok := ctx.Deadline()
		Ω(ok).Should(BeTrue())
	})
})

type testHandler struct {
	Records []*log15.Record
}

func (t *testHandler) Log(r *log15.Record) error {
	t.Records = append(t.Records, r)
	return nil
}
