package middleware_test

import (
	"net/http"
	"net/url"
	"strings"

	"context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogRequest", func() {
	var ctx context.Context
	var rw *testResponseWriter
	var req *http.Request
	var params url.Values
	var logger *testLogger
	var service *goa.Service

	payload := map[string]interface{}{"payload": 42}

	BeforeEach(func() {
		logger = new(testLogger)
		service = newService(logger)

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
		// Add Action name to the context to make sure we log it properly.
		ctx = goa.WithAction(ctx, "goo")

		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			return service.Send(ctx, 200, "ok")
		}
		lg := middleware.LogRequest(true)(h)
		Ω(lg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(logger.InfoEntries).Should(HaveLen(4))

		Ω(logger.InfoEntries[0].Data).Should(HaveLen(10))
		Ω(logger.InfoEntries[0].Data[0]).Should(Equal("req_id"))
		Ω(logger.InfoEntries[0].Data[2]).Should(Equal("POST"))
		Ω(logger.InfoEntries[0].Data[3]).Should(Equal("/goo?param=value"))

		Ω(logger.InfoEntries[1].Data).Should(HaveLen(4))
		Ω(logger.InfoEntries[1].Data[0]).Should(Equal("req_id"))
		Ω(logger.InfoEntries[1].Data[2]).Should(Equal("query"))
		Ω(logger.InfoEntries[1].Data[3]).Should(Equal("value"))

		Ω(logger.InfoEntries[2].Data).Should(HaveLen(4))
		Ω(logger.InfoEntries[2].Data[0]).Should(Equal("req_id"))
		Ω(logger.InfoEntries[2].Data[2]).Should(Equal("payload"))
		Ω(logger.InfoEntries[2].Data[3]).Should(Equal(42))

		Ω(logger.InfoEntries[3].Data).Should(HaveLen(12))
		Ω(logger.InfoEntries[3].Data[0]).Should(Equal("req_id"))
		Ω(logger.InfoEntries[3].Data[2]).Should(Equal("status"))
		Ω(logger.InfoEntries[3].Data[3]).Should(Equal(200))
		Ω(logger.InfoEntries[3].Data[4]).Should(Equal("bytes"))
		Ω(logger.InfoEntries[3].Data[5]).Should(Equal(5))
		Ω(logger.InfoEntries[3].Data[6]).Should(Equal("time"))
		Ω(logger.InfoEntries[3].Data[8]).Should(Equal("ctrl"))
		Ω(logger.InfoEntries[3].Data[9]).Should(Equal("test"))
		Ω(logger.InfoEntries[3].Data[10]).Should(Equal("action"))
		Ω(logger.InfoEntries[3].Data[11]).Should(Equal("goo"))
	})

	It("logs error codes", func() {
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			return goa.MissingParamError("foo")
		}
		rw.ParentHeader = make(http.Header)
		lg := middleware.LogRequest(false)(middleware.ErrorHandler(service, false)(h))
		Ω(lg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(logger.InfoEntries).Should(HaveLen(2))
		Ω(logger.InfoEntries[0].Data).Should(HaveLen(10))
		Ω(logger.InfoEntries[0].Data[0]).Should(Equal("req_id"))
		Ω(logger.InfoEntries[0].Data[2]).Should(Equal("POST"))
		Ω(logger.InfoEntries[0].Data[3]).Should(Equal("/goo?param=value"))

		Ω(logger.InfoEntries[1].Data).Should(HaveLen(14))
		Ω(logger.InfoEntries[1].Data[0]).Should(Equal("req_id"))
		Ω(logger.InfoEntries[1].Data[2]).Should(Equal("status"))
		Ω(logger.InfoEntries[1].Data[3]).Should(Equal(400))
		Ω(logger.InfoEntries[1].Data[4]).Should(Equal("error"))
		Ω(logger.InfoEntries[1].Data[5]).Should(HaveLen(8)) // Error ID
		Ω(logger.InfoEntries[1].Data[6]).Should(Equal("bytes"))
		Ω(logger.InfoEntries[1].Data[7]).Should(Equal(124))
		Ω(logger.InfoEntries[1].Data[8]).Should(Equal("time"))
		Ω(logger.InfoEntries[1].Data[10]).Should(Equal("ctrl"))
		Ω(logger.InfoEntries[1].Data[11]).Should(Equal("test"))
		Ω(logger.InfoEntries[1].Data[12]).Should(Equal("action"))
		Ω(logger.InfoEntries[1].Data[13]).Should(Equal("<unknown>"))
	})
})
