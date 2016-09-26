package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/goadesign/goa/design"
)

// TransformMapKey is the name of the metadata used to specify the key for mapping fields when
// generating the code that transforms one data structure into another.
const TransformMapKey = "transform:key"

var (
	// Templates used by GoTypeTransform
	transformT       *template.Template
	transformArrayT  *template.Template
	transformHashT   *template.Template
	transformObjectT *template.Template
)

// Initialize all templates
func init() {
	var err error
	fn := template.FuncMap{
		"tabs":               Tabs,
		"add":                func(a, b int) int { return a + b },
		"goify":              Goify,
		"gotyperef":          GoTypeRef,
		"gotypename":         GoTypeName,
		"transformAttribute": transformAttribute,
		"transformArray":     transformArray,
		"transformMap":       transformMap,
		"transformObject":    transformObject,
		"typeName":           typeName,
	}
	if transformT, err = template.New("transform").Funcs(fn).Parse(transformTmpl); err != nil {
		panic(err) // bug
	}
	if transformArrayT, err = template.New("transformArray").Funcs(fn).Parse(transformArrayTmpl); err != nil {
		panic(err) // bug
	}
	if transformHashT, err = template.New("transformMap").Funcs(fn).Parse(transformHashTmpl); err != nil {
		panic(err) // bug
	}
	if transformObjectT, err = template.New("transformObject").Funcs(fn).Parse(transformObjectTmpl); err != nil {
		panic(err) // bug
	}
}

// GoTypeTransform produces Go code that initializes the data structure defined by target from an
// instance of the data structure described by source. The algorithm matches object fields by name
// or using the value of the "transform:key" attribute metadata when present.
// The function returns an error if target is not compatible with source (different type, fields of
// different type etc). It ignores fields in target that don't have a match in source.
func GoTypeTransform(source, target *design.UserTypeDefinition, targetPkg, funcName string) (string, error) {
	var impl string
	var err error
	switch {
	case source.IsObject():
		if !target.IsObject() {
			return "", fmt.Errorf("source is an object but target type is %s", target.Type.Name())
		}
		impl, err = transformObject(source.(design.Object), target.(design.Object), targetPkg, target.TypeName, "source", "target", 1)
	case source.IsArray():
		if !target.IsArray() {
			return "", fmt.Errorf("source is an array but target type is %s", target.Type.Name())
		}
		impl, err = transformArray(source.(*design.Array), target.(*design.Array), targetPkg, "source", "target", 1)
	case source.IsHash():
		if !target.IsHash() {
			return "", fmt.Errorf("source is a hash but target type is %s", target.Type.Name())
		}
		impl, err = transformMap(source.(*design.Map), target.(*design.Map), targetPkg, "source", "target", 1)
	default:
		panic("cannot transform primitive types") // bug
	}

	if err != nil {
		return "", err
	}
	t := GoTypeRef(target, nil, 0, false)
	if strings.HasPrefix(t, "*") && len(targetPkg) > 0 {
		t = fmt.Sprintf("*%s.%s", targetPkg, t[1:])
	}
	data := map[string]interface{}{
		"Name":      funcName,
		"Source":    source,
		"Target":    target,
		"TargetRef": t,
		"TargetPkg": targetPkg,
		"Impl":      impl,
	}
	return RunTemplate(transformT, data), nil
}

// GoTypeTransformName generates a valid Go identifer that is adequate for naming the type
// transform function that creates an instance of the data structure described by target from an
// instance of the data strucuture described by source.
func GoTypeTransformName(source, target *design.UserTypeDefinition, suffix string) string {
	return fmt.Sprintf("%sTo%s%s", Goify(source.TypeName, true), Goify(target.TypeName, true), Goify(suffix, true))
}

