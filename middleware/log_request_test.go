package middleware_test

import (
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogRequest", func() {
	var ctx context.Context
	var rw http.ResponseWriter
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
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			return service.Send(ctx, 200, "ok")
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
