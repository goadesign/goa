package genmain_test

import (
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
			_, err = gexec.Build(testgenPackagePath)
			Ω(err).ShouldNot(HaveOccurred())
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
		noExample bool
	}{
		api: &design.APIDefinition{
			Name: "test api",
		},
		outDir:    "out_dir",
		designPkg: "design",
		target:    "app",
		force:     false,
	}

	Context("with options all options set", func() {
		BeforeEach(func() {

			generator = genmain.NewGenerator(
				genmain.API(args.api),
				genmain.OutDir(args.outDir),
				genmain.DesignPkg(args.designPkg),
				genmain.Target(args.target),
				genmain.Force(args.force),
			)
		})

		It("has all public properties set with expected value", func() {
			Ω(generator).ShouldNot(BeNil())
			Ω(generator.API.Name).Should(Equal(args.api.Name))
			Ω(generator.OutDir).Should(Equal(args.outDir))
			Ω(generator.DesignPkg).Should(Equal(args.designPkg))
			Ω(generator.Target).Should(Equal(args.target))
			Ω(generator.Force).Should(Equal(args.force))
		})

	})
})
