package genschema_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_schema"
	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	var files []string
	var genErr error
	var workspace *codegen.Workspace
	var testPkg *codegen.Package

	BeforeEach(func() {
		var err error
		workspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		testPkg, err = workspace.NewPackage("schematest")
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + testPkg.Abs(), "--design=foo", "--version=" + version.String()}
	})

	JustBeforeEach(func() {
		files, genErr = genschema.Generate()
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			dslengine.Reset()
			apidsl.API("test api", func() {
				apidsl.Title("dummy API with no resource")
				apidsl.Description("I told you it's dummy")
			})
			dslengine.Run()
		})

		It("generates a dummy schema", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(2))
			content, err := ioutil.ReadFile(filepath.Join(testPkg.Abs(), "schema", "schema.json"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically("==", 1))
			var s genschema.JSONSchema
			err = json.Unmarshal(content, &s)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

var _ = Describe("NewGenerator", func() {
	var generator *genschema.Generator

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

			generator = genschema.NewGenerator(
				genschema.API(args.api),
				genschema.OutDir(args.outDir),
			)
		})

		It("has all public properties set with expected value", func() {
			Ω(generator).ShouldNot(BeNil())
			Ω(generator.API.Name).Should(Equal(args.api.Name))
			Ω(generator.OutDir).Should(Equal(args.outDir))
		})
	})
})
