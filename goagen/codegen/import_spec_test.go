package codegen_test

import (
	. "github.com/goadesign/goa/design"
	. "github.com/goadesign/goa/design/apidsl"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/codegen"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AttributeImports", func() {
	Context("given an attribute definition with fields", func() {

		var imports []*codegen.ImportSpec
		var att *AttributeDefinition
		var st string
		var object Object

		Context("of object", func() {

			It("produces the import slice", func() {
				object = Object{
					"foo": &AttributeDefinition{Type: String},
					"bar": &AttributeDefinition{Type: Integer},
				}
				object["foo"].Metadata = dslengine.MetadataDefinition{
					"struct:field:type": []string{"json.RawMessage", "encoding/json"},
				}
				att = new(AttributeDefinition)
				att.Type = object
				imports = codegen.AttributeImports(att, imports, nil)

				i := []*codegen.ImportSpec{&codegen.ImportSpec{
					Path: "encoding/json",
				},
				}
				st = i[0].Path

				Ω(st).Should(Equal(imports[0].Path))
			})
		})

		Context("of hash", func() {

			It("produces the import slice", func() {
				elemType := &AttributeDefinition{Type: Integer}
				elemType.Metadata = dslengine.MetadataDefinition{
					"struct:field:type": []string{"json.RawMessage", "encoding/json"},
				}
				keyType := &AttributeDefinition{Type: Integer}
				elemType.Metadata = dslengine.MetadataDefinition{
					"struct:field:type": []string{"json.RawMessage", "encoding/json"},
				}
				hash := &Hash{KeyType: keyType, ElemType: elemType}

				att = new(AttributeDefinition)
				att.Type = hash
				imports = codegen.AttributeImports(att, imports, nil)

				i := []*codegen.ImportSpec{&codegen.ImportSpec{
					Path: "encoding/json",
				},
				}
				st = i[0].Path
				l := len(imports)

				Ω(st).Should(Equal(imports[0].Path))
				Ω(l).Should(Equal(1))
			})
		})

		Context("of array", func() {
			It("produces the import slice", func() {
				elemType := &AttributeDefinition{Type: Integer}
				elemType.Metadata = dslengine.MetadataDefinition{
					"struct:field:type": []string{"json.RawMessage", "encoding/json"},
				}
				array := &Array{ElemType: elemType}

				att = new(AttributeDefinition)
				att.Type = array
				imports = codegen.AttributeImports(att, imports, nil)

				i := []*codegen.ImportSpec{&codegen.ImportSpec{
					Path: "encoding/json",
				},
				}
				st = i[0].Path

				Ω(st).Should(Equal(imports[0].Path))
			})
		})

		Context("of UserTypeDefinition", func() {

			It("produces the struct go code", func() {
				var u *UserTypeDefinition
				object = Object{
					"bar": &AttributeDefinition{Type: String},
				}
				object["bar"].Metadata = dslengine.MetadataDefinition{
					"struct:field:type": []string{"json.RawMessage", "encoding/json"},
				}
				att = u.AttributeDefinition
				att.Type = object
				imports = codegen.AttributeImports(att, imports, nil)

				i := []*codegen.ImportSpec{&codegen.ImportSpec{
					Path: "encoding/json",
				},
				}
				st = i[0].Path

				Ω(st).Should(Equal(imports[0].Path))
			})
		})

		Context("of MediaTypeDefinition", func() {
			It("produces the struct go code", func() {
				var m *MediaTypeDefinition
				elemType := &AttributeDefinition{Type: Integer}
				elemType.Metadata = dslengine.MetadataDefinition{
					"struct:field:type": []string{"json.RawMessage", "encoding/json"},
				}
				array := &Array{ElemType: elemType}

				att = m.AttributeDefinition
				att.Type = array
				imports = codegen.AttributeImports(att, imports, nil)

				i := []*codegen.ImportSpec{&codegen.ImportSpec{
					Path: "encoding/json",
				},
				}
				st = i[0].Path

				Ω(st).Should(Equal(imports[0].Path))
			})
		})
	})
})
