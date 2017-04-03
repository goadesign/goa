package genapp_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
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
		os.Args = []string{"goagen", "--out=" + outDir, "--design=foo", "--version=" + version.String()}
		design.GeneratedMediaTypes = make(design.MediaTypeRoot)
		design.ProjectedMediaTypes = make(design.MediaTypeRoot)
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
			os.Args = []string{"goagen", "--out=" + outDir, "--design=foo", "--notest", "--version=" + version.String()}
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

			intAttr := &design.AttributeDefinition{
				Type:       design.Object{"foo": &design.AttributeDefinition{Type: design.Integer}},
				Validation: &dslengine.ValidationDefinition{Required: []string{"foo"}},
			}

			intMedia := &design.MediaTypeDefinition{
				Identifier: "application/vnd.goa.test.int",
				UserTypeDefinition: &design.UserTypeDefinition{
					AttributeDefinition: intAttr,
					TypeName:            "IntContainer",
				},
			}

			defaultView := &design.ViewDefinition{
				AttributeDefinition: intAttr,
				Name:                "default",
				Parent:              intMedia,
			}

			intMedia.Views = map[string]*design.ViewDefinition{"default": defaultView}

			design.Design = &design.APIDefinition{
				Name:        "testapi",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
				MediaTypes: map[string]*design.MediaTypeDefinition{
					design.ErrorMedia.Identifier: design.ErrorMedia,
					intMedia.Identifier:          intMedia,
				},
				Resources: map[string]*design.ResourceDefinition{
					"foo": {
						Name: "foo",
						Headers: &design.AttributeDefinition{
							Type: design.Object{
								"optionalResourceHeader": &design.AttributeDefinition{Type: design.Integer},
								"requiredResourceHeader": &design.AttributeDefinition{Type: design.String},
							},
							Validation: &dslengine.ValidationDefinition{Required: []string{"requiredResourceHeader"}},
						},
						Actions: map[string]*design.ActionDefinition{
							"show": {
								Name: "show",
								Params: &design.AttributeDefinition{
									Type: design.Object{
										"param":    &design.AttributeDefinition{Type: design.Integer},
										"time":     &design.AttributeDefinition{Type: design.DateTime},
										"uuid":     &design.AttributeDefinition{Type: design.UUID},
										"optional": &design.AttributeDefinition{Type: design.Integer},
										"required": &design.AttributeDefinition{Type: design.DateTime},
										"query":    &design.AttributeDefinition{Type: design.String},
									},
									Validation: &dslengine.ValidationDefinition{Required: []string{"required"}},
								},
								Headers: &design.AttributeDefinition{
									Type: design.Object{
										"optionalHeader": &design.AttributeDefinition{Type: design.Integer},
										"requiredHeader": &design.AttributeDefinition{Type: design.String},
									},
									Validation: &dslengine.ValidationDefinition{Required: []string{"requiredHeader", "requiredResourceHeader"}},
								},
								QueryParams: &design.AttributeDefinition{
									Type: design.Object{
										"optional": &design.AttributeDefinition{Type: design.Integer},
										"required": &design.AttributeDefinition{Type: design.DateTime},
										"query":    &design.AttributeDefinition{Type: design.String},
									},
									Validation: &dslengine.ValidationDefinition{Required: []string{"required"}},
								},
								Routes: []*design.RouteDefinition{
									{
										Verb: "GET",
										Path: "p/:param/u/:uuid/:required",
									},
									{
										Verb: "POST",
										Path: "",
									},
								},
								Responses: map[string]*design.ResponseDefinition{
									"ok": {
										Name:      "ok",
										Status:    200,
										MediaType: intMedia.Identifier,
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
										MediaType: "application/vnd.goa.error",
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

		It("does not call Validate on the resulting media type when it does not exist", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(8))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).ShouldNot(ContainSubstring("err = mt.Validate()"))
		})

		It("generates the ActionRouteResponse test methods ", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(8))
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("ShowFooOK("))
			// Multiple Routes
			Ω(content).Should(ContainSubstring("ShowFooOK1("))
			// Get returns an error media type
			Ω(content).Should(ContainSubstring("GetFooOK(t goatest.TInterface, ctx context.Context, service *goa.Service, ctrl app.FooController, optionalResourceHeader *int, requiredResourceHeader string, payload app.CustomName) (http.ResponseWriter, error)"))
		})

		It("generates the route path parameters", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring(`["param"] = []string{`))
			Ω(content).Should(ContainSubstring(`["uuid"] = []string{`))
			Ω(content).Should(ContainSubstring(`["required"] = []string{`))
			Ω(content).ShouldNot(ContainSubstring(`["time"] = []string{`))
		})

		It("properly handles query parameters", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring(`if optional != nil`))
			Ω(content).ShouldNot(ContainSubstring(`if required != nil`))
			Ω(content).Should(ContainSubstring(`if query != nil`))
		})

		It("properly handles headers", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring(`if optionalHeader != nil`))
			Ω(content).ShouldNot(ContainSubstring(`if requiredHeader != nil`))
			Ω(content).Should(ContainSubstring(`req.Header["requiredHeader"] = sliceVal`))
			Ω(content).Should(ContainSubstring(`if optionalResourceHeader != nil`))
			Ω(content).ShouldNot(ContainSubstring(`if requiredResourceHeader != nil`))
			Ω(content).Should(ContainSubstring(`req.Header["requiredResourceHeader"] = sliceVal`))
		})

		It("generates calls to new Context ", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("app.NewShowFooContext("))
		})

		It("generates calls controller action method", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring("ctrl.Show("))
		})

		It("generates non pointer references to primitive/array/hash payloads", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())

			Ω(content).Should(ContainSubstring(", payload app.CustomName)"))
		})

		It("generates header compliant with https://github.com/golang/go/issues/13560", func() {
			content, err := ioutil.ReadFile(filepath.Join(outDir, "app", "test", "foo_testing.go"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(strings.Split(string(content), "\n")).Should(ContainElement(MatchRegexp(`^// Code generated .* DO NOT EDIT\.$`)))
		})
	})
})
