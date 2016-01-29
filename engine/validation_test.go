package engine_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/dsl"
	"github.com/goadesign/goa/engine"
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
			engine.Errors = nil
			Type("bar", func() {
				dsl()
			})
			engine.RunDSL()
			if engine.Errors == nil {
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.EnumValidationDefinition{}))
				expected := &engine.EnumValidationDefinition{Values: []interface{}{"red", "blue"}}
				Ω(v.(*engine.EnumValidationDefinition)).Should(Equal(expected))
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.FormatValidationDefinition{}))
				expected := &engine.FormatValidationDefinition{Format: "email"}
				Ω(v.(*engine.FormatValidationDefinition)).Should(Equal(expected))
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.PatternValidationDefinition{}))
				expected := &engine.PatternValidationDefinition{Pattern: "^foo$"}
				Ω(v.(*engine.PatternValidationDefinition)).Should(Equal(expected))
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.MinimumValidationDefinition{}))
				expected := &engine.MinimumValidationDefinition{Min: 2}
				Ω(v.(*engine.MinimumValidationDefinition)).Should(Equal(expected))
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.MaximumValidationDefinition{}))
				expected := &engine.MaximumValidationDefinition{Max: 2}
				Ω(v.(*engine.MaximumValidationDefinition)).Should(Equal(expected))
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.MinLengthValidationDefinition{}))
				expected := &engine.MinLengthValidationDefinition{MinLength: 2}
				Ω(v.(*engine.MinLengthValidationDefinition)).Should(Equal(expected))
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(att.Validations).Should(HaveLen(1))
				v := att.Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.MaxLengthValidationDefinition{}))
				expected := &engine.MaxLengthValidationDefinition{MaxLength: 2}
				Ω(v.(*engine.MaxLengthValidationDefinition)).Should(Equal(expected))
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
				Ω(engine.Errors).Should(HaveOccurred())
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
				Ω(engine.Errors).ShouldNot(HaveOccurred())
				Ω(Design.Types["bar"].Validations).Should(HaveLen(1))
				v := Design.Types["bar"].Validations[0]
				Ω(v).Should(BeAssignableToTypeOf(&engine.RequiredValidationDefinition{}))
				expected := &engine.RequiredValidationDefinition{Names: []string{attName}}
				Ω(v.(*engine.RequiredValidationDefinition)).Should(Equal(expected))
			})
		})
	})
})
