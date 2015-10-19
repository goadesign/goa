package genschema_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/codegen/gen_schema"
	"github.com/raphael/goa/design"
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
			design.Design = &design.APIDefinition{Name: "foo"}
			var err error
			gen, err = genschema.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})

		It("instantiates a generator even if Design is not initialized", func() {
			design.Design = nil
			var err error
			gen, err = genschema.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
		})
	})
})

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/raphael/goa/codegen/gen_schema/goatest"

	var gen *genschema.Generator
	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		gopath := os.Getenv("GOPATH")
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"codegen", "--out=" + outDir, "--design=foo"}
	})

	JustBeforeEach(func() {
		var err error
		gen, err = genschema.NewGenerator()
		Ω(err).ShouldNot(HaveOccurred())
		files, genErr = gen.Generate(design.Design)
	})

	AfterEach(func() {
		os.RemoveAll(outDir)
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				Name:        "test api",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
			}
		})

		It("generates a dummy schema", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(1))
			content, err := ioutil.ReadFile(filepath.Join(genschema.JSONSchemaDir(), "schema.json"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically("==", 1))
			var s genschema.JSONSchema
			err = json.Unmarshal(content, &s)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})
