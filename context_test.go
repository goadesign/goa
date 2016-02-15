package goa_test

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/goadesign/goa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResponseData", func() {
	var data *goa.ResponseData
	var rw http.ResponseWriter
	var req *http.Request

	BeforeEach(func() {
		app := goa.New("test")
		var err error
		req, err = http.NewRequest("GET", "google.com", nil)
		Ω(err).ShouldNot(HaveOccurred())
		rw = &TestResponseWriter{Status: 42}
		ctx := goa.NewContext(context.Background(), app, rw, req)
		data = goa.Response(ctx)
	})

	Context("SwitchWriter", func() {
		var rwo http.ResponseWriter

		BeforeEach(func() {
			rwo = data.SwitchWriter(rw)
		})

		It("sets the response writer and returns the previous one", func() {
			Ω(rwo).Should(BeNil())
			rwo = data.SwitchWriter(&TestResponseWriter{Status: 43})
			Ω(rwo).ShouldNot(BeNil())
			Ω(rwo).Should(BeAssignableToTypeOf(&TestResponseWriter{}))
			trw := rwo.(*TestResponseWriter)
			Ω(trw.Status).Should(Equal(42))
		})
	})
})
