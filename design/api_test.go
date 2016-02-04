package design_test

import (
	"github.com/goadesign/goa/design"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SupportingPackages", func() {
	var enc *design.EncodingDefinition
	var pkgs map[string][]string

	var packagePath string
	var mimeTypes []string

	BeforeEach(func() {
		packagePath = "github.com/goadesign/goa"
		mimeTypes = []string{"application/json"}
	})

	JustBeforeEach(func() {
		enc = &design.EncodingDefinition{
			PackagePath: packagePath,
			MIMETypes:   mimeTypes,
		}
		pkgs = enc.SupportingPackages()
	})

	Context("with a valid definition with one media type and a package path", func() {
		It("returns a map with one element", func() {
			Ω(pkgs).Should(HaveLen(1))
			Ω(pkgs).Should(HaveKeyWithValue(packagePath, mimeTypes))
		})

	})

	Context("with a valid definition with one media type and no package path", func() {
		BeforeEach(func() {
			packagePath = "json"
		})

		It("returns a map with one element", func() {
			Ω(pkgs).Should(HaveLen(1))
			Ω(pkgs).Should(HaveKeyWithValue(packagePath, mimeTypes))
		})
	})

	Context("with mime types using different known encoder packages", func() {
		BeforeEach(func() {
			packagePath = ""
			mimeTypes = []string{"application/xml", "application/msgpack"}
		})

		It("returns all encoders", func() {
			Ω(pkgs).Should(HaveLen(2))
			Ω(pkgs).Should(HaveKeyWithValue("xml", []string{"application/xml"}))
			Ω(pkgs).Should(HaveKeyWithValue("github.com/goadesign/encoding/msgpack", []string{"application/msgpack"}))
		})
	})

	Context("with a unknown mime type and a package path", func() {
		BeforeEach(func() {
			packagePath = ""
			mimeTypes = []string{"application/vmd.custom"}
		})

		It("returns nil", func() {
			Ω(pkgs).Should(BeNil())
		})
	})

	Context("with known media types and a unknown mime type and a package path", func() {
		BeforeEach(func() {
			packagePath = ""
			mimeTypes = []string{"application/json", "application/xml", "application/vmd.custom"}
		})

		It("returns nil", func() {
			Ω(pkgs).Should(BeNil())
		})
	})
})
