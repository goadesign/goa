package meta_test

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/goagen"
	"github.com/raphael/goa/goagen/meta"
)

var _ = Describe("Run", func() {
	var compiledFiles []string
	var compileError error

	var factory string
	var debug bool
	var outputDir string
	var designPackage string
	var designPackageDir string
	var designPackageSource string

	var m *meta.Generator

	BeforeEach(func() {
		factory = ""
		debug = false
		outputDir = "/tmp"
		designPackage = "github.com/raphael/goa/testgoagoagen"
		designPackageSource = "foo"
		designPackageDir = filepath.Join(os.Getenv("GOPATH"), "src", designPackage)
		compiledFiles = nil
		compileError = nil
	})

	JustBeforeEach(func() {
		if designPackageDir != "" && designPackageSource != "" {
			err := os.MkdirAll(designPackageDir, 0777)
			Ω(err).ShouldNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(designPackageDir, "design.go"), []byte(designPackageSource), 0655)
			Ω(err).ShouldNot(HaveOccurred())
		}
		m = &meta.Generator{
			Factory: factory,
		}
		goagen.Debug = debug
		goagen.OutputDir = outputDir
		goagen.DesignPackagePath = designPackage
		compiledFiles, compileError = m.Generate()
	})

	AfterEach(func() {
		if designPackageDir != "" {
			os.RemoveAll(designPackageDir)
		}
	})

	Context("with no GOPATH environment variable", func() {
		var gopath string

		BeforeEach(func() {
			gopath = os.Getenv("GOPATH")
			os.Setenv("GOPATH", "")
		})

		AfterEach(func() {
			os.Setenv("GOPATH", gopath)
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("$GOPATH not defined"))
		})
	})

	Context("with an invalid GOPATH environment variable", func() {
		var gopath string
		const invalidPath = "DOES NOT EXIST"

		BeforeEach(func() {
			gopath = os.Getenv("GOPATH")
			os.Setenv("GOPATH", invalidPath)
		})

		AfterEach(func() {
			os.Setenv("GOPATH", gopath)
		})

		It("fails with a useful error message", func() {
			msg := fmt.Sprintf(`cannot find design package at path "%s/src/%s"`, invalidPath, designPackage)
			Ω(compileError).Should(MatchError(msg))
		})

	})

	Context("with an invalid design package path", func() {
		const invalidDesignPackage = "foobar"

		BeforeEach(func() {
			designPackage = invalidDesignPackage
		})

		It("fails with a useful error message", func() {
			path := filepath.Join(os.Getenv("GOPATH"), "src", designPackage)
			Ω(compileError).Should(MatchError(`cannot find design package at path "` + path + `"`))
		})
	})

	Context("with no go compiler in PATH", func() {
		var pathEnv string
		const invalidPath = "/foobar"

		BeforeEach(func() {
			factory = "design.NewFoo"
			pathEnv = os.Getenv("PATH")
			os.Setenv("PATH", invalidPath)
		})

		AfterEach(func() {
			os.Setenv("PATH", pathEnv)
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError(`failed to find a go compiler, looked in "` + os.Getenv("PATH") + `"`))
		})
	})

	Context("with no output directory specified", func() {
		BeforeEach(func() {
			factory = "design.NewFoo"
			outputDir = ""
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("missing output directory specification"))
		})
	})

	Context("with no design package path specified", func() {
		BeforeEach(func() {
			factory = "design.NewFoo"
			designPackage = ""
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("missing design package path specification"))
		})
	})

	Context("with design package content", func() {

		BeforeEach(func() {
			factory = "design.NewFoo"
			outputDir = "/tmp"
		})

		Context("that is not valid Go code", func() {
			BeforeEach(func() {
				designPackageSource = invalidSource
			})

			It("fails with a useful error message", func() {
				Ω(compileError.Error()).Should(ContainSubstring("unexpected go"))
			})
		})

		Context("whose code blows up", func() {
			BeforeEach(func() {
				designPackageSource = panickySource
			})

			It("fails with a useful error message", func() {
				Ω(compileError.Error()).Should(ContainSubstring("panic: kaboom"))
			})
		})

		Context("with valid code", func() {
			BeforeEach(func() {
				designPackageSource = validSource
			})

			It("successfully runs", func() {
				Ω(compileError).ShouldNot(HaveOccurred())
			})
		})

		Context("with code that returns generated file paths", func() {
			var filePaths = []string{"foo", "bar"}

			BeforeEach(func() {
				var b bytes.Buffer
				tmpl, err := template.New("source").Parse(validSourceTmpl)
				Ω(err).ShouldNot(HaveOccurred())
				err = tmpl.Execute(&b, filePaths)
				Ω(err).ShouldNot(HaveOccurred())
				designPackageSource = b.String()
			})

			It("returns the paths", func() {
				Ω(compileError).ShouldNot(HaveOccurred())
				Ω(compiledFiles).Should(Equal(filePaths))
			})
		})
	})
})

const (
	invalidSource = `package design
invalid go code
`

	panickySource = `package design
type Generator int
func (g *Generator) Generate() {}
func NewFoo(designPath string) (*Generator, error) { return new(Generator), nil }

func init() { panic("kaboom") }
`

	validSource = `package design
type Generator int
func (g *Generator) Generate() {}
func NewFoo(designPath string) (*Generator, error) { return new(Generator), nil }
`

	validSourceTmpl = `package design
import "fmt"
type Generator int
func (g *Generator) Generate() {
	{{range .}}fmt.Println("{{.}}")
	{{end}}
}
func NewFoo(designPath string) (*Generator, error) { return new(Generator), nil }
`
)
