package middleware_test

import (
	"net/http"
	"strings"
	"time"

	"context"

	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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
			return service.Send(ctx, 200, "ok")
		}
		t := middleware.Timeout(time.Duration(1))(h)
		err = t(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
		_, ok := newCtx.Deadline()
		Ω(ok).Should(BeTrue())
	})
})
