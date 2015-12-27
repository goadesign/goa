package genjs_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
	"github.com/raphael/goa/goagen/gen_js"
)

var _ = Describe("NewGenerator", func() {
	var gen *genjs.Generator

	Context("with dummy command line flags", func() {
		BeforeEach(func() {
			os.Args = []string{"codegen", "--out=_foo", "--design=bar"}
		})

		AfterEach(func() {
			os.RemoveAll("_foo")
		})

		It("instantiates a generator", func() {
			design.Design = &design.APIDefinition{Name: "foo"}
			var err error
			gen, err = genjs.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})

		It("instantiates a generator even if Design is not initialized", func() {
			design.Design = nil
			var err error
			gen, err = genjs.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})
	})
})

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/raphael/goa/goagen/gen_js/test_"

	var gen *genjs.Generator
	var outDir string
	var files []string
	var genErr error

	var oldCommand string

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"codegen", "--out=" + outDir, "--design=foo", "--host=baz"}
		oldCommand = codegen.CommandName
		codegen.CommandName = "app"
	})

	JustBeforeEach(func() {
		var err error
		gen, err = genjs.NewGenerator()
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
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
			}
		})

		It("generates a dummy js", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(2))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "js", "client.js"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 13))
		})
	})
})
