package apidsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// TestCD is a test container definition.
type TestCD struct {
	*AttributeDefinition
}

// Attribute returns a dummy attribute.
func (t *TestCD) Attribute() *AttributeDefinition {
	return t.AttributeDefinition
}

// DSL implements Source
func (t *TestCD) DSL() func() {
	return func() {
		Attribute("foo")
	}
}

// Context implement Definition
func (t *TestCD) Context() string {
	return "test"
}

// DSLName returns the DSL name.
func (t *TestCD) DSLName() string {
	return "TestCD"
}

// DependsOn returns the DSL dependencies.
func (t *TestCD) DependsOn() []dslengine.Root {
	return nil
}

// IterateSets implement Root
func (t *TestCD) IterateSets(it dslengine.SetIterator) {
	it([]dslengine.Definition{t})
}

// Reset is a no-op
func (t *TestCD) Reset() {}

var _ = Describe("ContainerDefinition", func() {
	var att *AttributeDefinition
	var testCD *TestCD
	BeforeEach(func() {
		dslengine.Reset()
		att = &AttributeDefinition{Type: Object{}}
		testCD = &TestCD{AttributeDefinition: att}
		dslengine.Register(testCD)
	})

	JustBeforeEach(func() {
		err := dslengine.Run()
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("contains attributes", func() {
		Ω(testCD.Attribute()).Should(Equal(att))
	})
})

var _ = Describe("Attribute", func() {
	var name string
	var dataType interface{}
	var description string
	var dsl func()

	var parent *AttributeDefinition

	BeforeEach(func() {
		dslengine.Reset()
		name = ""
		dataType = nil
		description = ""
		dsl = nil
	})

	JustBeforeEach(func() {
		Type("type", func() {
			if dsl == nil {
				if dataType == nil {
					Attribute(name)
				} else if description == "" {
					Attribute(name, dataType)
				} else {
					Attribute(name, dataType, description)
				}
			} else if dataType == nil {
				Attribute(name, dsl)
			} else if description == "" {
				Attribute(name, dataType, dsl)
			} else {
				Attribute(name, dataType, description, dsl)
			}
		})
		dslengine.Run()
		if t, ok := Design.Types["type"]; ok {
			parent = t.AttributeDefinition
		}
	})

	Context("with only a name", func() {
		BeforeEach(func() {
			name = "foo"
		})

		It("produces an attribute of type string", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(String))
		})
	})

	Context("with a name and datatype", func() {
		BeforeEach(func() {
			name = "foo"
			dataType = Integer
		})

		It("produces an attribute of given type", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(Integer))
		})
	})

	Context("with a name and uuid datatype", func() {
		BeforeEach(func() {
			name = "foo"
			dataType = UUID
		})

		It("produces an attribute of uuid type", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(UUID))
		})
	})

	Context("with a name and date datatype", func() {
		BeforeEach(func() {
			name = "foo"
			dataType = DateTime
		})

		It("produces an attribute of date type", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(DateTime))
		})
	})

	Context("with a name, datatype and description", func() {
		BeforeEach(func() {
			name = "foo"
			dataType = Integer
			description = "bar"
		})

		It("produces an attribute of given type and given description", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(Integer))
			Ω(o[name].Description).Should(Equal(description))
		})
	})

	Context("with a name and a DSL defining an enum validation", func() {
		BeforeEach(func() {
			name = "foo"
			dsl = func() { Enum("one", "two") }
		})

		It("produces an attribute of type string with a validation", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(String))
			Ω(o[name].Validation).ShouldNot(BeNil())
			Ω(o[name].Validation.Values).Should(Equal([]interface{}{"one", "two"}))
		})
	})

	Context("with a name, type datetime and a DSL defining a default value", func() {
		BeforeEach(func() {
			name = "foo"
			dataType = DateTime
			dsl = func() { Default("1978-06-30T10:00:00+09:00") }
		})

		It("produces an attribute of type string with a default value", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(DateTime))
			Ω(o[name].Validation).Should(BeNil())
			Ω(o[name].DefaultValue).Should(Equal(interface{}("1978-06-30T10:00:00+09:00")))
		})
	})

	Context("with a name, type integer and a DSL defining an enum validation", func() {
		BeforeEach(func() {
			name = "foo"
			dataType = Integer
			dsl = func() { Enum(1, 2) }
		})

		It("produces an attribute of type integer with a validation", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(Integer))
			Ω(o[name].Validation).ShouldNot(BeNil())
			Ω(o[name].Validation.Values).Should(Equal([]interface{}{1, 2}))
		})
	})

	Context("with a name, type integer, a description and a DSL defining an enum validation", func() {
		BeforeEach(func() {
			name = "foo"
			dataType = String
			description = "bar"
			dsl = func() { Enum("one", "two") }
		})

		It("produces an attribute of type integer with a validation and the description", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(String))
			Ω(o[name].Validation).ShouldNot(BeNil())
			Ω(o[name].Validation.Values).Should(Equal([]interface{}{"one", "two"}))
		})
	})

	Context("with a name and type uuid", func() {
		BeforeEach(func() {
			name = "birthdate"
			dataType = UUID
		})

		It("produces an attribute of type date with a validation and the description", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(UUID))
		})
	})

	Context("with a name and type date", func() {
		BeforeEach(func() {
			name = "birthdate"
			dataType = DateTime
		})

		It("produces an attribute of type date with a validation and the description", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(DateTime))
		})
	})

	Context("with a name and a type defined by name", func() {
		var Foo *UserTypeDefinition

		BeforeEach(func() {
			name = "fooatt"
			dataType = "foo"
			Foo = Type("foo", func() {
				Attribute("bar")
			})
		})

		It("produces an attribute of the corresponding type", func() {
			t := parent.Type
			Ω(t).ShouldNot(BeNil())
			Ω(t).Should(BeAssignableToTypeOf(Object{}))
			o := t.(Object)
			Ω(o).Should(HaveLen(1))
			Ω(o).Should(HaveKey(name))
			Ω(o[name].Type).Should(Equal(Foo))
		})
	})

	Context("with child attributes", func() {
		const childAtt = "childAtt"

		BeforeEach(func() {
			name = "foo"
			dsl = func() { Attribute(childAtt) }
		})

		Context("on an attribute that is not an object", func() {
			BeforeEach(func() {
				dataType = Integer
			})

			It("fails", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("on an attribute that does not have a type", func() {
			It("sets the type to Object", func() {
				t := parent.Type
				Ω(t).ShouldNot(BeNil())
				Ω(t).Should(BeAssignableToTypeOf(Object{}))
				o := t.(Object)
				Ω(o).Should(HaveLen(1))
				Ω(o).Should(HaveKey(name))
				Ω(o[name].Type).Should(BeAssignableToTypeOf(Object{}))
			})
		})

		Context("on an attribute of type Object", func() {
			BeforeEach(func() {
				dataType = Object{}
			})

			It("initializes the object attributes", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				t := parent.Type
				Ω(t).ShouldNot(BeNil())
				Ω(t).Should(BeAssignableToTypeOf(Object{}))
				o := t.(Object)
				Ω(o).Should(HaveLen(1))
				Ω(o).Should(HaveKey(name))
				Ω(o[name].Type).Should(BeAssignableToTypeOf(Object{}))
				co := o[name].Type.(Object)
				Ω(co).Should(HaveLen(1))
				Ω(co).Should(HaveKey(childAtt))
			})
		})
	})
})
