package goa_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa"
)

var _ = Describe("PathSelectVersionFunc", func() {
	var pattern, zeroVersion string
	var request *http.Request

	var fn goa.SelectVersionFunc
	var version string

	JustBeforeEach(func() {
		fn = goa.PathSelectVersionFunc(pattern, zeroVersion)
		version = fn(request)
	})

	Context("using the default settings", func() {
		BeforeEach(func() {
			pattern = "/:version/"
			zeroVersion = "api"
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

		Context("and an unversioned request", func() {
			BeforeEach(func() {
				var err error
				request, err = http.NewRequest("GET", "/api/foo", nil)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("routes to the unversioned controller", func() {
				Ω(version).Should(Equal(""))
			})
		})
	})

})
