package genmain_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/gen_main"
	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_main/goatest"

	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + outDir, "--design=foo", "--version=" + version.String()}
	})

	JustBeforeEach(func() {
		files, genErr = genmain.Generate()
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

		It("generates a dummy app", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(1))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "main.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 16))
			Ω(string(content)).Should(ContainSubstring(listenAndServeCode))
			_, err = gexec.Build(testgenPackagePath)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("via HTTPS", func() {
			BeforeEach(func() {
				design.Design.Schemes = []string{"https"}
			})

			It("generates a dummy app", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(1))
				content, err := ioutil.ReadFile(filepath.Join(outDir, "main.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 16))
				Ω(string(content)).Should(ContainSubstring(listenAndServeTLSCode))
				_, err = gexec.Build(testgenPackagePath)
				Ω(err).ShouldNot(HaveOccurred())
			})

		})
	})

	Context("with resources", func() {
		var resource *design.ResourceDefinition

		BeforeEach(func() {
			resource = &design.ResourceDefinition{
				Name:        "first",
				Description: "first stuff",
				Actions:     map[string]*design.ActionDefinition{},
			}
			alpha := &design.ActionDefinition{
				Parent:      resource,
				Name:        "alpha",
				Schemes:     []string{"http"},
				Description: "Alpha-like things",
			}
			resource.Actions[alpha.Name] = alpha
			design.Design = &design.APIDefinition{
				Name:        "whatever",
				Title:       "test API",
				Description: "Ain't matter none",
				Resources: map[string]*design.ResourceDefinition{
					"first": resource,
				},
			}
		})

		It("generates controllers ready for regeneration", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(2))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "first.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(content).Should(MatchRegexp("FirstController_Alpha: start_implement"))
			Ω(content).Should(MatchRegexp(`// FirstController_Alpha: start_implement\s*// Put your logic here\s*return nil\s*// FirstController_Alpha: end_implement`))
		})

		Context("regenerated with a new resource", func() {
			BeforeEach(func() {
				// Perform a first generation
				files, genErr = genmain.Generate()

				// Put some impl in the existing controller
				existing, err := ioutil.ReadFile(filepath.Join(outDir, "first.go"))
				Ω(err).ShouldNot(HaveOccurred())

				// First add an import for fmt, to make sure it remains
				existing = bytes.Replace(existing, []byte("import ("), []byte("import (\n\t\"fmt\")"), 1)

				// Next add some body that uses fmt
				existing = bytes.Replace(existing, []byte("// Put your logic here"), []byte("fmt.Println(\"I did it first\")"), 1)

				err = ioutil.WriteFile(filepath.Join(outDir, "first.go"), existing, os.ModePerm)
				Ω(err).ShouldNot(HaveOccurred())

				// Add an action to the existing resource
				beta := &design.ActionDefinition{
					Parent:      resource,
					Name:        "beta",
					Schemes:     []string{"http"},
					Description: "Beta-like things",
				}
				resource.Actions[beta.Name] = beta

				// Add a new resource
				resource2 := &design.ResourceDefinition{
					Name:        "second",
					Description: "second stuff",
					Actions:     map[string]*design.ActionDefinition{},
				}
				gamma := &design.ActionDefinition{
					Parent:      resource2,
					Name:        "gamma",
					Schemes:     []string{"http"},
					Description: "Gamma-like things",
				}
				resource2.Actions[gamma.Name] = gamma

				design.Design.Resources[resource2.Name] = resource2

				// Set up the regeneration for the JustBeforeEach
				os.Args = append(os.Args, "--regen")
			})

			It("generates scaffolding for new and existing resources", func() {
				Ω(genErr).Should(BeNil())
				Ω(files).Should(HaveLen(2))
				Ω(files).Should(ConsistOf(filepath.Join(outDir, "first.go"), filepath.Join(outDir, "second.go")))

				content, err := ioutil.ReadFile(filepath.Join(outDir, "second.go"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(content).Should(ContainSubstring("SecondController_Gamma: start_implement"))

			})

			It("regenerates controllers without modifying existing impls", func() {
				content, err := ioutil.ReadFile(filepath.Join(outDir, "first.go"))
				Ω(err).ShouldNot(HaveOccurred())

				// First make sure the new controller is in place
				Ω(content).Should(ContainSubstring("FirstController_Beta: start_implement"))

				// Check the fmt import
				Ω(string(content)).Should(MatchRegexp(`import \(\s*[^)]*\"fmt\"`))

				// Check the body is in place
				Ω(content).Should(MatchRegexp(`// FirstController_Alpha: start_implement\s*fmt.Println\("I did it first"\)\s*return nil\s*// FirstController_Alpha: end_implement`))
			})
		})

	})
})

var _ = Describe("NewGenerator", func() {
	var generator *genmain.Generator

	var args = struct {
		api       *design.APIDefinition
		outDir    string
		designPkg string
		target    string
		force     bool
		regen     bool
		noExample bool
	}{
		api: &design.APIDefinition{
			Name: "test api",
		},
		outDir:    "out_dir",
		designPkg: "design",
		target:    "app",
		force:     false,
		regen:     false,
	}

	Context("with options all options set", func() {
		BeforeEach(func() {

			generator = genmain.NewGenerator(
				genmain.API(args.api),
				genmain.OutDir(args.outDir),
				genmain.DesignPkg(args.designPkg),
				genmain.Target(args.target),
				genmain.Force(args.force),
				genmain.Regen(args.regen),
			)
		})

		It("has all public properties set with expected value", func() {
			Ω(generator).ShouldNot(BeNil())
			Ω(generator.API.Name).Should(Equal(args.api.Name))
			Ω(generator.OutDir).Should(Equal(args.outDir))
			Ω(generator.DesignPkg).Should(Equal(args.designPkg))
			Ω(generator.Target).Should(Equal(args.target))
			Ω(generator.Force).Should(Equal(args.force))
			Ω(generator.Regen).Should(Equal(args.regen))
		})

	})
})

const listenAndServeCode = `
	if err := service.ListenAndServe(":8080"); err != nil {
		service.LogError("startup", "err", err)
	}
`

const listenAndServeTLSCode = `
	if err := service.ListenAndServeTLS(":8080", "cert.pem", "key.pem"); err != nil {
		service.LogError("startup", "err", err)
	}
`
