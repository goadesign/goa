package genmain_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_main"
)

var _ = Describe("NewGenerator", func() {
	var gen *genmain.Generator

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
			gen, err = genmain.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})

		It("instantiates a generator even if Design is not initialized", func() {
			dsl.InitDesign()
			var err error
			gen, err = genmain.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})
	})
})

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_main/goatest"

	var gen *genmain.Generator
	var outDir string
	var files []string
	var genErr error

	var oldCommand string

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"codegen", "--out=" + outDir, "--design=foo"}
		oldCommand = codegen.CommandName
		codegen.CommandName = "app"
	})

	JustBeforeEach(func() {
		var err error
		gen, err = genmain.NewGenerator()
		Ω(err).ShouldNot(HaveOccurred())
		files, genErr = gen.Generate(design.Design)
	})

	AfterEach(func() {
		codegen.CommandName = oldCommand
		os.RemoveAll(outDir)
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

		It("generates a dummy app", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(1))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "main.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 16))
			_, err = gexec.Build(testgenPackagePath)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})