func transformAttribute(source, target *design.AttributeDefinition, targetPkg, sctx, tctx string, depth int) (string, error) {
	if source.Type.Kind() != target.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is of type %s but %s is of type %s",
			sctx, source.Type.Name(), tctx, target.Type.Name())
	}
	switch {
	case source.Type.IsArray():
		return transformArray(source.Type.(*design.Array), target.Type.(*design.Array), targetPkg, sctx, tctx, depth)
	case source.Type.IsHash():
		return transformMap(source.Type.(*design.Map), target.Type.(*design.Map), targetPkg, sctx, tctx, depth)
	case source.Type.IsObject():
		return transformObject(source.Type.(design.Object), target.Type.(design.Object), targetPkg, typeName(target), sctx, tctx, depth)
	default:
		return fmt.Sprintf("%s%s = %s\n", Tabs(depth), tctx, sctx), nil
	}
}

func transformObject(source, target design.Object, targetPkg, targetType, sctx, tctx string, depth int) (string, error) {
	attributeMap, err := computeMapping(source, target, sctx, tctx)
	if err != nil {
		return "", err
	}

	// First validate that all attributes are compatible - doing that in a template doesn't make
	// sense.
	for s, t := range attributeMap {
		sourceAtt := source[s]
		targetAtt := target[t]
		if sourceAtt.Type.Kind() != targetAtt.Type.Kind() {
			return "", fmt.Errorf("incompatible attribute types: %s.%s is of type %s but %s.%s is of type %s",
				sctx, source.Name(), sourceAtt.Type.Name(), tctx, target.Name(), targetAtt.Type.Name())
		}
	}

	// We're good - generate
	data := map[string]interface{}{
		"AttributeMap": attributeMap,
		"Source":       source,
		"Target":       target,
		"TargetPkg":    targetPkg,
		"TargetType":   targetType,
		"SourceCtx":    sctx,
		"TargetCtx":    tctx,
		"Depth":        depth,
	}
	return RunTemplate(transformObjectT, data), nil
}

func transformArray(source, target *design.Array, targetPkg, sctx, tctx string, depth int) (string, error) {
	if source.ElemType.Type.Kind() != target.ElemType.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is an array with elements of type %s but %s is an array with elements of type %s",
			sctx, source.ElemType.Type.Name(), tctx, target.ElemType.Type.Name())
	}
	data := map[string]interface{}{
		"Source":    source,
		"Target":    target,
		"TargetPkg": targetPkg,
		"SourceCtx": sctx,
		"TargetCtx": tctx,
		"Depth":     depth,
	}
	return RunTemplate(transformArrayT, data), nil
}

func transformMap(source, target *design.Hash, targetPkg, sctx, tctx string, depth int) (string, error) {
	if source.ElemType.Type.Kind() != target.ElemType.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is a hash with elements of type %s but %s is a hash with elements of type %s",
			sctx, source.ElemType.Type.Name(), tctx, target.ElemType.Type.Name())
	}
	if source.KeyType.Type.Kind() != target.KeyType.Type.Kind() {
		return "", fmt.Errorf("incompatible attribute types: %s is a hash with keys of type %s but %s is a hash with keys of type %s",
			sctx, source.KeyType.Type.Name(), tctx, target.KeyType.Type.Name())
	}
	data := map[string]interface{}{
		"Source":    source,
		"Target":    target,
		"TargetPkg": targetPkg,
		"SourceCtx": sctx,
		"TargetCtx": tctx,
		"Depth":     depth,
	}
	return RunTemplate(transformHashT, data), nil
}

// computeMapping returns a map that indexes the target type definition object attributes with the
// corresponding source type definition object attributes. An attribute is associated with another
// attribute if their map key match. The map key of an attribute is the value of the TransformMapKey
// metadata if present, the attribute name otherwise.
// The function returns an error if the TransformMapKey metadata is malformed (has no value).
func computeMapping(source, target design.Object, sctx, tctx string) (map[string]string, error) {
	attributeMap := make(map[string]string)
	sourceMap := make(map[string]string)
	targetMap := make(map[string]string)
	for name, att := range source {
		key := name
		if keys, ok := att.Metadata[TransformMapKey]; ok {
			if len(keys) == 0 {
				return nil, fmt.Errorf("invalid metadata transform key: missing value on attribte %s of %s", name, sctx)
			}
			key = keys[0]
		}
		sourceMap[key] = name
	}
	for name, att := range target {
		key := name
		if keys, ok := att.Metadata[TransformMapKey]; ok {
			if len(keys) == 0 {
				return nil, fmt.Errorf("invalid metadata transform key: missing value on attribute %s of %s", name, tctx)
			}
			key = keys[0]
		}
		targetMap[key] = name
	}
	for key, attName := range sourceMap {
		if targetAtt, ok := targetMap[key]; ok {
			attributeMap[attName] = targetAtt
		}
	}
	return attributeMap, nil
}

