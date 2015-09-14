package bootstrap_test

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/goagen/bootstrap"
)

var _ = Describe("Run", func() {
	var compiledFiles []string
	var compileError error

	var factory string
	var debug bool
	var designPackage string
	var designPackageDir string
	var designPackageSource string

	var meta *bootstrap.MetaGenerator

	BeforeEach(func() {
		factory = ""
		debug = false
		designPackage = ""
		designPackageDir = ""
		designPackageSource = ""
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
		meta = &bootstrap.MetaGenerator{
			Factory:       factory,
			Debug:         debug,
			DesignPackage: designPackage,
		}
		compiledFiles, compileError = meta.Generate()
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
			Ω(compileError).Should(MatchError(`invalid $GOPATH value "` + invalidPath + `"`))
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

	Context("with no factory specified", func() {
		BeforeEach(func() {
			designPackage = "github.com/raphael/goa" // must be a valid package path
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError(`missing generator factory method`))
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

	Context("with design package content", func() {

		BeforeEach(func() {
			factory = "design.NewFoo"
			designPackage = "github.com/raphael/goa/testgoagoagen"
			designPackageDir = filepath.Join(os.Getenv("GOPATH"), "src", designPackage)
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
func NewFoo(designPath string) *Generator { return new(Generator) }

func init() { panic("kaboom") }
`

	validSource = `package design
type Generator int
func (g *Generator) Generate() {}
func NewFoo(designPath string) *Generator { return new(Generator) }
`

	validSourceTmpl = `package design
import "fmt"
type Generator int
func (g *Generator) Generate() {
	{{range .}}fmt.Println("{{.}}")
	{{end}}
}
func NewFoo(designPath string) *Generator { return new(Generator) }
`
)
