package codegen_test

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/goagen/codegen"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func abs(elems ...string) string {
	r, err := filepath.Abs(filepath.Join(append([]string{""}, elems...)...))
	if err != nil {
		panic("abs: " + err.Error())
	}
	return r
}

var _ = Describe("Workspace", func() {
	Describe("WorkspaceFor", func() {
		oldGOPATH := build.Default.GOPATH
		xx := abs("xx")
		BeforeEach(func() {
			os.Setenv("GOPATH", xx)
		})
		AfterEach(func() {
			os.Setenv("GOPATH", oldGOPATH)
		})

		var (
			err    error
			gopath string
		)
		Context("with GOMOD", func() {
			var (
				f *os.File
			)
			BeforeEach(func() {
				f, err = os.OpenFile("go.mod", os.O_CREATE, 0755)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f.Close()).ShouldNot(HaveOccurred())
			})
			AfterEach(func() {
				Ω(os.RemoveAll("go.mod")).ShouldNot(HaveOccurred())
			})

			Context("with GO111MODULE=auto", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Unsetenv("GO111MODULE")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode workspace", func() {
						workspace, err := codegen.WorkspaceFor(abs("", "xx", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal(xx))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return a Module mode workspace", func() {
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						workspace, err := codegen.WorkspaceFor(abs)
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal(abs))
					})
				})
			})

			Context("with GO111MODULE=on", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "on")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a Module mode workspace", func() {
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						workspace, err := codegen.WorkspaceFor(abs)
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal(abs))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return a Module mode workspace", func() {
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						workspace, err := codegen.WorkspaceFor(abs)
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal(abs))
					})
				})
			})

			Context("with GO111MODULE=off", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "off")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode workspace", func() {
						workspace, err := codegen.WorkspaceFor(abs("", "xx", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal(xx))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.WorkspaceFor(abs("", "bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("", "bar", "xx", "42"), gopath)))
					})
				})
			})
		})

		Context("with no GOMOD", func() {
			Context("with GO111MODULE=auto", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Unsetenv("GO111MODULE")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode workspace", func() {
						workspace, err := codegen.WorkspaceFor(abs("xx", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal(abs("xx")))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.WorkspaceFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), gopath)))
					})
				})
			})

			Context("with GO111MODULE=on", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "on")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return an error", func() {
						_, err := codegen.WorkspaceFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), abs("xx"))))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.WorkspaceFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), gopath)))
					})
				})
			})

			Context("with GO111MODULE=off", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "off")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode workspace", func() {
						workspace, err := codegen.WorkspaceFor(abs("xx", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal(abs("xx")))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.WorkspaceFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), gopath)))
					})
				})
			})
		})
	})

	Describe("PackageFor", func() {
		oldGOPATH := build.Default.GOPATH
		BeforeEach(func() {
			os.Setenv("GOPATH", abs("xx"))
		})
		AfterEach(func() {
			os.Setenv("GOPATH", oldGOPATH)
		})

		var (
			err    error
			gopath string
		)
		Context("with GOMOD", func() {
			var (
				f *os.File
			)
			BeforeEach(func() {
				f, err = os.OpenFile("go.mod", os.O_CREATE, 0755)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f.Close()).ShouldNot(HaveOccurred())
			})
			AfterEach(func() {
				Ω(os.RemoveAll("go.mod")).ShouldNot(HaveOccurred())
			})

			Context("with GO111MODULE=auto", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Unsetenv("GO111MODULE")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package", func() {
						pkg, err := codegen.PackageFor(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(pkg.Path).To(Equal("bar/xx"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return a Module mode package", func() {
						ab, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						pkg, err := codegen.PackageFor(abs(ab, "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(pkg.Path).To(Equal("bar/xx"))
					})
				})
			})

			Context("with GO111MODULE=on", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "on")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a Module mode package", func() {
						ab, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						pkg, err := codegen.PackageFor(abs(ab, "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(pkg.Path).To(Equal("bar/xx"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return a Module mode package", func() {
						ab, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						pkg, err := codegen.PackageFor(abs(ab, "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(pkg.Path).To(Equal("bar/xx"))
					})
				})
			})

			Context("with GO111MODULE=off", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "off")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package", func() {
						pkg, err := codegen.PackageFor(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(pkg.Path).To(Equal("bar/xx"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackageFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), gopath)))
					})
				})
			})
		})

		Context("with no GOMOD", func() {
			Context("with GO111MODULE=auto", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Unsetenv("GO111MODULE")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package", func() {
						pkg, err := codegen.PackageFor(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(pkg.Path).To(Equal("bar/xx"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackageFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), gopath)))
					})
				})
			})

			Context("with GO111MODULE=on", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "on")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return an error", func() {
						_, err := codegen.PackageFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), abs("xx"))))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackageFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), gopath)))
					})
				})
			})

			Context("with GO111MODULE=off", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "off")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package", func() {
						pkg, err := codegen.PackageFor(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(pkg.Path).To(Equal("bar/xx"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackageFor(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "%s" not in Go workspace, adjust GOPATH %s or use modules`, abs("bar", "xx", "42"), gopath)))
					})
				})
			})
		})
	})

	Describe("Package.Abs", func() {
		var (
			err            error
			gopath         string
			f              *os.File
			oldGOPATH      = build.Default.GOPATH
			oldGO111MODULE = os.Getenv("GO111MODULE")
		)
		BeforeEach(func() {
			os.Setenv("GOPATH", "/xx")
			f, err = os.OpenFile("go.mod", os.O_CREATE, 0755)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(f.Close()).ShouldNot(HaveOccurred())
			os.Unsetenv("GO111MODULE")
		})
		AfterEach(func() {
			os.Setenv("GOPATH", oldGOPATH)
			Ω(os.RemoveAll("go.mod")).ShouldNot(HaveOccurred())
			os.Setenv("GO111MODULE", oldGO111MODULE)
		})

		Context("inside GOPATH", func() {
			It("should return the absolute path to the GOPATH directory", func() {
				pkg, err := codegen.PackageFor(abs("xx", "src", "bar", "xx", "42"))
				Ω(err).ShouldNot(HaveOccurred())
				Expect(pkg.Abs()).To(Equal(abs("xx", "src", "bar", "xx")))
			})
		})

		Context("outside GOPATH", func() {
			BeforeEach(func() {
				gopath, err = ioutil.TempDir(".", "go")
				Ω(err).ShouldNot(HaveOccurred())
				os.Setenv("GOPATH", gopath)
			})
			AfterEach(func() {
				Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
			})

			It("should return the absolute path to the Module directory", func() {
				ab, err := filepath.Abs(".")
				Ω(err).ShouldNot(HaveOccurred())
				pkg, err := codegen.PackageFor(abs(ab, "bar", "xx", "42"))
				Ω(err).ShouldNot(HaveOccurred())
				Expect(pkg.Abs()).To(Equal(abs(ab, "bar", "xx")))
			})
		})
	})

	Describe("PackagePath", func() {
		oldGOPATH := build.Default.GOPATH
		BeforeEach(func() {
			os.Setenv("GOPATH", abs("xx"))
		})
		AfterEach(func() {
			os.Setenv("GOPATH", oldGOPATH)
		})

		var (
			err    error
			gopath string
		)
		Context("with GOMOD", func() {
			var (
				f *os.File
			)
			BeforeEach(func() {
				f, err = os.OpenFile("go.mod", os.O_CREATE, 0755)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(f.Close()).ShouldNot(HaveOccurred())
			})
			AfterEach(func() {
				Ω(os.RemoveAll("go.mod")).ShouldNot(HaveOccurred())
			})

			Context("with GO111MODULE=auto", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Unsetenv("GO111MODULE")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package path", func() {
						p, err := codegen.PackagePath(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(p).To(Equal("bar/xx/42"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return a Module mode package path", func() {
						ab, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						p, err := codegen.PackagePath(abs(ab, "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(p).To(Equal("bar/xx/42"))
					})
				})
			})

			Context("with GO111MODULE=on", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "on")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a Module mode package path", func() {
						ab, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						p, err := codegen.PackagePath(abs(ab, "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(p).To(Equal("bar/xx/42"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return a Module mode package path", func() {
						ab, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						p, err := codegen.PackagePath(abs(ab, "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(p).To(Equal("bar/xx/42"))
					})
				})
			})

			Context("with GO111MODULE=off", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "off")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package path", func() {
						p, err := codegen.PackagePath(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(p).To(Equal("bar/xx/42"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackagePath(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf("%s does not contain a Go package", abs("bar", "xx", "42"))))
					})
				})
			})
		})

		Context("with no GOMOD", func() {
			Context("with GO111MODULE=auto", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Unsetenv("GO111MODULE")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package path", func() {
						p, err := codegen.PackagePath(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(p).To(Equal("bar/xx/42"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackagePath(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf("%s does not contain a Go package", abs("bar", "xx", "42"))))
					})
				})
			})

			Context("with GO111MODULE=on", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "on")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return an error", func() {
						_, err := codegen.PackagePath(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf("%s does not contain a Go package", abs("bar", "xx", "42"))))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackagePath(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf("%s does not contain a Go package", abs("bar", "xx", "42"))))
					})
				})
			})

			Context("with GO111MODULE=off", func() {
				oldGO111MODULE := os.Getenv("GO111MODULE")
				BeforeEach(func() {
					os.Setenv("GO111MODULE", "off")
				})
				AfterEach(func() {
					os.Setenv("GO111MODULE", oldGO111MODULE)
				})

				Context("inside GOPATH", func() {
					It("should return a GOPATH mode package path", func() {
						p, err := codegen.PackagePath(abs("xx", "src", "bar", "xx", "42"))
						Ω(err).ShouldNot(HaveOccurred())
						Expect(p).To(Equal("bar/xx/42"))
					})
				})

				Context("outside GOPATH", func() {
					BeforeEach(func() {
						gopath, err = ioutil.TempDir(".", "go")
						Ω(err).ShouldNot(HaveOccurred())
						os.Setenv("GOPATH", gopath)
					})
					AfterEach(func() {
						Ω(os.RemoveAll(gopath)).ShouldNot(HaveOccurred())
					})

					It("should return an error", func() {
						_, err := codegen.PackagePath(abs("bar", "xx", "42"))
						Ω(err).Should(Equal(fmt.Errorf("%s does not contain a Go package", abs("bar", "xx", "42"))))
					})
				})
			})
		})
	})

})
