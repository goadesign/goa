package version_test

import (
	"strconv"

	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("version", func() {
	var ver string

	JustBeforeEach(func() {
		ver = version.String()
	})

	Context("with the default build number", func() {
		It("should be properly formatted", func() {
			Ω(ver).Should(HavePrefix("v"))
		})
	})

	Context("checking compatibility", func() {
		var otherVersion string
		var compatible bool
		var compErr error

		JustBeforeEach(func() {
			compatible, compErr = version.Compatible(otherVersion)
		})

		Context("with a version with identical major", func() {
			BeforeEach(func() {
				otherVersion = "v" + strconv.Itoa(version.Major) + ".12.13"
			})
			It("returns true", func() {
				Ω(compErr).ShouldNot(HaveOccurred())
				Ω(compatible).Should(BeTrue())
			})
		})

		Context("with a version with different major", func() {
			BeforeEach(func() {
				otherVersion = "v99999121299999.1.0"
			})
			It("returns false", func() {
				Ω(compErr).ShouldNot(HaveOccurred())
				Ω(compatible).Should(BeFalse())
			})
		})

		Context("with a non version string", func() {
			BeforeEach(func() {
				otherVersion = "v99999121299999.2"
			})
			It("returns an error", func() {
				Ω(compErr).Should(HaveOccurred())
				Ω(compatible).Should(BeFalse())
			})
		})
	})

})
