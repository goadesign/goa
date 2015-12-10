package codegen_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/raphael/goa/design"
	. "github.com/raphael/goa/design/dsl"
	"github.com/raphael/goa/goagen/codegen"
)

var _ = Describe("code generation", func() {
	BeforeEach(func() {
		codegen.TempCount = 0
	})

	Describe("GoTypeDef", func() {
		Context("given an attribute definition with fields", func() {
			var att *AttributeDefinition
			var object Object
			var required *RequiredValidationDefinition
			var st string

			JustBeforeEach(func() {
				att = new(AttributeDefinition)
				att.Type = object
				if required != nil {
					att.Validations = []ValidationDefinition{required}
				}
				st = codegen.GoTypeDef(att, 0, true, false)
			})

			Context("of primitive types", func() {
				BeforeEach(func() {
					object = Object{
						"foo": &AttributeDefinition{Type: Integer},
						"bar": &AttributeDefinition{Type: String},
					}
					required = nil
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Bar string `json:\"bar,omitempty\"`\n" +
						"	Foo int `json:\"foo,omitempty\"`\n" +
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
					Ω(st).Should(Equal("struct {\n\tFoo map[int]int `json:\"foo,omitempty\"`\n}"))
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
					Ω(st).Should(Equal("struct {\n\tFoo []int `json:\"foo,omitempty\"`\n}"))
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
						"		KeyAtt string `json:\"keyAtt,omitempty\"`\n" +
						"	}]*struct {\n" +
						"		ElemAtt int `json:\"elemAtt,omitempty\"`\n" +
						"	} `json:\"foo,omitempty\"`\n" +
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
						"		Bar int `json:\"bar,omitempty\"`\n" +
						"	} `json:\"foo,omitempty\"`\n" +
						"}"
					Ω(st).Should(Equal(expected))
				})
			})

			Context("that are required", func() {
				BeforeEach(func() {
					object = Object{
						"foo": &AttributeDefinition{Type: Integer},
					}
					required = &RequiredValidationDefinition{
						Names: []string{"foo"},
					}
				})

				It("produces the struct go code", func() {
					expected := "struct {\n" +
						"	Foo int `json:\"foo\"`\n" +
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
					Ω(source).Should(Equal("[]*struct {\n\tBar string `json:\"bar,omitempty\"`\n\tFoo int `json:\"foo,omitempty\"`\n}"))
				})
			})
		})

	})

	Describe("Marshaler", func() {
		var marshaler, unmarshaler string
		var context, source, target string

		BeforeEach(func() {
			codegen.TempCount = 0
			context = ""
			source = "raw"
			target = "p"
		})

		Context("with a primitive type", func() {
			var p Primitive

			JustBeforeEach(func() {
				marshaler = codegen.TypeMarshaler(p, context, source, target)
				codegen.TempCount = 0
				unmarshaler = codegen.TypeUnmarshaler(p, context, source, target)
			})

			Context("integer", func() {
				BeforeEach(func() {
					p = Primitive(IntegerKind)
				})

				It("generates the marshaler code", func() {
					expected := `	p = raw`
					Ω(marshaler).Should(Equal(expected))
				})

				It("generates the unmarshaler code", func() {
					expected := `	if f, ok := raw.(float64); ok {
		p = int(f)
	} else {
		err = goa.InvalidAttributeTypeError(` + "``" + `, raw, "int", err)
	}`
					Ω(unmarshaler).Should(Equal(expected))
				})
			})

			Context("string", func() {
				BeforeEach(func() {
					p = Primitive(StringKind)
				})

				It("generates the marshaler code", func() {
					expected := `	p = raw`
					Ω(marshaler).Should(Equal(expected))
				})

				It("generates the unmarshaler code", func() {
					expected := `	if val, ok := raw.(string); ok {
		p = val
	} else {
		err = goa.InvalidAttributeTypeError(` + "``" + `, raw, "string", err)
	}`
					Ω(unmarshaler).Should(Equal(expected))
				})
			})
		})

		Context("with an array of primitive types", func() {
			var p *Array

			JustBeforeEach(func() {
				marshaler = codegen.TypeMarshaler(p, context, source, target)
				codegen.TempCount = 0
				unmarshaler = codegen.TypeUnmarshaler(p, context, source, target)
			})

			BeforeEach(func() {
				p = &Array{
					ElemType: &AttributeDefinition{
						Type: Primitive(IntegerKind),
					},
				}
			})

			It("generates the marshaler code", func() {
				Ω(marshaler).Should(Equal(arrayMarshaled))
			})

			It("generates the unmarshaler code", func() {
				Ω(unmarshaler).Should(Equal(arrayUnmarshaled))
			})
		})

		Context("with a simple object", func() {
			var o Object

			JustBeforeEach(func() {
				marshaler = codegen.TypeMarshaler(o, context, source, target)
				codegen.TempCount = 0
				unmarshaler = codegen.TypeUnmarshaler(o, context, source, target)
			})

			BeforeEach(func() {
				intAtt := &AttributeDefinition{Type: Primitive(IntegerKind)}
				o = Object{"foo": intAtt}
			})

			It("generates the marshaler code", func() {
				Ω(marshaler).Should(Equal(simpleMarshaled))
			})

			It("generates the unmarshaler code", func() {
				Ω(unmarshaler).Should(Equal(simpleUnmarshaled))
			})
		})

		Context("with a complex object", func() {
			var o Object

			JustBeforeEach(func() {
				marshaler = codegen.TypeMarshaler(o, context, source, target)
				codegen.TempCount = 0
				unmarshaler = codegen.TypeUnmarshaler(o, context, source, target)
			})

			BeforeEach(func() {
				ar := &Array{
					ElemType: &AttributeDefinition{
						Type: &Array{
							ElemType: &AttributeDefinition{
								Type: Primitive(IntegerKind),
							}}}}

				intAtt := &AttributeDefinition{Type: Primitive(IntegerKind)}
				arAtt := &AttributeDefinition{Type: ar}
				io := Object{"foo": intAtt, "bar": arAtt}
				ioAtt := &AttributeDefinition{Type: io}
				o = Object{"baz": ioAtt, "faz": intAtt}
			})

			It("generates the marshaler code", func() {
				Ω(marshaler).Should(Equal(complexMarshaled))
			})

			It("generates the unmarshaler code", func() {
				Ω(unmarshaler).Should(Equal(complexUnmarshaled))
			})

			Context("compiling", func() {
				var gopath, srcDir string
				var tmpl *template.Template
				var tmpFile *os.File
				var out []byte

				JustBeforeEach(func() {
					cmd := exec.Command("go", "build", "-o", "codegen")
					cmd.Env = os.Environ()
					cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s:%s", gopath, os.Getenv("GOPATH")))
					cmd.Dir = srcDir
					var err error
					out, err = cmd.CombinedOutput()
					Ω(out).Should(BeEmpty())
					Ω(err).ShouldNot(HaveOccurred())
				})

				BeforeEach(func() {
					var err error
					gopath, err = ioutil.TempDir("", "")
					Ω(err).ShouldNot(HaveOccurred())
					tmpl, err = template.New("main").Parse(mainTmpl)
					Ω(err).ShouldNot(HaveOccurred())
					srcDir = filepath.Join(gopath, "src", "test")
					err = os.MkdirAll(srcDir, 0755)
					Ω(err).ShouldNot(HaveOccurred())
					tmpFile, err = os.Create(filepath.Join(srcDir, "main.go"))
					Ω(err).ShouldNot(HaveOccurred())
				})

				Context("unmarshaler code", func() {
					BeforeEach(func() {
						unmarshaler = codegen.TypeUnmarshaler(o, context, source, target)
						data := map[string]interface{}{
							"raw": `interface{}(map[string]interface{}{
								"baz": map[string]interface{}{
									"foo": 345.0,
									"bar":[]interface{}{[]interface{}{1.0,2.0,3.0}},
								},
								"faz": 2.0,
							})`,
							"source":     unmarshaler,
							"target":     target,
							"targetType": codegen.GoTypeRef(o, 1),
						}
						err := tmpl.Execute(tmpFile, data)
						Ω(err).ShouldNot(HaveOccurred())
					})
					It("compiles", func() {
						Ω(string(out)).Should(BeEmpty())

						cmd := exec.Command("./codegen")
						cmd.Env = []string{fmt.Sprintf("PATH=%s", filepath.Join(gopath, "bin"))}
						cmd.Dir = srcDir
						code, err := cmd.CombinedOutput()
						Ω(string(code)).Should(Equal(`{"Baz":{"Bar":[[1,2,3]],"Foo":345},"Faz":2}`))
						Ω(err).ShouldNot(HaveOccurred())
					})
				})

				AfterEach(func() {
					os.RemoveAll(gopath)
				})

			})
		})

		Context("with a media type with links", func() {
			var testMediaType *MediaTypeDefinition
			var marshalerImpl string

			BeforeEach(func() {
				Design = nil
				Errors = nil
				fooMediaType := MediaType("application/fooMT", func() {
					Attribute("fooAtt", Integer)
					Attribute("href")
					View("link", func() {
						Attribute("href")
					})
				})
				Ω(Errors).ShouldNot(HaveOccurred())
				barMediaType := MediaType("application/barMT", func() {
					Attribute("barAtt", Integer)
					Attribute("href")
					View("link", func() {
						Attribute("href")
					})
				})
				Ω(Errors).ShouldNot(HaveOccurred())
				bazMediaType := MediaType("application/bazMT", func() {
					Attribute("bazAtt", Integer)
					Attribute("href")
					Attribute("name")
					View("bazLink", func() {
						Attribute("href")
						Attribute("name")
					})
				})
				Ω(Errors).ShouldNot(HaveOccurred())
				testMediaType = MediaType("application/test", func() {
					Attribute("foo", fooMediaType)
					Attribute("bar", barMediaType)
					Attribute("baz", bazMediaType)
					Links(func() {
						Link("foo")
						Link("baz", "bazLink")
					})
				})
				Ω(Errors).ShouldNot(HaveOccurred())
				RunDSL()
				Ω(Errors).ShouldNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				marshaler = codegen.MediaTypeMarshaler(testMediaType, context, source, target, "")
				marshalerImpl = codegen.MediaTypeMarshalerImpl(testMediaType, "default")
			})

			It("generates the marshaler code", func() {
				Ω(marshaler).Should(Equal(mtMarshaled))
				Ω(marshalerImpl).Should(Equal(mtMarshaledImpl))
			})
		})

		Context("with a collection media type", func() {
			var collectionMediaType *MediaTypeDefinition
			var marshalerImpl string

			BeforeEach(func() {
				Design = nil
				Errors = nil

				testMediaType := MediaType("application/testMT", func() {
					Attributes(func() {
						Attribute("id")
					})
				})
				Ω(Errors).ShouldNot(HaveOccurred())

				collectionMediaType = CollectionOf(testMediaType)
				Ω(Errors).ShouldNot(HaveOccurred())

				RunDSL()
				Ω(Errors).ShouldNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				marshaler = codegen.MediaTypeMarshaler(collectionMediaType, context, source, target, "")
				marshalerImpl = codegen.MediaTypeMarshalerImpl(collectionMediaType, "default")
			})

			It("generates the marshaler code", func() {
				Ω(marshaler).Should(Equal(collectionMtMarshaled))
				Ω(marshalerImpl).Should(Equal(collectionMtMarshaledImpl))
			})
		})

		Context("with two media types referring to each other", func() {
			var testMediaType *MediaTypeDefinition
			var testMediaType2 *MediaTypeDefinition
			var marshaler2 string

			BeforeEach(func() {
				Design = nil
				Errors = nil
				testMediaType = MediaType("application/test", func() {
					Attribute("id")
					Attribute("test2", CollectionOf("application/test2"))
					View("default", func() {
						Attribute("id")
					})
				})
				Ω(Errors).ShouldNot(HaveOccurred())
				testMediaType2 = MediaType("application/test2", func() {
					Attribute("id")
					Attribute("test", testMediaType)
					View("default", func() {
						Attribute("test")
					})
				})
				Ω(Errors).ShouldNot(HaveOccurred())
				RunDSL()
				Ω(Errors).ShouldNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				marshaler = codegen.MediaTypeMarshaler(testMediaType, context, source, target, "")
				marshaler2 = codegen.MediaTypeMarshaler(testMediaType2, context, source, target, "")
			})

			It("generates the marshaler code", func() {
				Ω(marshaler).Should(Equal(mtMarshaled))
				Ω(marshaler2).Should(Equal(mtMarshaled2))
			})
		})

	})
})

