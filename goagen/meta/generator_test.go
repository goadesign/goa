package meta_test

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run", func() {
	var compiledFiles []string
	var compileError error
	var outputWorkspace *codegen.Workspace
	var designWorkspace *codegen.Workspace

	var genfunc string
	var outputDir string
	var designPkgPath, setDesignPkgPath string
	var designPackageSource string
	var customFlags []string

	var m *meta.Generator

	BeforeEach(func() {
		genfunc = ""
		designPkgPath = "design"
		setDesignPkgPath = designPkgPath
		designPackageSource = "package design"
		var err error
		outputWorkspace, err = codegen.NewWorkspace("output")
		p, err := outputWorkspace.NewPackage("testOutput")
		Ω(err).ShouldNot(HaveOccurred())
		outputDir = p.Abs()
		designWorkspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		compiledFiles = nil
		compileError = nil
		customFlags = []string{"--custom=arg"}
	})

	JustBeforeEach(func() {
		if designPkgPath != "" {
			designPackage, err := designWorkspace.NewPackage(designPkgPath)
			Ω(err).ShouldNot(HaveOccurred())
			if designPackageSource != "" {
				file := designPackage.CreateSourceFile("design.go")
				err = ioutil.WriteFile(file.Abs(), []byte(designPackageSource), 0655)
				Ω(err).ShouldNot(HaveOccurred())
			}
		}
		m = &meta.Generator{
			Genfunc:       genfunc,
			Imports:       []*codegen.ImportSpec{codegen.SimpleImport(designPkgPath)},
			OutDir:        outputDir,
			CustomFlags:   customFlags,
			DesignPkgPath: setDesignPkgPath,
		}
		compiledFiles, compileError = m.Generate()
	})

	AfterEach(func() {
		designWorkspace.Delete()
		outputWorkspace.Delete()
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
			Ω(compileError).Should(MatchError("GOPATH not set"))
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
			Ω(compileError).Should(MatchError(HavePrefix(`invalid design package import path: cannot find package "design" in any of:`)))
			Ω(compileError).Should(MatchError(HaveSuffix(filepath.Join(invalidPath, "src", "design") + " (from $GOPATH)")))
		})

	})

	Context("with an invalid design package path", func() {
		const invalidDesignPackage = "foobar"

		BeforeEach(func() {
			setDesignPkgPath = invalidDesignPackage
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError(HavePrefix("invalid design package import path: cannot find package")))
			Ω(compileError).Should(MatchError(ContainSubstring(invalidDesignPackage)))
		})
	})

	Context("with no go compiler in PATH", func() {
		var pathEnv string
		const invalidPath = "/foobar"

		BeforeEach(func() {
			genfunc = "foo.Generate"
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
			genfunc = "foo.Generate"
			outputDir = ""
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("missing output directory flag"))
		})
	})

	Context("with no design package path specified", func() {
		BeforeEach(func() {
			genfunc = "foo.Generate"
			outputDir = ""
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("missing output directory flag"))
		})
	})

	Context("with no design package path specified", func() {
		BeforeEach(func() {
			genfunc = "foo.Generate"
			designPkgPath = ""
			setDesignPkgPath = ""
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("missing design package flag"))
		})
	})

	Context("with design package content", func() {

		BeforeEach(func() {
			genfunc = "foo.Generate"
			outputDir = os.TempDir()
		})

		Context("that is not valid Go code", func() {
			BeforeEach(func() {
				designPackageSource = invalidSource
			})

			It("fails with a useful error message", func() {
				Ω(compileError.Error()).Should(ContainSubstring("syntax error"))
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

			Context("with a comma separated list of path in GOPATH", func() {
				var gopath string
				BeforeEach(func() {
					gopath = os.Getenv("GOPATH")
					os.Setenv("GOPATH", fmt.Sprintf("%s%c%s", gopath, os.PathListSeparator, os.TempDir()))
				})

				AfterEach(func() {
					os.Setenv("GOPATH", gopath)
				})

				It("successfull runs", func() {
					Ω(compileError).ShouldNot(HaveOccurred())
				})
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
		Context("with code that uses custom flags", func() {
			BeforeEach(func() {
				var b bytes.Buffer
				tmpl, err := template.New("source").Parse(validSourceTmplWithCustomFlags)
				Ω(err).ShouldNot(HaveOccurred())
				err = tmpl.Execute(&b, "--custom=arg")
				Ω(err).ShouldNot(HaveOccurred())
				designPackageSource = b.String()

			})

			It("returns no error", func() {
				Ω(compileError).ShouldNot(HaveOccurred())
			})
		})
	})
})

const (
	invalidSource = `package foo
invalid go code
`

	panickySource = `package foo
func Generate() ([]string, error) {
	return nil, nil
}

func init() { panic("kaboom") }
`

	validSource = `package foo
func Generate() ([]string, error) {
	return nil, nil
}
`

	validSourceTmpl = `package foo
import "fmt"
func Generate() ([]string, error) {
	{{range .}}fmt.Println("{{.}}")
	{{end}}
	return nil, nil
}
`

	validSourceTmplWithCustomFlags = `package foo
import "fmt"
import "os"

func Generate() ([]string, error) {
	for _, arg := range os.Args {
		if arg == "{{.}}" {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("no flag {{.}} found")
}
`
)
