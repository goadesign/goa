package codegen

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/goadesign/goa/design"
)

// Finalizer is the code generator for the 'Finalize' type methods.
type Finalizer struct {
	assignmentT      *template.Template
	arrayAssignmentT *template.Template
	seen             map[*design.AttributeDefinition]map[*design.AttributeDefinition]*bytes.Buffer
}

// NewFinalizer instantiates a finalize code generator.
func NewFinalizer() *Finalizer {
	var (
		f   = &Finalizer{seen: make(map[*design.AttributeDefinition]map[*design.AttributeDefinition]*bytes.Buffer)}
		err error
	)
	fm := template.FuncMap{
		"tabs":         Tabs,
		"goify":        Goify,
		"gotyperef":    GoTypeRef,
		"add":          Add,
		"finalizeCode": f.Code,
	}
	f.assignmentT, err = template.New("assignment").Funcs(fm).Parse(assignmentTmpl)
	if err != nil {
		panic(err)
	}
	f.arrayAssignmentT, err = template.New("arrAssignment").Funcs(fm).Parse(arrayAssignmentTmpl)
	if err != nil {
		panic(err)
	}
	return f
}

// Code produces Go code that sets the default values for fields recursively for the given
// attribute.
func (f *Finalizer) Code(att *design.AttributeDefinition, target string, depth int) string {
	buf := f.recurse(att, att, target, depth)
	return buf.String()
}

func (f *Finalizer) recurse(root, att *design.AttributeDefinition, target string, depth int) *bytes.Buffer {
	var (
		buf   = new(bytes.Buffer)
		first = true
	)

	if s, ok := f.seen[root]; ok {
		if buf, ok := s[att]; ok {
			return buf
		}
		s[att] = buf
	} else {
		f.seen[root] = map[*design.AttributeDefinition]*bytes.Buffer{att: buf}
	}

	if o := att.Type.ToObject(); o != nil {
		o.IterateAttributes(func(n string, catt *design.AttributeDefinition) error {
			if att.HasDefaultValue(n) {
				data := map[string]interface{}{
					"target":     target,
					"field":      n,
					"catt":       catt,
					"depth":      depth,
					"isDatetime": catt.Type == design.DateTime,
					"defaultVal": PrintVal(catt.Type, catt.DefaultValue),
				}
				if !first {
					buf.WriteByte('\n')
				} else {
					first = false
				}
				buf.WriteString(RunTemplate(f.assignmentT, data))
			}
			a := f.recurse(root, catt, fmt.Sprintf("%s.%s", target, Goify(n, true)), depth+1).String()
			if a != "" {
				if catt.Type.IsObject() {
					a = fmt.Sprintf("%sif %s.%s != nil {\n%s\n%s}",
						Tabs(depth), target, Goify(n, true), a, Tabs(depth))
				}
				if !first {
					buf.WriteByte('\n')
				} else {
					first = false
				}
				buf.WriteString(a)
			}
			return nil
		})
	} else if a := att.Type.ToArray(); a != nil {
		data := map[string]interface{}{
			"elemType": a.ElemType,
			"target":   target,
			"depth":    1,
		}
		if as := RunTemplate(f.arrayAssignmentT, data); as != "" {
			buf.WriteString(as)
		}
	}
	return buf
}

// PrintVal prints the given value corresponding to the given data type.
// The value is already checked for the compatibility with the data type.
func PrintVal(t design.DataType, val interface{}) string {
	switch {
	case t.IsPrimitive():
		// For primitive types, simply print the value
		s := fmt.Sprintf("%#v", val)
		switch t {
		case design.Number:
			s = fmt.Sprintf("%f", val)
		case design.DateTime:
			s = fmt.Sprintf("time.Parse(time.RFC3339, %s)", s)
		}
		return s
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
			buffer.WriteString(fmt.Sprintf("%s: %s, ", PrintVal(h.KeyType.Type, k), PrintVal(h.ElemType.Type, v)))
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
			buffer.WriteString(fmt.Sprintf("%s, ", PrintVal(a.ElemType.Type, e)))
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

	arrayAssignmentTmpl = `{{ $a := finalizeCode .elemType "e" (add .depth 1) }}{{/*
*/}}{{ if $a }}{{ tabs .depth }}for _, e := range {{ .target }} {
{{ $a }}
{{ tabs .depth }}}{{ end }}`
)
