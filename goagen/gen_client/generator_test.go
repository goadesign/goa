package genclient_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_client/test_"

	var outDir string
	var files []string
	var genErr error

	var oldCommand string

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"codegen", "--out=" + outDir, "--design=foo"}
		oldCommand = codegen.CommandName
		codegen.CommandName = "app"
	})

	JustBeforeEach(func() {
		files, genErr = genclient.Generate([]interface{}{design.Design})
	})

	AfterEach(func() {
		codegen.CommandName = oldCommand
		os.RemoveAll(outDir)
	})

	Context("with a dummy API", func() {
		BeforeEach(func() {
			design.Design = &design.APIDefinition{
				APIVersionDefinition: &design.APIVersionDefinition{
					Name:        "testapi",
					Title:       "dummy API with no resource",
					Description: "I told you it's dummy",
				},
			}
		})

		It("generates a dummy app", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(5))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "testapi-cli", "main.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 16))
			_, err = gexec.Build(filepath.Join(testgenPackagePath, "client", "testapi-cli"))
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("with an action with an integer parameter with no default value", func() {
		BeforeEach(func() {
			codegen.TempCount = 0
			design.Design = &design.APIDefinition{
				APIVersionDefinition: &design.APIVersionDefinition{
					Name:        "testapi",
					Title:       "dummy API with no resource",
					Description: "I told you it's dummy",
				},
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								QueryParams: &design.AttributeDefinition{
									Type: design.Object{
										"param": &design.AttributeDefinition{Type: design.Integer},
									},
								},
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
								},
							},
						},
					},
				},
			}
			fooRes := design.Design.Resources["foo"]
			showAct := fooRes.Actions["show"]
			showAct.Parent = fooRes
			showAct.Routes[0].Parent = showAct
		})

		It("generates the correct command flag initialization code", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(6))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "client", "testapi-cli", "commands.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(strings.Split(string(content), "\n"))).Should(BeNumerically(">=", 3))
			Ω(content).Should(ContainSubstring("var tmp2 int"))
			Ω(content).Should(ContainSubstring(".Flags()"))
			_, err = gexec.Build(filepath.Join(testgenPackagePath, "client", "testapi-cli"))
			Ω(err).ShouldNot(HaveOccurred())

		})
	})
})
