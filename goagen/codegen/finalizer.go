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
		"tabs":               Tabs,
		"goify":              Goify,
		"gotyperef":          GoTypeRef,
		"add":                Add,
		"recursiveFinalizer": RecursiveFinalizer,
	}
	if assignmentT, err = template.New("assignment").Funcs(fm).Parse(assignmentTmpl); err != nil {
		panic(err)
	}
	if arrayAssignmentT, err = template.New("arrAssignment").Funcs(fm).Parse(arrayAssignmentTmpl); err != nil {
		panic(err)
	}
}

// RecursiveFinalizer produces Go code that sets the default values for fields recursively for the
// given attribute.
func RecursiveFinalizer(att *design.AttributeDefinition, target string, depth int) string {
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
			assignment := RecursiveFinalizer(
				catt,
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				depth+1,
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

// printVal prints the given value corresponding to the given data type.
// The value is already checked for the compatibility with the data type.
func printVal(t design.DataType, val interface{}) string {
	switch {
	case t.IsPrimitive():
		// For primitive types, simply print the value
		return fmt.Sprintf("%#v", val)
	case t.IsHash():
		// The input is a hash
		h := t.ToHash()
		hval := val.(map[interface{}]interface{})
		if len(hval) == 0 {
			return fmt.Sprintf("%s{}", GoTypeName(t, nil, 0, false))
		}
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s{", GoTypeName(t, nil, 0, false)))
		for k, v := range hval {
			buffer.WriteString(fmt.Sprintf("%s: %s, ", printVal(h.KeyType.Type, k), printVal(h.ElemType.Type, v)))
		}
		buffer.Truncate(buffer.Len() - 2) // remove ", "
		buffer.WriteString("}")
		return buffer.String()
	case t.IsArray():
		// Input is an array
		a := t.ToArray()
		aval := val.([]interface{})
		if len(aval) == 0 {
			return fmt.Sprintf("%s{}", GoTypeName(t, nil, 0, false))
		}
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s{", GoTypeName(t, nil, 0, false)))
		for _, e := range aval {
			buffer.WriteString(fmt.Sprintf("%s, ", printVal(a.ElemType.Type, e)))
		}
		buffer.Truncate(buffer.Len() - 2) // remove ", "
		buffer.WriteString("}")
		return buffer.String()
	default:
		// shouldn't happen as the value's compatibility is already checked.
		panic("unknown type")
	}
}

const (
	assignmentTmpl = `{{ if .catt.Type.IsPrimitive }}{{ $defaultName := (print "default" (goify .field true)) }}{{/*
*/}}{{ tabs .depth }}var {{ $defaultName }} = {{ .defaultVal }}
{{ tabs .depth }}if {{ .target }}.{{ goify .field true }} == nil {
{{ tabs .depth }}	{{ .target }}.{{ goify .field true }} = &{{ $defaultName }}
}{{ else }}{{ tabs .depth }}if {{ .target }}.{{ goify .field true }} == nil {
{{ tabs .depth }}	{{ .target }}.{{ goify .field true }} = {{ .defaultVal }}
}{{ end }}`

	arrayAssignmentTmpl = `{{ $assignment := recursiveFinalizer .elemType "e" (add .depth 1) }}{{/*
*/}}{{ if $assignment }}{{ tabs .depth }}for _, e := range {{ .target }} {
{{ $assignment }}
{{ tabs .depth }}}{{ end }}`
)
