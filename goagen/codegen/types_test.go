package codegen_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("code generation", func() {
	BeforeEach(func() {
		codegen.TempCount = 0
	})

	Describe("Goify", func() {
		Context("given a string with an initialism", func() {
			var str, goified, expected string
			var firstUpper bool
			JustBeforeEach(func() {
				goified = codegen.Goify(str, firstUpper)
			})

			Context("with first upper false", func() {
				BeforeEach(func() {
					firstUpper = false
					str = "blue_id"
					expected = "blueID"
				})
				It("creates a lowercased camelcased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with first upper false normal identifier", func() {
				BeforeEach(func() {
					firstUpper = false
					str = "blue"
					expected = "blue"
				})
				It("creates an uppercased camelcased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with first upper false and UUID", func() {
				BeforeEach(func() {
					firstUpper = false
					str = "blue_uuid"
					expected = "blueUUID"
				})
				It("creates an uppercased camelcased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with first upper true", func() {
				BeforeEach(func() {
					firstUpper = true
					str = "blue_id"
					expected = "BlueID"
				})
				It("creates an uppercased camelcased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with first upper true and UUID", func() {
				BeforeEach(func() {
					firstUpper = true
					str = "blue_uuid"
					expected = "BlueUUID"
				})
				It("creates an uppercased camelcased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with first upper true normal identifier", func() {
				BeforeEach(func() {
					firstUpper = true
					str = "blue"
					expected = "Blue"
				})
				It("creates an uppercased camelcased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with first upper false normal identifier", func() {
				BeforeEach(func() {
					firstUpper = false
					str = "Blue"
					expected = "blue"
				})
				It("creates a lowercased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with first upper true normal identifier", func() {
				BeforeEach(func() {
					firstUpper = true
					str = "Blue"
					expected = "Blue"
				})
				It("creates an uppercased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})
			Context("with invalid identifier", func() {
				BeforeEach(func() {
					firstUpper = true
					str = "Blue%50"
					expected = "Blue50"
				})
				It("creates an uppercased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})

			Context("with invalid identifier firstupper false", func() {
				BeforeEach(func() {
					firstUpper = false
					str = "Blue%50"
					expected = "blue50"
				})
				It("creates an uppercased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})

			Context("with only UUID and firstupper false", func() {
				BeforeEach(func() {
					firstUpper = false
					str = "UUID"
					expected = "uuid"
				})
				It("creates a lowercased string", func() {
					Ω(goified).Should(Equal(expected))
				})
			})

		})

	})

	Describe("GoTypeDef", func() {
		Context("given an attribute definition with fields", func() {
			var att *AttributeDefinition
			var object Object
			var required *dslengine.ValidationDefinition
			var st string

			JustBeforeEach(func() {
				att = new(AttributeDefinition)
				att.Type = object
				if required != nil {
					att.Validation = required
				}
				st = codegen.GoTypeDef(att, 0, true, false)
			})

			Context("of primitive types", func() {
				BeforeEach(func() {
					object = Object{
						"foo": &AttributeDefinition{Type: Integer},
						"bar": &AttributeDefinition{Type: String},
						"baz": &AttributeDefinition{Type: DateTime},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Bar *string `json:\"bar,omitempty\" xml:\"bar,omitempty\"`\n" +
						"	Baz *time.Time `json:\"baz,omitempty\" xml:\"baz,omitempty\"`\n" +
						"	Foo *int `json:\"foo,omitempty\" xml:\"foo,omitempty\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})
			})

			Context("of hash of primitive types", func() {
				BeforeEach(func() {
					elemType := &AttributeDefinition{Type: Integer}
					keyType := &AttributeDefinition{Type: Integer}
					hash := &Hash{KeyType: keyType, ElemType: elemType}
					object = Object{
						"foo": &AttributeDefinition{Type: hash},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					Ω(st).Should(Equal("struct {\n\tFoo map[int]int `json:\"foo,omitempty\" xml:\"foo,omitempty\"`\n}"))
				})
			})

			Context("of array of primitive types", func() {
				BeforeEach(func() {
					elemType := &AttributeDefinition{Type: Integer}
					array := &Array{ElemType: elemType}
					object = Object{
						"foo": &AttributeDefinition{Type: array},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					Ω(st).Should(Equal("struct {\n\tFoo []int `json:\"foo,omitempty\" xml:\"foo,omitempty\"`\n}"))
				})
			})

			Context("of hash of objects", func() {
				BeforeEach(func() {
					elem := Object{
						"elemAtt": &AttributeDefinition{Type: Integer},
					}
					key := Object{
						"keyAtt": &AttributeDefinition{Type: String},
					}
					elemType := &AttributeDefinition{Type: elem}
					keyType := &AttributeDefinition{Type: key}
					hash := &Hash{KeyType: keyType, ElemType: elemType}
					object = Object{
						"foo": &AttributeDefinition{Type: hash},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Foo map[*struct {\n" +
						"		KeyAtt *string `json:\"keyAtt,omitempty\" xml:\"keyAtt,omitempty\"`\n" +
						"	}]*struct {\n" +
						"		ElemAtt *int `json:\"elemAtt,omitempty\" xml:\"elemAtt,omitempty\"`\n" +
						"	} `json:\"foo,omitempty\" xml:\"foo,omitempty\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})
			})

			Context("of array of objects", func() {
				BeforeEach(func() {
					obj := Object{
						"bar": &AttributeDefinition{Type: Integer},
					}
					elemType := &AttributeDefinition{Type: obj}
					array := &Array{ElemType: elemType}
					object = Object{
						"foo": &AttributeDefinition{Type: array},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Foo []*struct {\n" +
						"		Bar *int `json:\"bar,omitempty\" xml:\"bar,omitempty\"`\n" +
						"	} `json:\"foo,omitempty\" xml:\"foo,omitempty\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})

				Context("that are required", func() {
					BeforeEach(func() {
						required = &dslengine.ValidationDefinition{
							Required: []string{"foo"},
						}
					})

					It("produces the struct go code", func() {
						expected := "struct {\n" +
							"	Foo []*struct {\n" +
							"		Bar *int `json:\"bar,omitempty\" xml:\"bar,omitempty\"`\n" +
							"	} `json:\"foo\" xml:\"foo\"`\n" +
							"}"
						Ω(st).Should(Equal(expected))
					})
				})
			})

			Context("that are required", func() {
				BeforeEach(func() {
					object = Object{
						"foo": &AttributeDefinition{Type: Integer},
					}
					required = &dslengine.ValidationDefinition{
						Required: []string{"foo"},
					}
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Foo int `json:\"foo\" xml:\"foo\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})
			})

		})

		Context("given an array", func() {
			var elemType *AttributeDefinition
			var source string

			JustBeforeEach(func() {
				array := &Array{ElemType: elemType}
				att := &AttributeDefinition{Type: array}
				source = codegen.GoTypeDef(att, 0, true, false)
			})

			Context("of primitive type", func() {
				BeforeEach(func() {
					elemType = &AttributeDefinition{Type: Integer}
				})

				It("produces the array go code", func() {
					Ω(source).Should(Equal("[]int"))
				})

			})

			Context("of object type", func() {
				BeforeEach(func() {
					object := Object{
						"foo": &AttributeDefinition{Type: Integer},
						"bar": &AttributeDefinition{Type: String},
					}
					elemType = &AttributeDefinition{Type: object}
				})

				It("produces the array go code", func() {
					Ω(source).Should(Equal("[]*struct {\n\tBar *string `json:\"bar,omitempty\" xml:\"bar,omitempty\"`\n\tFoo *int `json:\"foo,omitempty\" xml:\"foo,omitempty\"`\n}"))
				})
			})
		})

	})
})

var _ = Describe("GoTypeTransform", func() {
	var source, target *UserTypeDefinition
	var targetPkg, funcName string

	var transform string

	BeforeEach(func() {
		dslengine.Reset()
	})
	JustBeforeEach(func() {
		err := dslengine.Run()
		Ω(err).ShouldNot(HaveOccurred())
		transform, _ = codegen.GoTypeTransform(source, target, targetPkg, funcName)
	})

	Context("transforming simple objects", func() {
		const attName = "att"
		BeforeEach(func() {
			source = Type("Source", func() {
				Attribute(attName)
			})
			target = Type("Target", func() {
				Attribute(attName)
			})
			funcName = "Transform"
		})

		It("generates a simple assignment", func() {
			Ω(transform).Should(Equal(`func Transform(source *Source) (target *Target) {
	target = new(Target)
	target.Att = source.Att
	return
}
`))
		})
	})

	Context("transforming objects with attributes with map key metadata", func() {
		const mapKey = "key"
		BeforeEach(func() {
			source = Type("Source", func() {
				Attribute("foo", func() {
					Metadata(codegen.TransformMapKey, mapKey)
				})
			})
			target = Type("Target", func() {
				Attribute("bar", func() {
					Metadata(codegen.TransformMapKey, mapKey)
				})
			})
			funcName = "Transform"
		})

		It("generates a simple assignment", func() {
			Ω(transform).Should(Equal(`func Transform(source *Source) (target *Target) {
	target = new(Target)
	target.Bar = source.Foo
	return
}
`))
		})
	})

	Context("transforming objects with array attributes", func() {
		const attName = "att"
		BeforeEach(func() {
			source = Type("Source", func() {
				Attribute(attName, ArrayOf(Integer))
			})
			target = Type("Target", func() {
				Attribute(attName, ArrayOf(Integer))
			})
			funcName = "Transform"
		})

		It("generates a simple assignment", func() {
			Ω(transform).Should(Equal(`func Transform(source *Source) (target *Target) {
	target = new(Target)
	target.Att = make([]int, len(source.Att))
	for i, v := range source.Att {
		target.Att[i] = source.Att[i]
	}
	return
}
`))
		})
	})

	Context("transforming objects with hash attributes", func() {
		const attName = "att"
		BeforeEach(func() {
			elem := Type("elem", func() {
				Attribute("foo", Integer)
				Attribute("bar")
			})
			source = Type("Source", func() {
				Attribute(attName, HashOf(String, elem))
			})
			target = Type("Target", func() {
				Attribute(attName, HashOf(String, elem))
			})
			funcName = "Transform"
		})

		It("generates a simple assignment", func() {
			Ω(transform).Should(Equal(`func Transform(source *Source) (target *Target) {
	target = new(Target)
	target.Att = make(map[string]*Elem, len(source.Att))
	for k, v := range source.Att {
		var tk string
		tk = k
		var tv *Elem
		tv = new(Elem)
		tv.Bar = v.Bar
		tv.Foo = v.Foo
		target.Att[tk] = tv
	}
	return
}
`))
		})
	})

	Context("transforming objects with recursive attributes", func() {
		const attName = "att"
		BeforeEach(func() {
			inner := Type("inner", func() {
				Attribute("foo", Integer)
			})
			outer := Type("outer", func() {
				Attribute("in", inner)
			})
			array := Type("array", func() {
				Attribute("elem", ArrayOf(outer))
			})
			hash := Type("hash", func() {
				Attribute("elem", HashOf(Integer, outer))
			})
			source = Type("Source", func() {
				Attribute("outer", outer)
				Attribute("array", array)
				Attribute("hash", hash)
			})
			target = Type("Target", func() {
				Attribute("outer", outer)
				Attribute("array", array)
				Attribute("hash", hash)
			})
			funcName = "Transform"
		})

		It("generates the proper assignments", func() {
			Ω(transform).Should(Equal(`func Transform(source *Source) (target *Target) {
	target = new(Target)
	target.Array = new(Array)
	target.Array.Elem = make([]*Outer, len(source.Array.Elem))
	for i, v := range source.Array.Elem {
		target.Array.Elem[i] = new(Outer)
		target.Array.Elem[i].In = new(Inner)
		target.Array.Elem[i].In.Foo = source.Array.Elem[i].In.Foo
	}
	target.Hash = new(Hash)
	target.Hash.Elem = make(map[int]*Outer, len(source.Hash.Elem))
	for k, v := range source.Hash.Elem {
		var tk int
		tk = k
		var tv *Outer
		tv = new(Outer)
		tv.In = new(Inner)
		tv.In.Foo = v.In.Foo
		target.Hash.Elem[tk] = tv
	}
	target.Outer = new(Outer)
	target.Outer.In = new(Inner)
	target.Outer.In.Foo = source.Outer.In.Foo
	return
}
`))
		})
	})
})
