package genschema_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/design/apidsl"
	"goa.design/goa.v2/dslengine"
	"goa.design/goa.v2/pkg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	var files []string
	var genErr error
	var workspace *codegen.Workspace
	var testPkg *codegen.Package

	BeforeEach(func() {
		var err error
		workspace, err = codegen.NewWorkspace("test")
		Ω(err).ShouldNot(HaveOccurred())
		testPkg, err = workspace.NewPackage("schematest")
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + testPkg.Abs(), "--design=foo", "--version=" + pkg.Version()}
	})

	JustBeforeEach(func() {
		files, genErr = genschema.Generate()
	})

	AfterEach(func() {
		workspace.Delete()
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			dslengine.Reset()
			apidsl.API("test api", func() {
				apidsl.Title("dummy API with no resource")
				apidsl.Description("I told you it's dummy")
			})
			dslengine.Run()
		})

		It("generates a dummy schema", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(2))
			content, err := ioutil.ReadFile(filepath.Join(testPkg.Abs(), "schema", "schema.json"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically("==", 1))
			var s genschema.JSONSchema
			err = json.Unmarshal(content, &s)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})
