package genswagger_test

import (
	"github.com/goadesign/goa/design"
	genswagger "github.com/goadesign/goa/goagen/gen_swagger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewGenerator", func() {
	var generator *genswagger.Generator

	var args = struct {
		api    *design.APIDefinition
		outDir string
	}{
		api: &design.APIDefinition{
			Name: "test api",
		},
		outDir: "out_dir",
	}

	Context("with options all options set", func() {
		BeforeEach(func() {

			generator = genswagger.NewGenerator(
				genswagger.API(args.api),
				genswagger.OutDir(args.outDir),
			)
		})

		It("has all public properties set with expected value", func() {
			Ω(generator).ShouldNot(BeNil())
			Ω(generator.API.Name).Should(Equal(args.api.Name))
			Ω(generator.OutDir).Should(Equal(args.outDir))
		})
	})
})

var _ = Describe("jsonToYAML", func() {
	var rawYAML []byte
	var err error

	Context("with JSON which has a number type value", func() {
		BeforeEach(func() {

			rawYAML, _ = genswagger.JSONToYAML(
				[]byte(`{"id":1234567}`),
			)
		})

		It("converts JSON to YAML and keeps right number type", func() {
			Ω(string(rawYAML)).Should(Equal("id: 1234567\n"))
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})
