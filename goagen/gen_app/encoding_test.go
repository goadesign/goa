package genapp_test

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/gen_app"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BuildEncoders", func() {
	var info []*design.EncodingDefinition
	var encoder bool

	var data []*genapp.EncoderTemplateData
	var resErr error

	BeforeEach(func() {
		info = nil
		encoder = false
	})

	JustBeforeEach(func() {
		data, resErr = genapp.BuildEncoders(info, encoder)
	})

	Context("with a single definition using a single known MIME type for encoding", func() {
		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				MIMETypes: []string{"application/json"},
				Encoder:   true,
			}
			info = append(info, simple)
			encoder = true
		})

		It("generates a map with a single entry", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			jd := data[0]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal("github.com/goadesign/goa"))
			Ω(jd.PackageName).Should(Equal("goa"))
			Ω(jd.Function).Should(Equal("NewJSONEncoder"))
			Ω(jd.MIMETypes).Should(HaveLen(1))
			Ω(jd.MIMETypes[0]).Should(Equal("application/json"))
		})
	})

	Context("with a single definition using a single known MIME type for decoding", func() {
		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				MIMETypes: []string{"application/json"},
			}
			info = append(info, simple)
			encoder = false
		})

		It("generates a map with a single entry", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			jd := data[0]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal("github.com/goadesign/goa"))
			Ω(jd.PackageName).Should(Equal("goa"))
			Ω(jd.Function).Should(Equal("NewJSONDecoder"))
			Ω(jd.MIMETypes).Should(HaveLen(1))
			Ω(jd.MIMETypes[0]).Should(Equal("application/json"))
		})
	})

	Context("with a definition using a custom decoding package", func() {
		const packagePath = "github.com/goadesign/goa/design" // Just to pick something always available
		var mimeTypes = []string{"application/vnd.custom", "application/vnd.custom2"}

		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				PackagePath: packagePath,
				Function:    "NewDecoder",
				MIMETypes:   mimeTypes,
			}
			info = append(info, simple)
		})

		It("generates a map with a single entry", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			jd := data[0]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal(packagePath))
			Ω(jd.PackageName).Should(Equal("design"))
			Ω(jd.Function).Should(Equal("NewDecoder"))
			Ω(jd.MIMETypes).Should(ConsistOf(interface{}(mimeTypes[0]), interface{}(mimeTypes[1])))
		})
	})

	Context("with a definition using a custom decoding package for a known encoding", func() {
		const packagePath = "github.com/goadesign/goa/design" // Just to pick something always available
		var mimeTypes = []string{"application/json"}

		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				PackagePath: packagePath,
				Function:    "NewDecoder",
				MIMETypes:   mimeTypes,
			}
			info = append(info, simple)
		})

		It("generates a map with a single entry using the generic decoder factory name", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			jd := data[0]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal(packagePath))
			Ω(jd.PackageName).Should(Equal("design"))
			Ω(jd.Function).Should(Equal("NewDecoder"))
		})
	})

	Context("with a definition using a custom decoding package from goadesign", func() {
		const packagePath = "github.com/goadesign/goa/encoding/gogoprotobuf"
		var mimeTypes = []string{"application/x-protobuf"}

		BeforeEach(func() {
			simple := &design.EncodingDefinition{
				PackagePath: packagePath,
				MIMETypes:   mimeTypes,
			}
			info = append(info, simple)
		})

		It("generates a map with a single entry using the generic decoder factory name", func() {
			Ω(resErr).ShouldNot(HaveOccurred())
			Ω(data).Should(HaveLen(1))
			jd := data[0]
			Ω(jd).ShouldNot(BeNil())
			Ω(jd.PackagePath).Should(Equal(packagePath))
			Ω(jd.PackageName).Should(Equal("gogoprotobuf"))
			Ω(jd.Function).Should(Equal("NewDecoder"))
		})
	})
})
