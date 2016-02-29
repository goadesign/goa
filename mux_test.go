package goa_test

import (
	"net/http"

	"github.com/goadesign/goa"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PathSelectVersionFunc", func() {
	var pattern, param string
	var request *http.Request

	var fn goa.SelectVersionFunc
	var version string

	JustBeforeEach(func() {
		fn = goa.PathSelectVersionFunc(pattern, param)
		version = fn(request)
	})

	Context("using path versioning", func() {
		BeforeEach(func() {
			pattern = "/:version/"
			param = "version"
		})

		Context("and a versioned request", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("GET", "/v1/foo", nil)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("routes to the versioned controller", func() {
				Ω(version).Should(Equal("v1"))
			})
		})
	})

})
