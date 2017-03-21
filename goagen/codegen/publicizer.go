package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
)

var (
	simplePublicizeT    *template.Template
	recursivePublicizeT *template.Template
	objectPublicizeT    *template.Template
	arrayPublicizeT     *template.Template
	hashPublicizeT      *template.Template
)

func init() {
	var err error
	fm := template.FuncMap{
		"tabs":                Tabs,
		"goify":               Goify,
		"gotyperef":           GoTypeRef,
		"gotypedef":           GoTypeDef,
		"add":                 Add,
		"publicizer":          Publicizer,
		"recursivePublicizer": RecursivePublicizer,
	}
	if simplePublicizeT, err = template.New("simplePublicize").Funcs(fm).Parse(simplePublicizeTmpl); err != nil {
		panic(err)
	}
	if recursivePublicizeT, err = template.New("recursivePublicize").Funcs(fm).Parse(recursivePublicizeTmpl); err != nil {
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

// RecursivePublicizer produces code that copies fields from the private struct to the
// public struct
func RecursivePublicizer(att *design.AttributeDefinition, source, target string, depth int) string {
	var publications []string
	if o := att.Type.ToObject(); o != nil {
		if ds, ok := att.Type.(design.DataStructure); ok {
			att = ds.Definition()
		}
		o.IterateAttributes(func(n string, catt *design.AttributeDefinition) error {
			publication := Publicizer(
				catt,
				fmt.Sprintf("%s.%s", source, Goify(n, true)),
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				catt.Type.IsPrimitive() && !att.IsPrimitivePointer(n),
				depth+1,
				false,
			)
			publication = fmt.Sprintf("%sif %s.%s != nil {\n%s\n%s}",
				Tabs(depth), source, Goify(n, true), publication, Tabs(depth))
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
		publication = RunTemplate(simplePublicizeT, data)
	case att.Type.IsObject():
		if _, ok := att.Type.(*design.MediaTypeDefinition); ok {
			publication = RunTemplate(recursivePublicizeT, data)
		} else if _, ok := att.Type.(*design.UserTypeDefinition); ok {
			publication = RunTemplate(recursivePublicizeT, data)
		} else {
			publication = RunTemplate(objectPublicizeT, data)
		}
	case att.Type.IsArray():
		// If the array element is primitive type, we can simply copy the elements over (i.e) []string
		if att.Type.HasAttributes() {
			data["elemType"] = att.Type.ToArray().ElemType
			publication = RunTemplate(arrayPublicizeT, data)
		} else {
			publication = RunTemplate(simplePublicizeT, data)
		}
	case att.Type.IsHash():
		if att.Type.HasAttributes() {
			h := att.Type.ToHash()
			data["keyType"] = h.KeyType
			data["elemType"] = h.ElemType
			publication = RunTemplate(hashPublicizeT, data)
		} else {
			publication = RunTemplate(simplePublicizeT, data)
		}
	}
	return publication
}

const (
	simplePublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= {{ if .dereference }}*{{ end }}{{ .sourceField }}`

	recursivePublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= {{ .sourceField }}.Publicize()`

	objectPublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} = &{{ gotypedef .att .depth true false }}{}
{{ recursivePublicizer .att .sourceField .targetField .depth }}`

	arrayPublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= make({{ gotyperef .att.Type .att.AllRequired .depth false }}, len({{ .sourceField }})){{/*
*/}}{{ $i := printf "%s%d" "i" .depth }}{{ $elem := printf "%s%d" "elem" .depth }}
{{ tabs .depth }}for {{ $i }}, {{ $elem }} := range {{ .sourceField }} {
{{ tabs .depth }}{{ publicizer .elemType $elem (printf "%s[%s]" .targetField $i) .dereference (add .depth 1) false }}
{{ tabs .depth }}}`

	hashPublicizeTmpl = `{{ tabs .depth }}{{ .targetField }} {{ if .init }}:{{ end }}= make({{ gotyperef .att.Type .att.AllRequired .depth false }}, len({{ .sourceField }})){{/*
*/}}{{ $k := printf "%s%d" "k" .depth }}{{ $v := printf "%s%d" "v" .depth }}
{{ tabs .depth }}for {{ $k }}, {{ $v }} := range {{ .sourceField }} {
{{ $pubk := printf "%s%s" "pub" $k }}{{ $pubv := printf "%s%s" "pub" $v }}{{/*
*/}}{{ tabs (add .depth 1) }}{{ if .keyType.Type.IsObject }}var {{ $pubk }} {{ gotyperef .keyType.Type .AllRequired .depth false}}
{{ tabs (add .depth 1) }}if {{ $k }} != nil {
{{ tabs (add .depth 1) }}{{ publicizer .keyType $k $pubk .dereference (add .depth 1) false }}
{{ tabs (add .depth 1) }}}{{ else }}{{ publicizer .keyType $k $pubk .dereference (add .depth 1) true }}{{ end }}
{{ tabs (add .depth 1) }}{{if .elemType.Type.IsObject }}var {{ $pubv }} {{ gotyperef .elemType.Type .AllRequired .depth false }}
{{ tabs (add .depth 1) }}if {{ $v }} != nil {
{{ tabs (add .depth 1) }}{{ publicizer .elemType $v $pubv .dereference (add .depth 1) false }}
{{ tabs (add .depth 1) }}}{{ else }}{{ publicizer .elemType $v $pubv .dereference (add .depth 1) true }}{{ end }}
{{ tabs .depth }}	{{ printf "%s[%s]" .targetField $pubk }} = {{ $pubv }}
{{ tabs .depth }}}`
)
