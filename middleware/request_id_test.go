package middleware_test

import (
	"net/http"
	"net/url"

	"context"

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
		req.Header.Set(middleware.RequestIDHeader, reqID)
		rw = new(testResponseWriter)
		params = url.Values{"query": []string{"value"}}
		service.Encoder.Register(goa.NewJSONEncoder, "*/*")
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

	It("truncates request ID when it exceeds a default limit", func() {
		var newCtx context.Context
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newCtx = ctx
			return service.Send(ctx, 200, "ok")
		}
		tooLong := makeRequestID(2 * middleware.DefaultRequestIDLengthLimit)
		expected := makeRequestID(middleware.DefaultRequestIDLengthLimit)
		req.Header.Set(middleware.RequestIDHeader, tooLong)
		rg := middleware.RequestID()(h)
		Ω(rg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(middleware.ContextRequestID(newCtx)).Should(Equal(expected))
	})

	It("sets the request ID in the context for a custom header and limit", func() {
		var newCtx context.Context
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newCtx = ctx
			return service.Send(ctx, 200, "ok")
		}
		req.Header.Set("Foo", "abcdefghij")
		rg := middleware.RequestIDWithHeaderAndLengthLimit("Foo", 7)(h)
		Ω(rg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(middleware.ContextRequestID(newCtx)).Should(Equal("abcdefg"))
	})

	It("allows any request ID when length limit is negative", func() {
		var newCtx context.Context
		h := func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
			newCtx = ctx
			return service.Send(ctx, 200, "ok")
		}
		original := makeRequestID(2 * middleware.DefaultRequestIDLengthLimit)
		req.Header.Set(middleware.RequestIDHeader, string(original))
		rg := middleware.RequestIDWithHeaderAndLengthLimit(middleware.RequestIDHeader, -1)(h)
		Ω(rg(ctx, rw, req)).ShouldNot(HaveOccurred())
		Ω(middleware.ContextRequestID(newCtx)).Should(Equal(string(original)))
	})

})

func makeRequestID(length int) string {
	buffer := make([]byte, length)
	for i := range buffer {
		buffer[i] = 'x'
	}
	return string(buffer)
}
