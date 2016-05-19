package middleware_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ErrorHandler", func() {
	var service *goa.Service
	var h goa.Handler
	var verbose bool

	var rw *testResponseWriter

	BeforeEach(func() {
		service = nil
		h = nil
		verbose = true
		rw = nil
	})

	JustBeforeEach(func() {
		rw = newTestResponseWriter()
		eh := middleware.ErrorHandler(service, verbose)(h)
		req, err := http.NewRequest("GET", "/foo", nil)
		Ω(err).ShouldNot(HaveOccurred())
		ctx := newContext(service, rw, req, nil)
		err = eh(ctx, rw, req)
		Ω(err).ShouldNot(HaveOccurred())
	})

	Context("with a handler returning a Go error", func() {
		var requestCtx context.Context

		BeforeEach(func() {
			service = newService(nil)
			h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				requestCtx = ctx
				return errors.New("boom")
			}
		})

		It("turns Go errors into HTTP 500 responses", func() {
			Ω(rw.Status).Should(Equal(500))
			Ω(rw.ParentHeader["Content-Type"]).Should(Equal([]string{"text/plain"}))
			Ω(string(rw.Body)).Should(Equal(`"boom"` + "\n"))
		})

		Context("not verbose", func() {
			BeforeEach(func() {
				verbose = false
			})

			It("hides the error details", func() {
				var decoded goa.Error
				Ω(rw.Status).Should(Equal(500))
				Ω(rw.ParentHeader["Content-Type"]).Should(Equal([]string{goa.ErrorMediaIdentifier}))
				err := service.Decoder.Decode(&decoded, bytes.NewBuffer(rw.Body), "application/json")
				Ω(err).ShouldNot(HaveOccurred())
				msg := fmt.Sprintf("%v", *goa.ErrInternal(`Internal Server Error [zzz]`))
				msg = regexp.QuoteMeta(msg)
				msg = strings.Replace(msg, "zzz", ".+", 1)
				Ω(fmt.Sprintf("%v", decoded)).Should(MatchRegexp(msg))
			})
		})
	})

	Context("with a handler returning a goa error", func() {
		var gerr *goa.Error

		BeforeEach(func() {
			service = newService(nil)
			gerr = &goa.Error{
				Status:     418,
				Detail:     "teapot",
				Code:       "code",
				MetaValues: map[string]interface{}{"foobar": 42},
			}
			h = func(ctx context.Context, rw http.ResponseWriter, req *http.Request) error {
				return gerr
			}
		})

		It("maps goa errors to HTTP responses", func() {
			var decoded goa.Error
			Ω(rw.Status).Should(Equal(gerr.Status))
			Ω(rw.ParentHeader["Content-Type"]).Should(Equal([]string{goa.ErrorMediaIdentifier}))
			err := service.Decoder.Decode(&decoded, bytes.NewBuffer(rw.Body), "application/json")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fmt.Sprintf("%v", decoded)).Should(Equal(fmt.Sprintf("%v", *gerr)))
		})
	})
})
