package middleware_test

import (
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequestID", func() {
	const reqID = "request id"
	var ctx context.Context
	var rw http.ResponseWriter
	var req *http.Request
	var params url.Values
	var service *goa.Service

	BeforeEach(func() {
		service = newService(nil)

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
			return service.Send(ctx, 200, "ok")
		}
		rg := middleware.RequestID()(h)
		Ω(rg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(middleware.ContextRequestID(newCtx)).Should(Equal(reqID))
	})
})
