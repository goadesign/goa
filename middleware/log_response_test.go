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
