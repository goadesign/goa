package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa.v2/design"
)

var (
	simplePublicizeT    *template.Template
	recursivePublicizeT *template.Template
	objectPublicizeT    *template.Template
	arrayPublicizeT     *template.Template
	mapPublicizeT       *template.Template
)

func init() {
	fm := template.FuncMap{
		"goify":               Goify,
		"gotyperef":           GoTypeRef,
		"gotypedef":           GoTypeDef,
		"publicizer":          Publicizer,
		"recursivePublicizer": RecursivePublicizer,
	}
	simplePublicizeT = template.Must(template.New("simplePublicize").Funcs(fm).Parse(simplePublicizeTmpl))
	recursivePublicizeT = template.Must(template.New("recursivePublicize").Funcs(fm).Parse(recursivePublicizeTmpl))
	objectPublicizeT = template.Must(template.New("objectPublicize").Funcs(fm).Parse(objectPublicizeTmpl))
	arrayPublicizeT = template.Must(template.New("arrPublicize").Funcs(fm).Parse(arrayPublicizeTmpl))
	mapPublicizeT = template.Must(template.New("mapPublicize").Funcs(fm).Parse(mapPublicizeTmpl))
}

// RecursivePublicizer produces code that copies fields from the private struct
// to the public struct
func RecursivePublicizer(att *design.AttributeExpr, source, target string) string {
	var publications []string
	if o := design.AsObject(att.Type); o != nil {
		WalkAttributes(o, func(n string, catt *design.AttributeExpr) error {
			publication := Publicizer(
				catt,
				fmt.Sprintf("%s.%s", source, Goify(n, true)),
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				design.IsPrimitive(catt.Type) && !att.IsPrimitivePointer(n),
				false,
			)
			publication = fmt.Sprintf("if %s.%s != nil {\n%s\n}",
				source, Goify(n, true), publication)
			publications = append(publications, publication)
			return nil
		})
	}
	return strings.Join(publications, "\n")
}

// hasAttributes returns true if t is an object or an array or map of objects.
func hasAttributes(t design.DataType) bool {
	switch {
	case design.IsPrimitive(t):
		return false
	case design.IsArray(t):
		return hasAttributes(design.AsArray(t).ElemType.Type)
	case design.IsMap(t):
		m := design.AsMap(t)
		return hasAttributes(m.ElemType.Type) || hasAttributes(m.KeyType.Type)
	case design.IsObject(t):
		return true
	default:
		panic(fmt.Sprintf("unknown data type %T", t)) // bug
	}
}

// Publicizer publicizes a single attribute based on the type.
func Publicizer(att *design.AttributeExpr, source, target string, dereference bool, init bool) string {
	var publication bytes.Buffer
	data := map[string]interface{}{
		"sourceField": source,
		"targetField": target,
		"att":         att,
		"dereference": dereference,
		"init":        init,
	}
	switch {
	case design.IsPrimitive(att.Type):
		simplePublicizeT.Execute(&publication, data)
	case design.IsObject(att.Type):
		if _, ok := att.Type.(design.UserType); ok {
			recursivePublicizeT.Execute(&publication, data)
		} else {
			objectPublicizeT.Execute(&publication, data)
		}
	case design.IsArray(att.Type):
		// If the array element is primitive type, we can simply copy
		// the elements over (i.e) []string
		if hasAttributes(att.Type) {
			data["elemType"] = design.AsArray(att.Type).ElemType
			arrayPublicizeT.Execute(&publication, data)
		} else {
			simplePublicizeT.Execute(&publication, data)
		}
	case design.IsMap(att.Type):
		if hasAttributes(att.Type) {
			m := design.AsMap(att.Type)
			data["keyType"] = m.KeyType
			data["elemType"] = m.ElemType
			mapPublicizeT.Execute(&publication, data)
		} else {
			simplePublicizeT.Execute(&publication, data)
		}
	}
	return publication.String()
}

const (
	simplePublicizeTmpl = `{{ .target }} {{ if .init }}:{{ end }}= {{ if .dereference }}*{{ end }}{{ .source }}`

	recursivePublicizeTmpl = `{{ .target }} {{ if .init }}:{{ end }}= {{ .source }}.Publicize()`

	objectPublicizeTmpl = `{{ .target }} = &{{ gotypedef .att true false }}{}
{{ recursivePublicizer .att .source .target }}`

	arrayPublicizeTmpl = `{{ .target }} {{ if .init }}:{{ end }}= make({{ gotyperef .att.Type .att.AllRequired false }}, len({{ .source }}))
for {{ .i }}, {{ .elem }} := range {{ .source }} {
	{{ publicizer .elemType .elem (printf "%s[%s]" .target .i) .dereference false }}
}`

	mapPublicizeTmpl = `{{ .target }} {{ if .init }}:{{ end }}= make({{ gotyperef .att.Type .att.AllRequired false }}, len({{ .sourceField }}))
for {{ .key }}, {{ .val }} := range {{ .source }} {
{{- $pubk := printf "%s%s" "pub" .key -}}
{{- $pubv := printf "%s%s" "pub" .val }}
{{- if .keyIsObject }}
	var {{ $pubk }} {{ gotyperef .keyType.Type .AllRequired false}}
	if {{ .key }} != nil {
		{{ publicizer .keyType .key $pubk .dereference false }}
	}
{{- else -}}
	{{ publicizer .keyType .key $pubk .dereference true }}
{{ end -}}
{{- if .elemIsObject }}
	var {{ $pubv }} {{ gotyperef .elemType.Type .AllRequired false }}
	if {{ .val }} != nil {
		{{ publicizer .elemType .val $pubv .dereference false }}
	}
{{- else -}}
	{{ publicizer .elemType .val $pubv .dereference true }}
{{ end -}}
	{{ printf "%s[%s]" .targetField $pubk }} = {{ $pubv }}
}`
)
