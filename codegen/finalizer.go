package codegen

import (
	"bytes"
	"fmt"
	"text/template"

	"goa.design/goa.v2/design"
)

// Finalizer is the code generator for the 'Finalize' type methods. These methods
// convert an instance of a private type (one whose fields are all pointers so
// that missing serialized fields can be computed) to a public type (once whose
// fields are not pointers for required attributes and that has been validated).
type Finalizer struct {
	assignmentT      *template.Template
	arrayAssignmentT *template.Template
	seen             map[*design.AttributeExpr]map[*design.AttributeExpr]*bytes.Buffer
}

// NewFinalizer instantiates a finalize code generator.
func NewFinalizer() *Finalizer {
	f := &Finalizer{seen: make(map[*design.AttributeExpr]map[*design.AttributeExpr]*bytes.Buffer)}
	fm := template.FuncMap{"finalizeCode": f.Code}
	f.assignmentT = template.Must(template.New("assignment").Funcs(fm).Parse(assignmentTmpl))
	f.arrayAssignmentT = template.Must(template.New("arrAssignment").Funcs(fm).Parse(arrayAssignmentTmpl))

	return f
}

// Code produces Go code that sets the default values for fields recursively for
// the given attribute. target is the name of the variable being initialized.
func (f *Finalizer) Code(att *design.AttributeExpr, target string) string {
	buf := f.recurse(att, att, target)
	return buf.String()
}

func (f *Finalizer) recurse(root, att *design.AttributeExpr, target string) *bytes.Buffer {
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
		f.seen[root] = map[*design.AttributeExpr]*bytes.Buffer{att: buf}
	}

	if o := design.AsObject(att.Type); o != nil {
		WalkAttributes(o, func(n string, catt *design.AttributeExpr) error {
			if att.HasDefaultValue(n) {
				data := map[string]interface{}{
					"target":      target,
					"field":       Goify(n, true),
					"isPrimitive": design.IsPrimitive(catt.Type),
					"defaultVar":  "default" + Goify(n, true),
					"defaultVal":  PrintVal(catt.Type, catt.DefaultValue),
				}
				if !first {
					buf.WriteByte('\n')
				}
				first = false
				f.assignmentT.Execute(buf, data)
			}
			a := f.recurse(root, catt, fmt.Sprintf("%s.%s", target, Goify(n, true))).String()
			if a != "" {
				if design.IsObject(catt.Type) {
					a = fmt.Sprintf("if %s.%s != nil {\n%s\n}",
						target, Goify(n, true), a)
				}
				if !first {
					buf.WriteByte('\n')
				}
				first = false
				buf.WriteString(a)
			}
			return nil
		})
	} else if a := design.AsArray(att.Type); a != nil {
		data := map[string]interface{}{
			"elemType": a.ElemType,
			"target":   target,
		}
		f.arrayAssignmentT.Execute(buf, data)
	}
	return buf
}

// PrintVal prints the given value corresponding to the given data type.
// The value is already checked for the compatibility with the data type.
func PrintVal(t design.DataType, val interface{}) string {
	switch {
	case design.AsMap(t) != nil:
		m := design.AsMap(t)
		mval := val.(map[interface{}]interface{})
		if len(mval) == 0 {
			return fmt.Sprintf("%s{}", GoType(t, true))
		}
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s{", GoType(t, true)))
		for k, v := range mval {
			buffer.WriteString(fmt.Sprintf("%s: %s, ",
				PrintVal(m.KeyType.Type, k),
				PrintVal(m.ElemType.Type, v)),
			)
		}
		buffer.Truncate(buffer.Len() - 2) // remove ", "
		buffer.WriteString("}")
		return buffer.String()

	case design.AsArray(t) != nil:
		aval := val.([]interface{})
		a := design.AsArray(t)
		if len(aval) == 0 {
			return fmt.Sprintf("%s{}", GoType(t, true))
		}
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("%s{", GoType(t, true)))
		for _, e := range aval {
			buffer.WriteString(fmt.Sprintf("%s, ", PrintVal(a.ElemType.Type, e)))
		}
		buffer.Truncate(buffer.Len() - 2) // remove ", "
		buffer.WriteString("}")
		return buffer.String()

	default:
		// For primitive types, simply print the value
		s := fmt.Sprintf("%#v", val)
		if t.Kind() == design.Float32Kind || t.Kind() == design.Float64Kind {
			s = fmt.Sprintf("%f", val)
		}
		return s
	}
}

const (
	assignmentTmpl = `{{ if .isPrimitive -}}

{{ .defaultVar }} := {{ .defaultVal }}
if {{ .target }}.{{ .field }} == nil {
	{{ .target }}.{{ .field }} = &{{ .defaultVar }}
}

{{- else -}}

if {{ .target }}.{{ .field }} == nil {
	{{ .target }}.{{ .field }} = {{ .defaultVal }}
}

{{- end }}`

	arrayAssignmentTmpl = `{{ $a := finalizeCode .elemType "e" (add .depth 1) }}
{{- if $a -}}
for _, e := range {{ .target }} {
	{{ $a }}
}
{{- end }}`
)
