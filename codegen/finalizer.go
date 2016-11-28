package codegen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"goa.design/goa.v2/design"
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
func RecursiveFinalizer(att *design.AttributeExpr, target string, depth int, vs ...map[string]bool) string {
	var assignments []string
	o, ok := att.Type.(design.Object)
	if ok {
		if ut, ok := att.Type.(design.UserType); ok {
			if len(vs) == 0 {
				vs = []map[string]bool{make(map[string]bool)}
			} else if _, ok := vs[0][ut.Name()]; ok {
				return ""
			}
			vs[0][ut.Name()] = true
			att = ut.Attribute()
		}
		o.WalkAttributes(func(n string, catt *design.AttributeExpr) error {
			if att.HasDefaultValue(n) {
				data := map[string]interface{}{
					"target":     target,
					"field":      n,
					"catt":       catt,
					"depth":      depth,
					"isDatetime": catt.Validation != nil && catt.Validation.Format == design.FormatDateTime,
					"defaultVal": printVal(catt, catt.DefaultValue),
				}
				assignments = append(assignments, RunTemplate(assignmentT, data))
			}
			assignment := RecursiveFinalizer(
				catt,
				fmt.Sprintf("%s.%s", target, Goify(n, true)),
				depth+1,
				vs...,
			)
			if assignment != "" {
				if design.IsObject(catt.Type) {
					assignment = fmt.Sprintf("%sif %s.%s != nil {\n%s\n%s}",
						Tabs(depth), target, Goify(n, true), assignment, Tabs(depth))
				}
				assignments = append(assignments, assignment)
			}
			return nil
		})
	} else if a := design.AsArray(att.Type); a != nil {
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

// printVal prints the value corresponding to the given attribute.
// The value is already checked for the compatibility with the data type.
func printVal(att *design.AttributeExpr, val interface{}) string {
	t := att.Type
	switch {
	case design.IsPrimitive(t):
		// For primitive types, simply print the value
		s := fmt.Sprintf("%#v", val)
		if att.Validation != nil && att.Validation.Format == design.FormatDateTime {
			s = fmt.Sprintf("time.Parse(time.RFC3339, %s)", s)
		}
		return s
	case design.IsMap(t):
		// The input is a hash
		m := design.AsMap(t)
		mval := val.(map[interface{}]interface{})
		if len(mval) == 0 {
			return fmt.Sprintf("%s{}", GoTypeName(t, nil, 0, false))
		}
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s{", GoTypeName(t, nil, 0, false)))
		for k, v := range mval {
			buffer.WriteString(fmt.Sprintf("%s: %s, ", printVal(m.KeyType, k), printVal(m.ElemType, v)))
		}
		buffer.Truncate(buffer.Len() - 2) // remove ", "
		buffer.WriteString("}")
		return buffer.String()
	case design.IsArray(t):
		// Input is an array
		a := design.AsArray(t)
		aval := val.([]interface{})
		if len(aval) == 0 {
			return fmt.Sprintf("%s{}", GoTypeName(t, nil, 0, false))
		}
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s{", GoTypeName(t, nil, 0, false)))
		for _, e := range aval {
			buffer.WriteString(fmt.Sprintf("%s, ", printVal(a.ElemType, e)))
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
*/}}{{ tabs .depth }}var {{ $defaultName }}{{if .isDatetime}}, _{{end}} = {{ .defaultVal }}
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
