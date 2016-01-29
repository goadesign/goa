package design_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validation", func() {
	Context("with a type attribute", func() {
		const attName = "attName"
		var dsl func()

		var att *AttributeDefinition

		JustBeforeEach(func() {
			InitDesign()
			dslengine.Errors = nil
			Type("bar", func() {
				dsl()
			})
			dslengine.Run()
			if dslengine.Errors == nil {
				Ω(Design.Types).ShouldNot(BeNil())
				Ω(Design.Types).Should(HaveKey("bar"))
				Ω(Design.Types["bar"]).ShouldNot(BeNil())
				Ω(Design.Types["bar"].Type).Should(BeAssignableToTypeOf(Object{}))
				o := Design.Types["bar"].Type.(Object)
				Ω(o).Should(HaveKey(attName))
				att = o[attName]
			}
		})

		Context("with a valid enum validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Enum("red", "blue")
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.EnumValidationDefinition{}))
				expected := &dslengine.EnumValidationDefinition{Values: []interface{}{"red", "blue"}}
				Ω(v.(*dslengine.EnumValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an incompatible enum validation type", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Enum(1, "blue")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid format validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Format("email")
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.FormatValidationDefinition{}))
				expected := &dslengine.FormatValidationDefinition{Format: "email"}
				Ω(v.(*dslengine.FormatValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid format validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Format("emailz")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid pattern validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Pattern("^foo$")
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.PatternValidationDefinition{}))
				expected := &dslengine.PatternValidationDefinition{Pattern: "^foo$"}
				Ω(v.(*dslengine.PatternValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid pattern validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Pattern("[invalid")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with an invalid format validation type", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Format("email")
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid min value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Minimum(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.MinimumValidationDefinition{}))
				expected := &dslengine.MinimumValidationDefinition{Min: 2}
				Ω(v.(*dslengine.MinimumValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid min value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Minimum(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid max value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						Maximum(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.MaximumValidationDefinition{}))
				expected := &dslengine.MaximumValidationDefinition{Max: 2}
				Ω(v.(*dslengine.MaximumValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid max value validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						Maximum(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid min length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, ArrayOf(Integer), func() {
						MinLength(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.MinLengthValidationDefinition{}))
				expected := &dslengine.MinLengthValidationDefinition{MinLength: 2}
				Ω(v.(*dslengine.MinLengthValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid min length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						MinLength(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a valid max length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String, func() {
						MaxLength(2)
					})
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.MaxLengthValidationDefinition{}))
				expected := &dslengine.MaxLengthValidationDefinition{MaxLength: 2}
				Ω(v.(*dslengine.MaxLengthValidationDefinition)).Should(Equal(expected))
			})
		})

		Context("with an invalid max length validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, Integer, func() {
						MaxLength(2)
					})
				}
			})

			It("produces an error", func() {
				Ω(dslengine.Errors).Should(HaveOccurred())
			})
		})

		Context("with a required field validation", func() {
			BeforeEach(func() {
				dsl = func() {
					Attribute(attName, String)
					Required(attName)
				}
			})

			It("records the validation", func() {
				Ω(dslengine.Errors).ShouldNot(HaveOccurred())
				Ω(Design.Types["bar"].Validations).Should(HaveLen(1))
				v := Design.Types["bar"].Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&dslengine.RequiredValidationDefinition{}))
				expected := &dslengine.RequiredValidationDefinition{Names: []string{attName}}
				Ω(v.(*dslengine.RequiredValidationDefinition)).Should(Equal(expected))
			})
		})
	})
})
