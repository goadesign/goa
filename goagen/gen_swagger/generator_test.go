package genswagger_test

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/gen_swagger"
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
