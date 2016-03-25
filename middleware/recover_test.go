package middleware_test

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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
