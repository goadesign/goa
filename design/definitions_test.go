package design_test

import (
	"path"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsRequired", func() {
	var required string
	var attName string

	var attribute *design.AttributeDefinition
	var res bool

	JustBeforeEach(func() {
		integer := &design.AttributeDefinition{Type: design.Integer}
		attribute = &design.AttributeDefinition{
			Type:       design.Object{required: integer},
			Validation: &dslengine.ValidationDefinition{Required: []string{required}},
		}
		res = attribute.IsRequired(attName)
	})

	Context("called on a required field", func() {
		BeforeEach(func() {
			attName = "required"
			required = "required"
		})

		It("returns true", func() {
			Ω(res).Should(BeTrue())
		})
	})

	Context("called on a non-required field", func() {
		BeforeEach(func() {
			attName = "non-required"
			required = "required"
		})

		It("returns false", func() {
			Ω(res).Should(BeFalse())
		})
	})
})

var _ = Describe("IterateHeaders", func() {
	It("works when Parent.Headers is nil", func() {
		// create a Resource with no headers, Action with one header
		resource := &design.ResourceDefinition{}
		action := &design.ActionDefinition{
			Parent: resource,
			Headers: &design.AttributeDefinition{
				Type: design.Object{
					"a": &design.AttributeDefinition{Type: design.String},
				},
			},
		}
		names := []string{}
		// iterator that collects header names
		it := func(name string, _ bool, _ *design.AttributeDefinition) error {
			names = append(names, name)
			return nil
		}
		Ω(action.IterateHeaders(it)).Should(Succeed(), "despite action.Parent.Headers being nil")
		Ω(names).Should(ConsistOf("a"))
	})

})
var _ = Describe("Finalize ActionDefinition", func() {
	Context("with an action with no response", func() {
		var action *design.ActionDefinition

		BeforeEach(func() {
			// create a Resource with responses, Action with no response
			resource := &design.ResourceDefinition{
				Responses: map[string]*design.ResponseDefinition{
					"NotFound": &design.ResponseDefinition{Name: "NotFound", Status: 404},
				},
			}
			action = &design.ActionDefinition{Parent: resource}
		})

		It("does not panic and merges the resource responses", func() {
			Ω(action.Finalize).ShouldNot(Panic())
			Ω(action.Responses).Should(HaveKey("NotFound"))
		})
	})
})

var _ = Describe("FullPath", func() {

	Context("Given a base resource and a resource with an action with a route", func() {
		var resource, parentResource *design.ResourceDefinition
		var action *design.ActionDefinition
		var route *design.RouteDefinition

		var actionPath string
		var resourcePath string
		var parentResourcePath string

		JustBeforeEach(func() {
			showAct := &design.ActionDefinition{}
			showRoute := &design.RouteDefinition{
				Path:   parentResourcePath,
				Parent: showAct,
			}
			showAct.Routes = []*design.RouteDefinition{showRoute}
			parentResource = &design.ResourceDefinition{}
			parentResource.Actions = map[string]*design.ActionDefinition{"show": showAct}
			parentResource.Name = "foo"
			design.Design.Resources = map[string]*design.ResourceDefinition{"foo": parentResource}
			showAct.Parent = parentResource

			action = &design.ActionDefinition{}
			route = &design.RouteDefinition{
				Path:   actionPath,
				Parent: action,
			}
			action.Routes = []*design.RouteDefinition{route}
			resource = &design.ResourceDefinition{}
			resource.Actions = map[string]*design.ActionDefinition{"action": action}
			resource.BasePath = resourcePath
			resource.ParentName = parentResource.Name
			action.Parent = resource
		})

		Context("with relative routes", func() {
			BeforeEach(func() {
				actionPath = "/action"
				resourcePath = "/resource"
				parentResourcePath = "/parent"
			})

			It("FullPath concatenates them", func() {
				Ω(route.FullPath()).Should(Equal(path.Join(parentResourcePath, resourcePath, actionPath)))
			})

			Context("with an action with absolute route", func() {
				BeforeEach(func() {
					actionPath = "//action"
				})

				It("FullPath uses it", func() {
					Ω(route.FullPath()).Should(Equal(actionPath[1:]))
				})
			})

			Context("with n resource with absolute route", func() {
				BeforeEach(func() {
					resourcePath = "//resource"
				})

				It("FullPath uses it", func() {
					Ω(route.FullPath()).Should(Equal(resourcePath[1:] + "/" + actionPath[1:]))
				})
			})
		})
	})
})

var _ = Describe("AllParams", func() {
	Context("Given a resource with a parent and an action with a route", func() {
		var (
			resource, parent *design.ResourceDefinition
			action           *design.ActionDefinition
			allParams        design.Object
		)

		BeforeEach(func() {
			// Parent resource
			{
				baseParams := &design.AttributeDefinition{Type: design.Object{
					"pbasepath":  &design.AttributeDefinition{Type: design.String},
					"pbasequery": &design.AttributeDefinition{Type: design.String},
				}}
				parent = &design.ResourceDefinition{
					Name:                "parent",
					CanonicalActionName: "canonical",
					BasePath:            "/:pbasepath",
					Params:              baseParams,
				}
				canParams := &design.AttributeDefinition{Type: design.Object{
					"canpath":  &design.AttributeDefinition{Type: design.String},
					"canquery": &design.AttributeDefinition{Type: design.String},
				}}
				canonical := &design.ActionDefinition{
					Name:   "canonical",
					Parent: parent,
					Params: canParams,
				}
				croute := &design.RouteDefinition{
					Path:   "/:canpath",
					Parent: canonical,
				}
				canonical.Routes = []*design.RouteDefinition{croute}
				parent.Actions = map[string]*design.ActionDefinition{"canonical": canonical}
			}

			// Resource
			{
				baseParams := &design.AttributeDefinition{Type: design.Object{
					"basepath":  &design.AttributeDefinition{Type: design.String},
					"basequery": &design.AttributeDefinition{Type: design.String},
				}}
				resource = &design.ResourceDefinition{
					Name:       "child",
					ParentName: "parent",
					BasePath:   "/:basepath",
					Params:     baseParams,
				}
			}

			// Action
			{
				params := &design.AttributeDefinition{Type: design.Object{
					"path":     &design.AttributeDefinition{Type: design.String},
					"query":    &design.AttributeDefinition{Type: design.String},
					"basepath": &design.AttributeDefinition{Type: design.String},
				}}
				action = &design.ActionDefinition{
					Name:   "action",
					Parent: resource,
					Params: params,
				}
				route := &design.RouteDefinition{
					Path:   "/:path",
					Parent: action,
				}
				action.Routes = []*design.RouteDefinition{route}
				resource.Actions = map[string]*design.ActionDefinition{"action": action}
			}
			design.Design.Resources = map[string]*design.ResourceDefinition{"resource": resource, "parent": parent}
		})

		JustBeforeEach(func() {
			allParams = action.AllParams().Type.ToObject()
			Ω(allParams).ShouldNot(BeNil())
		})

		It("returns both path and query parameters", func() {
			for p := range action.Params.Type.ToObject() {
				Ω(allParams).Should(HaveKey(p))
			}
		})

		It("returns all path parameters recursively", func() {
			for _, p := range []string{"path", "basepath", "canpath", "pbasepath"} {
				Ω(allParams).Should(HaveKey(p))
			}
		})

		It("does not return the query parameters of the parent resource canonical action", func() {
			for _, p := range []string{"basequery", "canquery", "pbasequery"} {
				Ω(allParams).ShouldNot(HaveKey(p))
			}
		})
	})
})
