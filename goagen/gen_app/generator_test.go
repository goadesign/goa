package genapp_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/gen_app"
)

var _ = Describe("NewGenerator", func() {
	var gen *genapp.Generator

	Context("with dummy command line flags", func() {
		BeforeEach(func() {
			os.Args = []string{"goagen", "--out=foo", "--design=bar", "--force"}
		})

		It("instantiates a generator with initialized writers", func() {
			design.Design = &design.APIDefinition{Name: "foo"}
			var err error
			gen, err = genapp.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
			Ω(gen.ContextsWriter).ShouldNot(BeNil())
			Ω(gen.HandlersWriter).ShouldNot(BeNil())
			Ω(gen.ResourcesWriter).ShouldNot(BeNil())
		})

		It("instantiates a generator with initialized writers even if Design is not initialized", func() {
			design.Design = nil
			var err error
			gen, err = genapp.NewGenerator()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(gen).ShouldNot(BeNil())
			Ω(gen.ContextsWriter).ShouldNot(BeNil())
			Ω(gen.HandlersWriter).ShouldNot(BeNil())
			Ω(gen.ResourcesWriter).ShouldNot(BeNil())
		})
	})
})

var _ = Describe("Generate", func() {
	var gen *genapp.Generator
	var outDir string
	var files []string
	var genErr error

	BeforeEach(func() {
		var err error
		outDir, err = ioutil.TempDir("", "")
		Ω(err).ShouldNot(HaveOccurred())
		os.Args = []string{"goagen", "--out=" + outDir, "--design=foo"}
	})

	JustBeforeEach(func() {
		var err error
		gen, err = genapp.NewGenerator()
		Ω(err).ShouldNot(HaveOccurred())
		files, genErr = gen.Generate(design.Design)
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

		It("generates correct empty files", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(3))
			isEmptySource := func(filename string) {
				contextsContent, err := ioutil.ReadFile(filepath.Join(outDir, "app", filename))
				Ω(err).ShouldNot(HaveOccurred())
				lines := strings.Split(string(contextsContent), "\n")
				Ω(lines).ShouldNot(BeEmpty())
				Ω(len(lines)).Should(BeNumerically(">", 1))
				last2Lines := lines[len(lines)-2] + "\n" + lines[len(lines)-1]
				Ω(last2Lines).Should(Equal("package app\n"))
			}
			isEmptySource("contexts.go")
			isEmptySource("handlers.go")
			isEmptySource("resources.go")
		})
	})

	Context("with a simple API", func() {
		BeforeEach(func() {
			required := design.ValidationDefinition(&design.RequiredValidationDefinition{
				Names: []string{"id"},
			})
			idAt := design.AttributeDefinition{
				Type:        design.String,
				Description: "widget id",
			}
			params := design.AttributeDefinition{
				Type: design.Object{
					"id": &idAt,
				},
				Validations: []design.ValidationDefinition{required},
			}
			resp := design.ResponseDefinition{
				Name:        "ok",
				Status:      200,
				Description: "get of widgets",
				MediaType:   "vnd.rightscale.goagen.test.widgets",
			}
			route := design.RouteDefinition{
				Verb: "GET",
				Path: "/:id",
			}
			at := design.AttributeDefinition{
				Type: design.String,
			}
			ut := design.UserTypeDefinition{
				AttributeDefinition: &at,
				TypeName:            "id",
			}
			res := design.ResourceDefinition{
				Name:            "Widget",
				BasePath:        "/widgets",
				Description:     "Widgetty",
				MediaType:       "vnd.rightscale.goagen.test.widgets",
				CanonicalAction: "get",
			}
			get := design.ActionDefinition{
				Name:        "get",
				Description: "get widgets",
				Parent:      &res,
				Routes:      []*design.RouteDefinition{&route},
				Responses:   map[string]*design.ResponseDefinition{"ok": &resp},
				Params:      &params,
			}
			res.Actions = map[string]*design.ActionDefinition{"get": &get}
			mt := design.MediaTypeDefinition{
				UserTypeDefinition: &ut,
				Identifier:         "vnd.rightscale.goagen.test.widgets",
			}
			design.Design = &design.APIDefinition{
				Name:        "test api",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
				Resources:   map[string]*design.ResourceDefinition{"Widget": &res},
				MediaTypes:  map[string]*design.MediaTypeDefinition{"vnd.rightscale.goagen.test.widgets": &mt},
			}
		})

		It("generates the corresponding code", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(3))
			data := map[string]string{"outDir": outDir, "design": "foo"}
			contextsCodeT, err := template.New("context").Parse(contextsCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			var b bytes.Buffer
			err = contextsCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			contextsCode := b.String()

			handlersCodeT, err := template.New("handlers").Parse(handlersCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			b.Reset()
			err = handlersCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			handlersCode := b.String()

			resourcesCodeT, err := template.New("resources").Parse(resourcesCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			b.Reset()
			err = resourcesCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			resourcesCode := b.String()

			isSource := func(filename, content string) {
				contextsContent, err := ioutil.ReadFile(filepath.Join(outDir, "app", filename))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(contextsContent)).Should(Equal(content))
			}

			isSource("contexts.go", contextsCode)
			isSource("handlers.go", handlersCode)
			isSource("resources.go", resourcesCode)
		})
	})
})

const contextsCodeTmpl = `//************************************************************************//
// test api: Application Contexts
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out={{.outDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

// GetWidgetContext provides the Widget get action context.
type GetWidgetContext struct {
	goa.Context
	Id string
}

// NewGetWidgetContext parses the incoming request URL and body, performs validations and creates the
// context used by the Widget controller get action.
func NewGetWidgetContext(c goa.Context) (*GetWidgetContext, error) {
	var err error
	ctx := GetWidgetContext{Context: c}
	rawId, ok := c.Get("id")
	if !ok {
		err = goa.MissingParamError("id", err)
	} else {
		ctx.Id = rawId
	}
	return &ctx, err
}

// OK sends a HTTP response with status code 200.
func (c *GetWidgetContext) OK(resp *id) error {
	return c.JSON(200, resp)
}
`

const handlersCodeTmpl = `//************************************************************************//
// test api: Application Handlers
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out={{.outDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

func init() {
	goa.RegisterHandlers(
		&goa.HandlerFactory{"Widget", "get", "GET", "/:id", getWidgetsHandler},
	)
}

func getWidgetsHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *GetWidgetContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action get Widget, expected 'func(c *GetWidgetContext) error'")
	}
	return func(c goa.Context) error {
		ctx, err := NewGetWidgetContext(c)
		if err != nil {
			return err
		}
		return h(ctx)
	}, nil
}
`

const resourcesCodeTmpl = `//************************************************************************//
// test api: Application Resources
//
// Generated with goagen v0.0.1, command line:
// $ goagen
// --out={{.outDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

// Widgetty
// Media type: vnd.rightscale.goagen.test.widgets
type Widget string

// WidgetHref returns the resource href.
func WidgetHref(string) string {
	return fmt.Sprintf("/:id")
}
`
