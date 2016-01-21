package dsl_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Attribute", func() {
	var name string
	var dataType DataType
	var description string
	var dsl func()

	var parent *AttributeDefinition

	BeforeEach(func() {
		InitDesign()
		Errors = nil
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
		RunDSL()
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
			Ω(o[name].Validations).Should(HaveLen(1))
			Ω(o[name].Validations[0]).Should(BeAssignableToTypeOf(&EnumValidationDefinition{}))
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
			Ω(o[name].Validations).Should(HaveLen(1))
			Ω(o[name].Validations[0]).Should(BeAssignableToTypeOf(&EnumValidationDefinition{}))
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
			Ω(o[name].Validations).Should(HaveLen(1))
			Ω(o[name].Validations[0]).Should(BeAssignableToTypeOf(&EnumValidationDefinition{}))
			Ω(o[name].Description).Should(Equal(description))
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
				Ω(Errors).Should(HaveOccurred())
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
				Ω(Errors).ShouldNot(HaveOccurred())
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
