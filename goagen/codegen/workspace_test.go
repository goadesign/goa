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

var _ = Describe("Workspace", func() {
	Describe("WorkspaceFor", func() {
		oldGOPATH := build.Default.GOPATH
		BeforeEach(func() {
			os.Setenv("GOPATH", "/xx")
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
						workspace, err := codegen.WorkspaceFor("/xx/bar/xx/42")
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal("/xx"))
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
						workspace, err := codegen.WorkspaceFor("/xx/bar/xx/42")
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal("/xx"))
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
						_, err := codegen.WorkspaceFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
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
						workspace, err := codegen.WorkspaceFor("/xx/bar/xx/42")
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal("/xx"))
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
						_, err := codegen.WorkspaceFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
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
						_, err := codegen.WorkspaceFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH /xx or use modules`)))
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
						_, err := codegen.WorkspaceFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
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
						workspace, err := codegen.WorkspaceFor("/xx/bar/xx/42")
						Ω(err).ShouldNot(HaveOccurred())
						Expect(workspace.Path).To(Equal("/xx"))
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
						_, err := codegen.WorkspaceFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
					})
				})
			})
		})
	})

	Describe("PackageFor", func() {
		oldGOPATH := build.Default.GOPATH
		BeforeEach(func() {
			os.Setenv("GOPATH", "/xx")
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
						pkg, err := codegen.PackageFor("/xx/src/bar/xx/42")
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
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						pkg, err := codegen.PackageFor(filepath.Join(abs, "bar/xx/42"))
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
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						pkg, err := codegen.PackageFor(filepath.Join(abs, "bar/xx/42"))
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
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						pkg, err := codegen.PackageFor(filepath.Join(abs, "bar/xx/42"))
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
						pkg, err := codegen.PackageFor("/xx/src/bar/xx/42")
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
						_, err := codegen.PackageFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
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
						pkg, err := codegen.PackageFor("/xx/src/bar/xx/42")
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
						_, err := codegen.PackageFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
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
						_, err := codegen.PackageFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH /xx or use modules`)))
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
						_, err := codegen.PackageFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
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
						pkg, err := codegen.PackageFor("/xx/src/bar/xx/42")
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
						_, err := codegen.PackageFor("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf(`Go source file "/bar/xx/42" not in Go workspace, adjust GOPATH %s or use modules`, gopath)))
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
				pkg, err := codegen.PackageFor("/xx/src/bar/xx/42")
				Ω(err).ShouldNot(HaveOccurred())
				Expect(pkg.Abs()).To(Equal("/xx/src/bar/xx"))
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
				abs, err := filepath.Abs(".")
				Ω(err).ShouldNot(HaveOccurred())
				pkg, err := codegen.PackageFor(filepath.Join(abs, "bar/xx/42"))
				Ω(err).ShouldNot(HaveOccurred())
				Expect(pkg.Abs()).To(Equal(filepath.Join(abs, "bar/xx")))
			})
		})
	})

	Describe("PackagePath", func() {
		oldGOPATH := build.Default.GOPATH
		BeforeEach(func() {
			os.Setenv("GOPATH", "/xx")
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
						p, err := codegen.PackagePath("/xx/src/bar/xx/42")
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
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						p, err := codegen.PackagePath(filepath.Join(abs, "bar/xx/42"))
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
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						p, err := codegen.PackagePath(filepath.Join(abs, "bar/xx/42"))
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
						abs, err := filepath.Abs(".")
						Ω(err).ShouldNot(HaveOccurred())
						p, err := codegen.PackagePath(filepath.Join(abs, "bar/xx/42"))
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
						p, err := codegen.PackagePath("/xx/src/bar/xx/42")
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
						_, err := codegen.PackagePath("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf("/bar/xx/42 does not contain a Go package")))
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
						p, err := codegen.PackagePath("/xx/src/bar/xx/42")
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
						_, err := codegen.PackagePath("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf("/bar/xx/42 does not contain a Go package")))
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
						_, err := codegen.PackagePath("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf("/bar/xx/42 does not contain a Go package")))
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
						_, err := codegen.PackagePath("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf("/bar/xx/42 does not contain a Go package")))
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
						p, err := codegen.PackagePath("/xx/src/bar/xx/42")
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
						_, err := codegen.PackagePath("/bar/xx/42")
						Ω(err).Should(Equal(fmt.Errorf("/bar/xx/42 does not contain a Go package")))
					})
				})
			})
		})
	})

})
