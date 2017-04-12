package middleware_test

import (
	"fmt"
	"net/http"

	"context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Recover", func() {
	var h goa.Handler
	var err error

	JustBeforeEach(func() {
		rg := middleware.Recover()(h)
		err = rg(nil, nil, nil)
	})

	Context("with a handler that panics with a string", func() {
		BeforeEach(func() {
			h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				panic("boom")
			}
		})

		It("creates an error from the panic message", func() {
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(HavePrefix("panic: boom\n"))
		})
	})

	Context("with a handler that panics with an error", func() {
		BeforeEach(func() {
			h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				panic(fmt.Errorf("boom"))
			}
		})

		It("creates an error from the panic error message", func() {
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(HavePrefix("panic: boom\n"))
		})
	})

	Context("with a handler that panics with something else", func() {
		BeforeEach(func() {
			h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				panic(42)
			}
		})

		It("creates a generic error message", func() {
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(HavePrefix("unknown panic\n"))
		})
	})
})
