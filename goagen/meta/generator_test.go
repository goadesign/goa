package meta_test

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/meta"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run", func() {

	const invalidPkgPath = "foobar"

	var (
		compiledFiles   []string
		compileError    error
		outputWorkspace *codegen.Workspace
		designWorkspace *codegen.Workspace
		genWorkspace    *codegen.Workspace

		outputDir     string
		designPkgPath string
		genPkgSource  string
		customFlags   []string

		m *meta.Generator
	)

	BeforeEach(func() {
		designPkgPath = "design"
		genPkgSource = "package gen\nfunc Generate() ([]string, error) { return nil, nil }"

		var err error

		outputWorkspace, err = codegen.NewWorkspace("output")
		p, err := outputWorkspace.NewPackage("testOutput")
		Ω(err).ShouldNot(HaveOccurred())
		outputDir = p.Abs()

		designWorkspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())

		genWorkspace, err = codegen.NewWorkspace("gen")
		Ω(err).ShouldNot(HaveOccurred())

		compiledFiles = nil
		compileError = nil
		customFlags = []string{"--custom=arg"}
	})

	JustBeforeEach(func() {
		if designPkgPath != "" && designPkgPath != invalidPkgPath {
			designPackage, err := designWorkspace.NewPackage(designPkgPath)
			Ω(err).ShouldNot(HaveOccurred())
			file, err := designPackage.CreateSourceFile("design.go")
			Ω(err).ShouldNot(HaveOccurred())
			_, err = file.Write([]byte("package design"))
			Ω(err).ShouldNot(HaveOccurred())
			file.Close()
		}

		genPackage, err := genWorkspace.NewPackage("gen")
		Ω(err).ShouldNot(HaveOccurred())
		file, err := genPackage.CreateSourceFile("gen.go")
		Ω(err).ShouldNot(HaveOccurred())
		_, err = file.Write([]byte(genPkgSource))
		Ω(err).ShouldNot(HaveOccurred())
		file.Close()

		m = &meta.Generator{
			Genfunc:       "gen.Generate",
			Imports:       []*codegen.ImportSpec{codegen.SimpleImport("gen")},
			OutDir:        outputDir,
			CustomFlags:   customFlags,
			DesignPkgPath: designPkgPath,
		}
		compiledFiles, compileError = m.Generate()
	})

	AfterEach(func() {
		designWorkspace.Delete()
		outputWorkspace.Delete()
		genWorkspace.Delete()
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
		BeforeEach(func() {
			designPkgPath = invalidPkgPath
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError(HavePrefix("invalid design package import path: cannot find package")))
			Ω(compileError).Should(MatchError(ContainSubstring(invalidPkgPath)))
		})
	})

	Context("with no go compiler in PATH", func() {
		var pathEnv string
		const invalidPath = "/foobar"

		BeforeEach(func() {
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
			outputDir = ""
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("missing output directory flag"))
		})
	})

	Context("with no design package path specified", func() {
		BeforeEach(func() {
			designPkgPath = ""
		})

		It("fails with a useful error message", func() {
			Ω(compileError).Should(MatchError("missing design package flag"))
		})
	})

	Context("with gen package content", func() {

		BeforeEach(func() {
			outputDir = os.TempDir()
		})

		Context("that is not valid Go code", func() {
			BeforeEach(func() {
				genPkgSource = invalidSource
			})

			It("fails with a useful error message", func() {
				Ω(compileError.Error()).Should(ContainSubstring("syntax error"))
			})
		})

		Context("whose code blows up", func() {
			BeforeEach(func() {
				genPkgSource = panickySource
			})

			It("fails with a useful error message", func() {
				Ω(compileError.Error()).Should(ContainSubstring("panic: kaboom"))
			})
		})

		Context("with valid code", func() {
			BeforeEach(func() {
				genPkgSource = validSource
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
				genPkgSource = b.String()
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
				genPkgSource = b.String()

			})

			It("returns no error", func() {
				Ω(compileError).ShouldNot(HaveOccurred())
			})
		})
	})
})

const (
	invalidSource = `package gen
invalid go code
`

	panickySource = `package gen
func Generate() ([]string, error) {
	return nil, nil
}

func init() { panic("kaboom") }
`

	validSource = `package gen
func Generate() ([]string, error) {
	return nil, nil
}
`

	validSourceTmpl = `package gen
import "fmt"
func Generate() ([]string, error) {
	{{range .}}fmt.Println("{{.}}")
	{{end}}
	return nil, nil
}
`

	validSourceTmplWithCustomFlags = `package gen
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
