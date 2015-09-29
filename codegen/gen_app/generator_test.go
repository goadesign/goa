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
	"github.com/raphael/goa/codegen/gen_app"
	"github.com/raphael/goa/design"
)

var _ = Describe("NewGenerator", func() {
	var gen *genapp.Generator

	Context("with dummy command line flags", func() {
		BeforeEach(func() {
			os.Args = []string{"codegen", "--out=_foo", "--design=bar", "--force"}
		})

		AfterEach(func() {
			os.RemoveAll("_foo")
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
		os.Args = []string{"codegen", "--out=" + outDir, "--design=foo"}
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
			Ω(files).Should(HaveLen(4))
			isEmptySource := func(filename string) {
				contextsContent, err := ioutil.ReadFile(filepath.Join(outDir, "app", filename))
				Ω(err).ShouldNot(HaveOccurred())
				lines := strings.Split(string(contextsContent), "\n")
				Ω(lines).ShouldNot(BeEmpty())
				Ω(len(lines)).Should(BeNumerically(">", 1))
			}
			isEmptySource("contexts.go")
			isEmptySource("handlers.go")
			isEmptySource("resources.go")
			isEmptySource("media_types.go")
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
				MediaType:   "vnd.rightscale.codegen.test.widgets",
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
				MediaType:       "vnd.rightscale.codegen.test.widgets",
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
				Identifier:         "vnd.rightscale.codegen.test.widgets",
			}
			design.Design = &design.APIDefinition{
				Name:        "test api",
				Title:       "dummy API with no resource",
				Description: "I told you it's dummy",
				Resources:   map[string]*design.ResourceDefinition{"Widget": &res},
				MediaTypes:  map[string]*design.MediaTypeDefinition{"vnd.rightscale.codegen.test.widgets": &mt},
			}
		})

		It("generates the corresponding code", func() {
			Ω(genErr).Should(BeNil())
			Ω(files).Should(HaveLen(4))
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

			mediaTypesCodeT, err := template.New("media types").Parse(mediaTypesCodeTmpl)
			Ω(err).ShouldNot(HaveOccurred())
			b.Reset()
			err = mediaTypesCodeT.Execute(&b, data)
			Ω(err).ShouldNot(HaveOccurred())
			mediaTypesCode := b.String()

			isSource := func(filename, content string) {
				contextsContent, err := ioutil.ReadFile(filepath.Join(outDir, "app", filename))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(contextsContent)).Should(Equal(content))
			}

			isSource("contexts.go", contextsCode)
			isSource("handlers.go", handlersCode)
			isSource("resources.go", resourcesCode)
			isSource("media_types.go", mediaTypesCode)
		})
	})
})

const contextsCodeTmpl = `//************************************************************************//
// test api: Application Contexts
//
// Generated with codegen v0.0.1, command line:
// $ codegen
// --out={{.outDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import (
	"github.com/raphael/goa"
	"strconv"
)

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
// Generated with codegen v0.0.1, command line:
// $ codegen
// --out={{.outDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

import (
	"fmt"
	"github.com/raphael/goa"
)

func init() {
	goa.RegisterHandlers(
		&goa.HandlerFactory{"widgets", "get", "GET", "/:id", getWidgetsHandler},
	)
}

func getWidgetsHandler(userHandler interface{}) (goa.Handler, error) {
	h, ok := userHandler.(func(c *GetWidgetContext) error)
	if !ok {
		return nil, fmt.Errorf("invalid handler signature for action get widgets, expected 'func(c *GetWidgetContext) error'")
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
// Generated with codegen v0.0.1, command line:
// $ codegen
// --out={{.outDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

// Widgetty
type Widget string

// WidgetHref returns the resource href.
func WidgetHref() string {
	return fmt.Sprintf("/:id", )
}
`

const mediaTypesCodeTmpl = `//************************************************************************//
// test api: Application Media Types
//
// Generated with codegen v0.0.1, command line:
// $ codegen
// --out={{.outDir}}
// --design={{.design}}
//
// The content of this file is auto-generated, DO NOT MODIFY
//************************************************************************//

package app

// id media type
// Identifier: vnd.rightscale.codegen.test.widgets
type id string
`