const transformTmpl = `func {{ .Name }}(source {{ gotyperef .Source nil 0 false }}) (target {{ .TargetRef }}) {
{{ .Impl }}	return
}
`

const transformObjectTmpl = `{{ tabs .Depth }}{{ .TargetCtx }} = new({{ if .TargetPkg }}{{ .TargetPkg }}.{{ end }}{{ if .TargetType }}{{ .TargetType }}{{ else }}{{ gotyperef .Target.Type .Target.AllRequired 1 false }}{{ end }})
{{ range $source, $target := .AttributeMap }}{{/*
*/}}{{ $sourceAtt := index $.Source $source }}{{ $targetAtt := index $.Target $target }}{{/*
*/}}{{ $source := goify $source true }}{{ $target := goify $target true }}{{/*
*/}}{{     if $sourceAtt.Type.IsArray }}{{ transformArray  ToArray($sourceAtt.Type) ToArray($targetAtt.Type)  $.TargetPkg (printf "%s.%s" $.SourceCtx $source) (printf "%s.%s" $.TargetCtx $target) $.Depth }}{{/*
*/}}{{ else if $sourceAtt.Type.IsHash }}{{ transformMap    ToMap($sourceAtt.Type)   ToMap($targetAtt.Type)    $.TargetPkg (printf "%s.%s" $.SourceCtx $source) (printf "%s.%s" $.TargetCtx $target) $.Depth }}{{/*
*/}}{{ else if $sourceAtt.Type.IsObject }}{{ transformObject ToObject($sourceAtt.Type) ToObject($targetAtt.Type) $.TargetPkg (typeName $targetAtt) (printf "%s.%s" $.SourceCtx $source) (printf "%s.%s" $.TargetCtx $target) $.Depth }}{{/*
*/}}{{ else }}{{ tabs $.Depth }}{{ $.TargetCtx }}.{{ $target }} = {{ $.SourceCtx }}.{{ $source }}
{{ end }}{{ end }}`

const transformArrayTmpl = `{{ tabs .Depth }}{{ .TargetCtx}} = make([]{{ gotyperef .Target.ElemType.Type nil 0 false }}, len({{ .SourceCtx }}))
{{ tabs .Depth }}for i, v := range {{ .SourceCtx }} {
{{ transformAttribute .Source.ElemType .Target.ElemType .TargetPkg (printf "%s[i]" .SourceCtx) (printf "%s[i]" .TargetCtx) (add .Depth 1) }}{{/*
*/}}{{ tabs .Depth }}}
`

const transformHashTmpl = `{{ tabs .Depth }}{{ .TargetCtx }} = make(map[{{ gotyperef .Target.KeyType.Type nil 0 false }}]{{ gotyperef .Target.ElemType.Type nil 0 false }}, len({{ .SourceCtx }}))
{{ tabs .Depth }}for k, v := range {{ .SourceCtx }} {
{{ tabs .Depth }}	var tk {{ gotyperef .Target.KeyType.Type nil 0 false }}
{{ transformAttribute .Source.KeyType .Target.KeyType .TargetPkg "k" "tk" (add .Depth 1) }}{{/*
*/}}{{ tabs .Depth }}	var tv {{ gotyperef .Target.ElemType.Type nil 0 false }}
{{ transformAttribute .Source.ElemType .Target.ElemType .TargetPkg "v" "tv" (add .Depth 1) }}{{/*
*/}}{{ tabs .Depth }}	{{ .TargetCtx }}[tk] = tv
{{ tabs .Depth }}}
`
