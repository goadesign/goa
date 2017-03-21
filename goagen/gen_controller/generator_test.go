package gencontroller_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_controller"
	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	var workspace *codegen.Workspace
	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		var err error
		workspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		outDir, err = ioutil.TempDir(workspace.Path, "")
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + outDir, "--design=foo", "--version=" + version.String()}
	})

	JustBeforeEach(func() {
		files, genErr = gencontroller.Generate()
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("with a simple API", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				Name: "testapi",
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
								},
							},
						},
					},
				},
			}
			fooRes := design.Design.Resources["foo"]
			showAct := fooRes.Actions["show"]
			showAct.Parent = fooRes
			showAct.Routes[0].Parent = showAct
		})

		It("generates a simple controller", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(1))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 16))
		})
	})
})

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
