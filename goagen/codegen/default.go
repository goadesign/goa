package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
)

var (
	assignmentT      *template.Template
	arrayAssignmentT *template.Template
)

// init instantiates the templates
func init() {
	var err error
	fm := template.FuncMap{
		"tabs":              Tabs,
		"goify":             Goify,
		"gotyperef":         GoTypeRef,
		"add":               Add,
		"recursiveAssigner": RecursiveAssigner,
	}
	if assignmentT, err = template.New("assignment").Funcs(fm).Parse(assignmentTmpl); err != nil {
		panic(err)
	}
	if arrayAssignmentT, err = template.New("assignment").Funcs(fm).Parse(arrayAssignmentTmpl); err != nil {
		panic(err)
	}
}

func printVal(t design.DataType, val interface{}) string {
	switch {
	case t.IsPrimitive():
		// For primitive types, simply print the value
		return fmt.Sprintf("%#v", val)
	case t.IsHash():
		// The input is a hash
		h := t.ToHash()
		var hval map[interface{}]interface{}
		switch val.(type) {
		case design.HashVal:
			hval = map[interface{}]interface{}(val.(design.HashVal))
		default:
			hval = val.(map[interface{}]interface{})
		}
		if len(hval) == 0 {
			return fmt.Sprintf("%s{}", GoPackageTypeName(t, nil, 0))
		} else {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("%s{", GoPackageTypeName(t, nil, 0)))
			for k, v := range hval {
				buffer.WriteString(fmt.Sprintf("%s: %s, ", printVal(h.KeyType.Type, k), printVal(h.ElemType.Type, v)))
			}
			buffer.WriteString("}")
			return buffer.String()
		}
	case t.IsArray():
		// Input is an array
		a := t.ToArray()
		var aval []interface{}
		switch val.(type) {
		case design.ArrayVal:
			aval = []interface{}(val.(design.ArrayVal))
		default:
			aval = val.([]interface{})
		}
		if len(aval) == 0 {
			return fmt.Sprintf("%s{}", GoPackageTypeName(t, nil, 0))
		} else {
			var buffer bytes.Buffer
			buffer.WriteString(fmt.Sprintf("%s{", GoPackageTypeName(t, nil, 0)))
			for _, e := range aval {
				buffer.WriteString(fmt.Sprintf("%s, ", printVal(a.ElemType.Type, e)))
			}
			buffer.WriteString("}")
			return buffer.String()
		}
	default:
		// shouldn't happen as the value's compatibility is already checked.
		panic("unknown type")
	}
}

func RecursiveAssigner(att *design.AttributeDefinition, target, context string, depth int) string {
	var assignments []string
	if o := att.Type.ToObject(); o != nil {
		if mt, ok := att.Type.(*design.MediaTypeDefinition); ok {
			att = mt.AttributeDefinition
		} else if ut, ok := att.Type.(*design.UserTypeDefinition); ok {
			att = ut.AttributeDefinition
		}
		o.IterateAttributes(func(n string, catt *design.AttributeDefinition) error {
			if att.HasDefaultValue(n) {
				data := map[string]interface{}{
					"target":     target,
					"field":      n,
					"catt":       catt,
					"depth":      depth,
					"defaultVal": printVal(catt.Type, catt.DefaultValue),
				}
				assignments = append(assignments, RunTemplate(assignmentT, data))
			}
			actualDepth := depth
			if catt.Type.IsObject() {
				actualDepth = depth + 1
			}
			assignment := RecursiveAssigner(
				catt,
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				fmt.Sprintf("%s.%s", context, n),
				actualDepth,
			)
			if assignment != "" {
				if catt.Type.IsObject() {
					assignment = fmt.Sprintf("%sif %s.%s != nil {\n%s\n%s}",
						Tabs(depth), target, Goify(n, true), assignment, Tabs(depth))
				}
				assignments = append(assignments, assignment)
			}
			return nil
		})
	} else if a := att.Type.ToArray(); a != nil {
		data := map[string]interface{}{
			"elemType": a.ElemType,
			"context":  context,
			"target":   target,
			"depth":    1,
		}
		assignment := RunTemplate(arrayAssignmentT, data)
		if assignment != "" {
			assignments = append(assignments, assignment)
		}
	}
	return strings.Join(assignments, "\n")
}

const (
	assignmentTmpl = `{{if .catt.Type.IsPrimitive}}{{$defaultName := (print "default" (goify .field true))}}{{tabs .depth}}var {{$defaultName}} {{gotyperef .catt.Type nil 0}}
{{tabs .depth}}if {{.target}}.{{goify .field true}} == {{$defaultName}} {
{{tabs .depth}}{{.target}}.{{goify .field true}} = {{.defaultVal}}}{{else}}{{tabs .depth}}if {{.target}}.{{goify .field true}} == nil {
{{tabs .depth}}{{.target}}.{{goify .field true}} = {{.defaultVal}}
}{{end}}`

	arrayAssignmentTmpl = `{{$assignment := recursiveAssigner .elemType "e" (printf "%s[*]" .context) (add .depth 1)}}{{/*
*/}}{{if $assignment}}{{tabs .depth}}for _, e := range {{.target}} {
{{$assignment}}
{{tabs .depth}}}{{end}}`
)
