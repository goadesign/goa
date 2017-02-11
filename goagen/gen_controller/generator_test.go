package gencontroller_test

import (
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/gen_controller"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewGenerator", func() {
	var generator *gencontroller.Generator

	var args = struct {
		api       *design.APIDefinition
		outDir    string
		designPkg string
		appPkg    string
		force     bool
		pkg       string
		resource  string
		noExample bool
	}{
		api: &design.APIDefinition{
			Name: "test api",
		},
		outDir:    "out_dir",
		designPkg: "design",
		appPkg:    "app",
		pkg:       "controller",
		resource:  "controller",
		force:     false,
	}

	Context("with options all options set", func() {
		BeforeEach(func() {

			generator = gencontroller.NewGenerator(
				gencontroller.API(args.api),
				gencontroller.OutDir(args.outDir),
				gencontroller.DesignPkg(args.designPkg),
				gencontroller.AppPkg(args.appPkg),
				gencontroller.Pkg(args.pkg),
				gencontroller.Resource(args.resource),
				gencontroller.Force(args.force),
			)
		})

		It("has all public properties set with expected value", func() {
			Ω(generator).ShouldNot(BeNil())
			Ω(generator.API.Name).Should(Equal(args.api.Name))
			Ω(generator.OutDir).Should(Equal(args.outDir))
			Ω(generator.DesignPkg).Should(Equal(args.designPkg))
			Ω(generator.AppPkg).Should(Equal(args.appPkg))
			Ω(generator.Pkg).Should(Equal(args.pkg))
			Ω(generator.Resource).Should(Equal(args.resource))
			Ω(generator.Force).Should(Equal(args.force))
		})

	})
})
