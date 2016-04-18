package gentest_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_test/test_"

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
		files, genErr = gentest.Generate()
	})

	AfterEach(func() {
		codegen.CommandName = oldCommand
		os.RemoveAll(outDir)
	})

	Context("with an basic action", func() {
		BeforeEach(func() {
			codegen.TempCount = 0

			userType := &design.UserTypeDefinition{
				AttributeDefinition: &design.AttributeDefinition{Type: &design.Array{ElemType: &design.AttributeDefinition{Type: design.String}}},
				TypeName:            "CustomName",
			}

			design.Design = &design.APIDefinition{
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								Params: &design.AttributeDefinition{
									Type: design.Object{
										"param": &design.AttributeDefinition{Type: design.Integer},
										"time":  &design.AttributeDefinition{Type: design.DateTime},
									},
								},
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
									{
										Verb: "POST",
										Path: "",
									},
								},
								Responses: map[string]*design.ResponseDefinition{
									"ok": {
										Name: "ok",
									},
								},
							},
							"get": {
								Name: "get",
								Params: &design.AttributeDefinition{
									Type: design.Object{
										"param": &design.AttributeDefinition{Type: design.Integer},
										"time":  &design.AttributeDefinition{Type: design.DateTime},
									},
								},
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "",
									},
								},
								Payload: userType,
								Responses: map[string]*design.ResponseDefinition{
									"ok": {
										Name: "ok",
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

		It("generates the ActionRouteResponse test methods ", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(2))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("ShowFooOK("))
			Ω(content).Should(ContainSubstring("ShowFooOKCtx("))
			// Multiple Routes
			Ω(content).Should(ContainSubstring("ShowFooOK1("))
			Ω(content).Should(ContainSubstring("ShowFooOK1Ctx("))
		})

		It("generates calls to new Context ", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("app.NewShowFooContext("))
		})

		It("generates calls controller action method", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("ctrl.Show("))
		})

		It("generates non pointer references to primitive/array/hash payloads", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("payload app.CustomName) {"))
		})

	})
})