const (
	arrayMarshaled = `	tmp1 := make([]int, len(raw))
	for tmp2, tmp3 := range raw {
		tmp1[tmp2] = tmp3
	}
	p = tmp1`

	arrayUnmarshaled = `	if val, ok := raw.([]interface{}); ok {
		p = make([]int, len(val))
		for tmp1, v := range val {
			if f, ok := v.(float64); ok {
				p[tmp1] = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(` + "`" + `[*]` + "`" + `, v, "int", err)
			}
		}
	} else {
		err = goa.InvalidAttributeTypeError(` + "``" + `, raw, "array", err)
	}`

	simpleMarshaled = `	tmp1 := map[string]interface{}{
		"foo": raw.Foo,
	}
	p = tmp1`

	simpleUnmarshaled = `	if val, ok := raw.(map[string]interface{}); ok {
		p = new(struct {
			Foo int
		})
		if v, ok := val["foo"]; ok {
			var tmp1 int
			if f, ok := v.(float64); ok {
				tmp1 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(` + "`" + `.Foo` + "`" + `, v, "int", err)
			}
			p.Foo = tmp1
		}
	} else {
		err = goa.InvalidAttributeTypeError(` + "``" + `, raw, "dictionary", err)
	}`

	complexMarshaled = `	tmp1 := map[string]interface{}{
		"faz": raw.Faz,
	}
	if raw.Baz != nil {
		tmp2 := map[string]interface{}{
			"foo": raw.Baz.Foo,
		}
		if raw.Baz.Bar != nil {
			tmp3 := make([][]int, len(raw.Baz.Bar))
			for tmp4, tmp5 := range raw.Baz.Bar {
				tmp6 := make([]int, len(tmp5))
				for tmp7, tmp8 := range tmp5 {
					tmp6[tmp7] = tmp8
				}
				tmp3[tmp4] = tmp6
			}
			tmp2["bar"] = tmp3
		}
		tmp1["baz"] = tmp2
	}
	p = tmp1`

	complexUnmarshaled = `	if val, ok := raw.(map[string]interface{}); ok {
		p = new(struct {
			Baz *struct {
				Bar [][]int
				Foo int
			}
			Faz int
		})
		if v, ok := val["baz"]; ok {
			var tmp1 *struct {
				Bar [][]int
				Foo int
			}
			if val, ok := v.(map[string]interface{}); ok {
				tmp1 = new(struct {
					Bar [][]int
					Foo int
				})
				if v, ok := val["bar"]; ok {
					var tmp2 [][]int
					if val, ok := v.([]interface{}); ok {
						tmp2 = make([][]int, len(val))
						for tmp3, v := range val {
							if val, ok := v.([]interface{}); ok {
								tmp2[tmp3] = make([]int, len(val))
								for tmp4, v := range val {
									if f, ok := v.(float64); ok {
										tmp2[tmp3][tmp4] = int(f)
									} else {
										err = goa.InvalidAttributeTypeError(` + "`" + `.Baz.Bar[*][*]` + "`" + `, v, "int", err)
									}
								}
							} else {
								err = goa.InvalidAttributeTypeError(` + "`" + `.Baz.Bar[*]` + "`" + `, v, "array", err)
							}
						}
					} else {
						err = goa.InvalidAttributeTypeError(` + "`" + `.Baz.Bar` + "`" + `, v, "array", err)
					}
					tmp1.Bar = tmp2
				}
				if v, ok := val["foo"]; ok {
					var tmp5 int
					if f, ok := v.(float64); ok {
						tmp5 = int(f)
					} else {
						err = goa.InvalidAttributeTypeError(` + "`" + `.Baz.Foo` + "`" + `, v, "int", err)
					}
					tmp1.Foo = tmp5
				}
			} else {
				err = goa.InvalidAttributeTypeError(` + "`" + `.Baz` + "`" + `, v, "dictionary", err)
			}
			p.Baz = tmp1
		}
		if v, ok := val["faz"]; ok {
			var tmp6 int
			if f, ok := v.(float64); ok {
				tmp6 = int(f)
			} else {
				err = goa.InvalidAttributeTypeError(` + "`" + `.Faz` + "`" + `, v, "int", err)
			}
			p.Faz = tmp6
		}
	} else {
		err = goa.InvalidAttributeTypeError(` + "``" + `, raw, "dictionary", err)
	}`

	mtMarshaled           = `	p, err = MarshalTest(raw, err)`
	collectionMtMarshaled = `	p, err = MarshalTestmtCollection(raw, err)`
	mtMarshaled2          = `	p, err = MarshalTest2(raw, err)`

	mtMarshaledImpl = `// MarshalTest validates and renders an instance of Test into a interface{}
// using view "default".
func MarshalTest(source *Test, inErr error) (target map[string]interface{}, err error) {
	err = inErr
	tmp1 := map[string]interface{}{
	}
	if source.Bar != nil {
		tmp1["bar"], err = MarshalBarmt(source.Bar, err)
	}
	if source.Baz != nil {
		tmp1["baz"], err = MarshalBazmt(source.Baz, err)
	}
	if source.Foo != nil {
		tmp1["foo"], err = MarshalFoomt(source.Foo, err)
	}
	target = tmp1
	return
}`

	collectionMtMarshaledImpl = `// MarshalTestmtCollection validates and renders an instance of TestmtCollection into a interface{}
// using view "default".
func MarshalTestmtCollection(source TestmtCollection, inErr error) (target []map[string]interface{}, err error) {
	err = inErr
	target = make([]map[string]interface{}, len(source))
	for i, res := range source {
			target[i], err = MarshalTestmt(res, err)
	}
	return
}`

	mainTmpl = `package main

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/raphael/goa"
)

func main() {
	var err error
	raw := {{.raw}}
	var {{.target}} {{.targetType}}
{{.source}}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	b, err := json.Marshal({{.target}})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
	fmt.Print(string(b))
}
`
)
