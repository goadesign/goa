package genapp_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/gen_app"
	"github.com/goadesign/goa/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {
	const testgenPackagePath = "github.com/goadesign/goa/goagen/gen_app/test_"

	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		gopath := filepath.SplitList(os.Getenv("GOPATH"))[0]
		outDir = filepath.Join(gopath, "src", testgenPackagePath)
		err := os.MkdirAll(outDir, 0777)
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "app", "--out=" + outDir, "--design=foo", "--version=" + version.String()}
		design.GeneratedMediaTypes = make(design.MediaTypeRoot)
	})

	JustBeforeEach(func() {
		files, genErr = genapp.Generate()
	})

	AfterEach(func() {
		os.RemoveAll(outDir)
		delete(codegen.Reserved, "app")
	})

	Context("with notest flag", func() {
		BeforeEach(func() {
			os.Args = []string{"goagen", "app", "--out=" + outDir, "--design=foo", "--notest", "--version=" + version.String()}
		})

		It("does not generate tests", func() {
			_, err := os.Stat(filepath.Join(outDir, "app", "test"))
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
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
				MediaTypes: map[string]*design.MediaTypeDefinition{
					design.ErrorMedia.Identifier: design.ErrorMedia,
				},
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
										"uuid":  &design.AttributeDefinition{Type: design.UUID},
									},
								},
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "p/:param/u/:uuid",
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
										"uuid":  &design.AttributeDefinition{Type: design.UUID},
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
										Name:      "ok",
										Type:      design.ErrorMedia,
										MediaType: "application/vnd.api.error+json",
									},
								},
							},
						},
					},
				},
			}
			fooRes := design.Design.Resources["foo"]
			for _, a := range fooRes.Actions {
				a.Parent = fooRes
				a.Routes[0].Parent = a
			}
		})

		It("generates the ActionRouteResponse test methods ", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(8))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("ShowFooOK("))
			Ω(content).Should(ContainSubstring("ShowFooOKWithContext("))
			// Multiple Routes
			Ω(content).Should(ContainSubstring("ShowFooOK1("))
			Ω(content).Should(ContainSubstring("ShowFooOK1WithContext("))
			// Get returns an error media type
			Ω(content).Should(ContainSubstring("GetFooOK(t *testing.T, ctrl app.FooController, payload app.CustomName) (http.ResponseWriter, *goa.Error)"))
		})

		It("generates the route path parameters", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring(`["param"] = []string{`))
			Ω(content).Should(ContainSubstring(`["uuid"] = []string{`))
			Ω(content).ShouldNot(ContainSubstring(`["time"] = []string{`))
		})

		It("generates calls to new Context ", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("app.NewShowFooContext("))
		})

		It("generates calls controller action method", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("ctrl.Show("))
		})

		It("generates non pointer references to primitive/array/hash payloads", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring(", payload app.CustomName)"))
		})

	})
})
