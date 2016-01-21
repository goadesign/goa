package genschema_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_schema"
)

var _ = Describe("NewGenerator", func() {
	var gen *genschema.Generator

	Context("with dummy command line flags", func() {
		BeforeEach(func() {
			os.Args = []string{"codegen", "--out=_foo", "--design=bar"}
		})

		AfterEach(func() {
			os.RemoveAll("_foo")
		})

		It("instantiates a generator", func() {
			design.Design = &design.APIDefinition{
				APIVersionDefinition: &design.APIVersionDefinition{
					Name: "foo",
				},
			}
			var err error
			gen, err = genschema.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})

		It("instantiates a generator even if Design is not initialized", func() {
			dsl.InitDesign()
			var err error
			gen, err = genschema.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})
	})
})

var _ = Describe("Generate", func() {
	var gen *genschema.Generator
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
		os.Args = []string{"codegen", "--out=" + testPkg.Abs(), "--design=foo"}
	})

	JustBeforeEach(func() {
		var err error
		gen, err = genschema.NewGenerator()
		Ω(err).ShouldNot(HaveOccurred())
		files, genErr = gen.Generate(design.Design)
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				APIVersionDefinition: &design.APIVersionDefinition{
					Name:        "test api",
					Title:       "dummy API with no resource",
					Description: "I told you it's dummy",
				},
			}
		})

		It("generates a dummy schema", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(3))
			content, err := ioutil.ReadFile(filepath.Join(genschema.JSONSchemaDir(), "schema.json"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically("==", 1))
			var s genschema.JSONSchema
			err = json.Unmarshal(content, &s)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})
