package middleware_test

import (
	"net/http"
	"regexp"
	"strings"

	"context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RequireHeader", func() {
	var ctx context.Context
	var req *http.Request
	var rw http.ResponseWriter
	var service *goa.Service
	headerName := "Some-Header"

	BeforeEach(func() {
		var err error
		service = newService(nil)
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
			return service.Send(ctx, http.StatusOK, "ok")
		}
		t := middleware.RequireHeader(
			service,
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
			service,
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
			service,
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
			return service.Send(ctx, http.StatusOK, "ok")
		}
		t := middleware.RequireHeader(
			service,
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
			service,
			nil,
			headerName,
			regexp.MustCompile("^some value$"),
			http.StatusNotFound)(h)
		err := t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(goa.ContextResponse(ctx).Status).Should(Equal(http.StatusNotFound))
	})
})
