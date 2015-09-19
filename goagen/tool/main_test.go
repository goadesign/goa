package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/goagen/app"
)

var _ = Describe("goagen", func() {
	Context("with a valid command line", func() {
		const cmdName = "app"

		BeforeEach(func() {
			os.Args = []string{"goagen", cmdName, "--target", "target", "-o", ".", "--design", "d"}
		})

		It("command returns the correct command", func() {
			cmd := command()
			Ω(cmd).ShouldNot(BeNil())
			Ω(cmd.Name()).Should(Equal(cmdName))
		})
	})
})

var _ = Describe("Application code generation", func() {
	var designCode string

	var gen *app.Generator
	var outDir string
	var designDir string
	var designPackagePath string
	var files []string
	var genErr error

	BeforeEach(func() {
		var err error
		outDir, err = ioutil.TempDir("", "")
		Ω(err).ShouldNot(HaveOccurred())
		basePath := filepath.Join(os.Getenv("GOPATH"), "src")
		designDir, err = ioutil.TempDir(basePath, "goagentest")
		designPackagePath = filepath.Base(designDir)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + outDir, "--design=" + designPackagePath}
	})

	JustBeforeEach(func() {
		f, err := os.Create(filepath.Join(designDir, "design.go"))
		Ω(err).ShouldNot(HaveOccurred())
		f.WriteString(designCode)
		f.Close()
		gen, err = app.NewGenerator()
		Ω(err).ShouldNot(HaveOccurred())
		files, genErr = gen.Generate()
	})

	AfterEach(func() {
		os.RemoveAll(outDir)
		os.RemoveAll(designDir)
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			designCode = wrapCode(`
			design.Design = &design.APIDefinition{
				Name:        "test api",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
			}`)
		})

		It("generates correct empty files", func() {
			// TBD: CALL GOAGEN TOOL
			//Ω(genErr).Should(BeNil())
			//Ω(files).Should(HaveLen(3))
		})
	})
})

func wrapCode(code string) string {
	return `package userDesign

import "github.com/raphael/goa/design"

` + code
}
