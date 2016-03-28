package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
)

var (
	primitivePublicizeT *template.Template
	objectPublicizeT    *template.Template
	arrayPublicizeT     *template.Template
	hashPublicizeT      *template.Template
)

func init() {
	var err error
	fm := template.FuncMap{
		"tabs":       Tabs,
		"goify":      Goify,
		"gotyperef":  GoTypeRef,
		"add":        Add,
		"publicizer": Publicizer,
	}
	if primitivePublicizeT, err = template.New("primitivePublicize").Funcs(fm).Parse(primitivePublicizeTmpl); err != nil {
		panic(err)
	}
	if objectPublicizeT, err = template.New("objectPublicize").Funcs(fm).Parse(objectPublicizeTmpl); err != nil {
		panic(err)
	}
	if arrayPublicizeT, err = template.New("arrPublicize").Funcs(fm).Parse(arrayPublicizeTmpl); err != nil {
		panic(err)
	}
	if hashPublicizeT, err = template.New("hashPublicize").Funcs(fm).Parse(hashPublicizeTmpl); err != nil {
		panic(err)
	}
}

// Recursive publicizer produces code that copies fields from the private struct to the
// public struct
func RecursivePublicizer(att *design.AttributeDefinition, source, target string, depth int) string {
	var publications []string
	if o := att.Type.ToObject(); o != nil {
		if mt, ok := att.Type.(*design.MediaTypeDefinition); ok {
			// Hmm media types should never get here
			att = mt.AttributeDefinition
		} else if ut, ok := att.Type.(*design.UserTypeDefinition); ok {
			att = ut.AttributeDefinition
		}
		o.IterateAttributes(func(n string, catt *design.AttributeDefinition) error {
			publication := Publicizer(
				catt,
				fmt.Sprintf("%s.%s", source, Goify(n, true)),
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				catt.Type.IsPrimitive() && !att.IsPrimitivePointer(n),
				depth,
				false,
			)
			publication = fmt.Sprintf("if %s.%s != nil {\n%s\n}", source, Goify(n, true), publication)
			publications = append(publications, publication)
			return nil
		})
	}
	return strings.Join(publications, "\n")
}

// Publicizer publicizes a single attribute based on the type.
func Publicizer(att *design.AttributeDefinition, sourceField, targetField string, dereference bool, depth int, init bool) string {
	var publication string
	data := map[string]interface{}{
		"sourceField": sourceField,
		"targetField": targetField,
		"depth":       depth,
		"att":         att,
		"dereference": dereference,
		"init":        init,
	}
	switch {
	case att.Type.IsPrimitive():
		publication = RunTemplate(primitivePublicizeT, data)
	case att.Type.IsObject():
		publication = RunTemplate(objectPublicizeT, data)
	case att.Type.IsArray():
		// If the array element is primitive type, we can simply copy the elements over (i.e) []string
		if arr := att.Type.ToArray(); arr.ElemType.Type.IsPrimitive() {
			publication = RunTemplate(primitivePublicizeT, data)
		} else {
			data["elemType"] = arr.ElemType
			publication = RunTemplate(arrayPublicizeT, data)
		}
	case att.Type.IsHash():
		if h := att.Type.ToHash(); h.KeyType.Type.IsPrimitive() && h.ElemType.Type.IsPrimitive() {
			publication = RunTemplate(primitivePublicizeT, data)
		} else {
			data["keyType"] = h.KeyType
			data["elemType"] = h.ElemType
			publication = RunTemplate(hashPublicizeT, data)
		}
	}
	return publication
}

const (
	primitivePublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= {{ if .dereference }}*{{ end }}{{ .sourceField }}`

	objectPublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= {{ .sourceField }}.Publicize()`

	arrayPublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= make({{ gotyperef .att.Type .att.AllRequired .depth false }}, len({{ .sourceField }})){{/*
*/}}{{ $i := printf "%s%d" "i" .depth }}{{ $elem := printf "%s%d" "elem" .depth }}
{{ tabs .depth }}for {{ $i }}, {{ $elem }} := range {{ .sourceField }} {
{{ tabs .depth }}	{{ publicizer .elemType $elem (printf "%s[%s]" .targetField $i) .dereference (add .depth 1) false }}
{{ tabs .depth }}}`

	hashPublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= make({{ gotyperef .att.Type .att.AllRequired .depth false }}, len({{ .sourceField }})){{/*
*/}}{{ $k := printf "%s%d" "k" .depth }}{{ $v := printf "%s%d" "v" .depth }}
{{ tabs .depth }}for {{ $k }}, {{ $v }} := range {{ .sourceField }} { {{ $pubk := printf "%s%s" "pub" $k }}{{ $pubv := printf "%s%s" "pub" $v }}
{{ tabs .depth }}	{{ publicizer .keyType $k $pubk .dereference (add .depth 1) true }}
{{ tabs .depth }}	{{ publicizer .elemType $v $pubv .dereference (add .depth 1) true }}
{{ tabs .depth }}	{{ printf "%s[%s]" .targetField $pubk }} = {{ $pubv }}
{{ tabs .depth }}}`
)
