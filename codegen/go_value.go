// This file renders design values as Go expressions while type names and
// pointer choices are still available to the generator. Generated programs
// receive only the declarations and expression needed for the selected value.
package codegen

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"goa.design/goa/v3/expr"
)

type (
	// GoValueCode contains the Go source needed to build one design value.
	// Declarations must be written before Expression in the same block.
	GoValueCode struct {
		// Declarations initializes values whose addresses appear in Expression.
		Declarations []string
		// Expression is the final typed Go value.
		Expression string
	}

	// UnionConstructorResolver returns the planned constructor for one OneOf
	// branch. The returned name includes any package qualifier needed by the
	// generated file.
	UnionConstructorResolver func(attribute *expr.AttributeExpr, branch string) (string, error)

	// GoTypeLayoutResolver returns the linked Go type selected by one generated
	// package for an attribute and its pointer policy.
	GoTypeLayoutResolver interface {
		GoTypeLayout(attribute *expr.AttributeExpr, policy GoLayoutPolicy) (LinkedGoType, error)
	}

	// goValueRenderer holds the target layout and names local declarations made
	// while rendering one value.
	goValueRenderer struct {
		resolveUnion UnionConstructorResolver
		prefix       string
		nextLocal    int
		declarations []string
	}

	// goMapValue keeps one authored map entry with its stable generation order.
	goMapValue struct {
		order string
		key   reflect.Value
		value reflect.Value
	}
)

// RenderGoValue renders value using the linked Go layout selected for attribute.
// pointer reports whether the complete value is stored through a pointer.
// resolveUnion is required only when attribute contains a OneOf value.
// It checks the supplied value's literal shape, not attribute.DefaultValue or
// schema validation rules. The Go compiler checks the authored custom target.
func RenderGoValue(attribute *expr.AttributeExpr, value any, layout LinkedGoType, pointer bool, resolveUnion UnionConstructorResolver, localPrefix string) (GoValueCode, error) {
	if attribute == nil {
		return GoValueCode{}, fmt.Errorf("render Go value: attribute must not be nil")
	}
	if layout.plan == nil {
		return GoValueCode{}, fmt.Errorf("render Go value for %s: linked Go type must not be empty", attribute.Type.Name())
	}
	prefix := Goify(localPrefix, false)
	if prefix == "" {
		prefix = "defaultValue"
	}
	renderer := &goValueRenderer{
		resolveUnion: resolveUnion,
		prefix:       prefix,
	}
	expression, err := renderer.render(attribute, reflect.ValueOf(value), layout, pointer)
	if err != nil {
		return GoValueCode{}, fmt.Errorf("render Go value for %s: %w", attribute.Type.Name(), err)
	}
	return GoValueCode{
		Declarations: renderer.declarations,
		Expression:   expression,
	}, nil
}

// planGoTypeWithAttributor records the final names selected by an attribute
// scope. Pointer choices still come only from GoTypePlan.
func planGoTypeWithAttributor(attribute *expr.AttributeExpr, policy GoLayoutPolicy, attributor Attributor) (LinkedGoType, error) {
	const owner = "goa.local/generated"
	plan, err := PlanGoType(attribute, GoTypePlanOptions{
		Owner:            owner,
		Policy:           policy,
		RetainNamedValue: true,
		Bind: func(request GoTypeBindingRequest) (GoTypeBinding, error) {
			return GoTypeBinding{
				Owner: request.InheritedOwner,
				name: attributor.Name(
					request.Attribute,
					attributor.Package(request.Attribute),
					policy.Pointer,
					policy.UseDefault,
				),
			}, nil
		},
	})
	if err != nil {
		return LinkedGoType{}, err
	}
	imports := make(map[string]string)
	for _, preference := range plan.ImportPreferences() {
		imports[preference.Path] = preference.Name
	}
	return plan.Link(owner, func(importPath string) string {
		return imports[importPath]
	}), nil
}

// render writes the expression for one value and records any local values
// whose addresses are required by the target layout.
func (r *goValueRenderer) render(attribute *expr.AttributeExpr, value reflect.Value, layout LinkedGoType, pointer bool) (string, error) {
	concrete, err := concreteGoValue(value)
	if err != nil {
		return "", err
	}
	var expression string
	shape := layout
	if layout.plan.kind == GoNamed {
		if layout.plan.value == nil {
			return "", fmt.Errorf("named type %q has no planned value layout", attribute.Type.Name())
		}
		shape = layout.Enter(layout.plan.value)
	}
	switch {
	case expr.IsPrimitive(attribute.Type):
		expression, err = r.renderPrimitive(attribute, concrete, layout)
	case expr.IsArray(attribute.Type):
		expression, err = r.renderArray(attribute, concrete, layout, shape)
	case expr.IsMap(attribute.Type):
		expression, err = r.renderMap(attribute, concrete, layout, shape)
	case expr.IsObject(attribute.Type):
		expression, err = r.renderObject(attribute, concrete, layout, shape, pointer)
		pointer = false
	case expr.IsUnion(attribute.Type):
		expression, err = r.renderUnion(attribute, concrete, layout)
	default:
		err = fmt.Errorf("unsupported design type %T", attribute.Type)
	}
	if err != nil {
		return "", err
	}
	if !pointer {
		return expression, nil
	}
	local := r.localName()
	r.declarations = append(r.declarations, fmt.Sprintf("var %s %s = %s", local, layout.Def(), expression))
	return "&" + local, nil
}

// renderPrimitive writes a primitive literal and converts it to the planned
// named type when the design or field metadata requires one.
func (r *goValueRenderer) renderPrimitive(attribute *expr.AttributeExpr, value reflect.Value, layout LinkedGoType) (string, error) {
	primitive := underlyingPrimitive(attribute.Type)
	var target string
	if custom, _ := GetMetaType(attribute); custom != "" {
		// String accepts byte values only with an explicit custom target,
		// including one spelled "string". Named custom values keep their
		// source precision and use the final name chosen by the layout.
		if primitive == expr.String && value.Kind() == reflect.Slice ||
			custom != GoNativeTypeName(primitive) && value.Type().PkgPath() != "" {
			target = layout.Def()
		}
	}
	return renderPrimitiveLiteral(primitive, value, target)
}

// renderArray writes every element using the element layout selected by the
// same generator context.
func (r *goValueRenderer) renderArray(attribute *expr.AttributeExpr, value reflect.Value, layout, shape LinkedGoType) (string, error) {
	if value.Kind() != reflect.Array && value.Kind() != reflect.Slice {
		return "", fmt.Errorf("array default has Go type %s", value.Type())
	}
	array := expr.AsArray(attribute.Type)
	items := make([]string, value.Len())
	elementLayout := shape.Enter(shape.plan.element)
	for index := range items {
		item, err := r.render(array.ElemType, value.Index(index), elementLayout, elementLayout.plan.definitionPointer)
		if err != nil {
			return "", fmt.Errorf("array element %d: %w", index, err)
		}
		items[index] = item
	}
	return fmt.Sprintf("%s{%s}", layout.Def(), strings.Join(items, ", ")), nil
}

// renderMap writes map entries in source order independent form so generated
// files remain stable across runs.
func (r *goValueRenderer) renderMap(attribute *expr.AttributeExpr, value reflect.Value, layout, shape LinkedGoType) (string, error) {
	if value.Kind() != reflect.Map {
		return "", fmt.Errorf("map default has Go type %s", value.Type())
	}
	mapped := expr.AsMap(attribute.Type)
	values, err := orderedMapValues(mapped.KeyType, value)
	if err != nil {
		return "", err
	}
	entries := make([]string, 0, len(values))
	keyLayout := shape.Enter(shape.plan.key)
	elementLayout := shape.Enter(shape.plan.element)
	for _, item := range values {
		key, err := r.render(mapped.KeyType, item.key, keyLayout, keyLayout.plan.definitionPointer)
		if err != nil {
			return "", fmt.Errorf("map key: %w", err)
		}
		element, err := r.render(mapped.ElemType, item.value, elementLayout, elementLayout.plan.definitionPointer)
		if err != nil {
			return "", fmt.Errorf("map value for %s: %w", key, err)
		}
		entries = append(entries, key+":"+element)
	}
	return fmt.Sprintf("%s{%s}", layout.Def(), strings.Join(entries, ", ")), nil
}

// renderObject writes only the fields supplied by the authored value. Pointer
// fields use local declarations when Go cannot take a scalar literal address.
func (r *goValueRenderer) renderObject(attribute *expr.AttributeExpr, value reflect.Value, layout, shape LinkedGoType, pointer bool) (string, error) {
	object := expr.AsObject(attribute.Type)
	fields := make([]string, 0, len(*object))
	for index, field := range *object {
		fieldValue, present, err := authoredObjectField(value, field.Name, field.Attribute)
		if err != nil {
			return "", err
		}
		if !present {
			continue
		}
		fieldLayout := shape.Enter(shape.plan.fields[index])
		expression, err := r.render(field.Attribute, fieldValue, fieldLayout, fieldLayout.plan.fieldPointer)
		if err != nil {
			return "", fmt.Errorf("object field %q: %w", field.Name, err)
		}
		name := fieldLayout.Field(true)
		fields = append(fields, name+": "+expression)
	}
	prefix := ""
	if pointer {
		prefix = "&"
	}
	return fmt.Sprintf("%s%s{%s}", prefix, layout.Def(), strings.Join(fields, ", ")), nil
}

// renderUnion selects the authored branch and calls the constructor retained
// by the generated service package.
func (r *goValueRenderer) renderUnion(attribute *expr.AttributeExpr, value reflect.Value, layout LinkedGoType) (string, error) {
	if r.resolveUnion == nil {
		return "", fmt.Errorf("OneOf default requires a planned branch constructor")
	}
	if value.Kind() != reflect.Map {
		return "", fmt.Errorf("OneOf default has Go type %s", value.Type())
	}
	union := expr.AsUnion(attribute.Type)
	tagValue, found := reflectedMapValue(value, union.GetTypeKey())
	if !found {
		return "", fmt.Errorf("OneOf default is missing %q", union.GetTypeKey())
	}
	tag, ok := tagValue.Interface().(string)
	if !ok {
		return "", fmt.Errorf("OneOf default field %q has Go type %s", union.GetTypeKey(), tagValue.Type())
	}
	branchValue, found := reflectedMapValue(value, union.GetValueKey())
	if !found {
		return "", fmt.Errorf("OneOf default is missing %q", union.GetValueKey())
	}
	var (
		branch      *expr.NamedAttributeExpr
		branchIndex int
	)
	for index, candidate := range union.Values {
		if candidate.Name == tag {
			branch = candidate
			branchIndex = index
			break
		}
	}
	if branch == nil {
		return "", fmt.Errorf("OneOf default selects unknown branch %q", tag)
	}
	constructor, err := r.resolveUnion(attribute, tag)
	if err != nil {
		return "", fmt.Errorf("resolve OneOf branch %q: %w", tag, err)
	}
	branchLayout := layout.Enter(layout.plan.branches[branchIndex])
	argument, err := r.render(branch.Attribute, branchValue, branchLayout, branchLayout.plan.referencePointer)
	if err != nil {
		return "", fmt.Errorf("OneOf branch %q: %w", tag, err)
	}
	return fmt.Sprintf("%s(%s)", constructor, argument), nil
}

// localName returns a name that is unique inside the caller-provided block.
func (r *goValueRenderer) localName() string {
	r.nextLocal++
	return r.prefix + strconv.Itoa(r.nextLocal)
}

// orderedMapValues sorts authored map keys before rendering either side of an
// entry. This keeps local variable names stable when map values need pointers.
func orderedMapValues(attribute *expr.AttributeExpr, value reflect.Value) ([]goMapValue, error) {
	if !expr.IsPrimitive(attribute.Type) {
		return nil, fmt.Errorf("map default key uses unsupported design type %T", attribute.Type)
	}
	primitive := underlyingPrimitive(attribute.Type)
	values := make([]goMapValue, 0, value.Len())
	iterator := value.MapRange()
	for iterator.Next() {
		key, err := concreteGoValue(iterator.Key())
		if err != nil {
			return nil, fmt.Errorf("map key: %w", err)
		}
		order, err := renderPrimitiveLiteral(primitive, key, "")
		if err != nil {
			return nil, fmt.Errorf("map key: %w", err)
		}
		values = append(values, goMapValue{order: order, key: iterator.Key(), value: iterator.Value()})
	}
	sort.Slice(values, func(left, right int) bool {
		return values[left].order < values[right].order
	})
	return values, nil
}

// concreteGoValue removes interface and pointer wrappers while rejecting nil,
// which cannot satisfy a non-nil authored default.
func concreteGoValue(value reflect.Value) (reflect.Value, error) {
	if !value.IsValid() {
		return reflect.Value{}, fmt.Errorf("default must not be nil")
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, fmt.Errorf("default must not be nil")
		}
		value = value.Elem()
	}
	return value, nil
}

// renderPrimitiveLiteral checks and writes one primitive value. A nonempty
// target is the planned custom Go type: keep the source numeric precision and
// convert the literal to that type. Schema bounds remain owned by expr.
func renderPrimitiveLiteral(primitive expr.Primitive, value reflect.Value, target string) (string, error) {
	if primitive == expr.Any {
		if target == "" {
			return renderAnyValue(value)
		}
		// Custom Any values retain the scalar and byte forms supported by
		// custom literals; ordinary Any containers use renderAnyValue.
		switch value.Kind() {
		case reflect.Bool:
			primitive = expr.Boolean
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			primitive = expr.Int64
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			primitive = expr.UInt64
		case reflect.Float32, reflect.Float64:
			primitive = expr.Float64
		case reflect.String:
			primitive = expr.String
		case reflect.Slice:
			primitive = expr.Bytes
		default:
			return "", fmt.Errorf("default for custom Go type %s has unsupported underlying type %s", value.Type(), value.Kind())
		}
	}
	var literal string
	switch primitive.Kind() {
	case expr.BooleanKind:
		if value.Kind() != reflect.Bool {
			return "", fmt.Errorf("boolean default has Go type %s", value.Type())
		}
		literal = strconv.FormatBool(value.Bool())
	case expr.IntKind, expr.Int32Kind, expr.Int64Kind:
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			literal = strconv.FormatInt(value.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			literal = strconv.FormatUint(value.Uint(), 10)
		default:
			return "", fmt.Errorf("integer default has Go type %s", value.Type())
		}
	case expr.UIntKind, expr.UInt32Kind, expr.UInt64Kind:
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if value.Int() < 0 {
				return "", fmt.Errorf("unsigned integer default must not be negative")
			}
			literal = strconv.FormatInt(value.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			literal = strconv.FormatUint(value.Uint(), 10)
		default:
			return "", fmt.Errorf("unsigned integer default has Go type %s", value.Type())
		}
	case expr.Float32Kind, expr.Float64Kind:
		var number float64
		switch value.Kind() {
		case reflect.Float32, reflect.Float64:
			number = value.Float()
			if target != "" {
				literal = strconv.FormatFloat(number, 'g', -1, value.Type().Bits())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			number = float64(value.Int())
			if target != "" {
				literal = strconv.FormatInt(value.Int(), 10)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			number = float64(value.Uint())
			if target != "" {
				literal = strconv.FormatUint(value.Uint(), 10)
			}
		default:
			return "", fmt.Errorf("number default has Go type %s", value.Type())
		}
		if math.IsInf(number, 0) || math.IsNaN(number) {
			return "", fmt.Errorf("number default must be finite")
		}
		if target == "" {
			bits := 64
			if primitive.Kind() == expr.Float32Kind {
				bits = 32
			}
			literal = strconv.FormatFloat(number, 'g', -1, bits)
		}
	case expr.StringKind, expr.BytesKind:
		switch value.Kind() {
		case reflect.String:
			literal = strconv.Quote(value.String())
			if primitive == expr.Bytes && target == "" {
				literal = "[]byte(" + literal + ")"
			}
		case reflect.Slice:
			if primitive == expr.String && target == "" || value.Type().Elem().Kind() != reflect.Uint8 {
				return "", fmt.Errorf("%s default has Go type %s", primitive.Name(), value.Type())
			}
			items := make([]string, value.Len())
			for index := range items {
				items[index] = fmt.Sprintf("0x%x", value.Index(index).Uint())
			}
			literal = "[]byte{" + strings.Join(items, ", ") + "}"
		default:
			return "", fmt.Errorf("%s default has Go type %s", primitive.Name(), value.Type())
		}
	default:
		return "", fmt.Errorf("unsupported primitive %s", primitive.Name())
	}
	if target != "" {
		return target + "(" + literal + ")", nil
	}
	return literal, nil
}

// renderAnyValue writes JSON-compatible Go values with explicit container
// types so the generated expression does not depend on a design-time Go type.
func renderAnyValue(value reflect.Value) (string, error) {
	if goValueIsNil(value) {
		return "nil", nil
	}
	concrete, err := concreteGoValue(value)
	if err != nil {
		return "", err
	}
	switch concrete.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(concrete.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(concrete.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(concrete.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		if math.IsInf(concrete.Float(), 0) || math.IsNaN(concrete.Float()) {
			return "", fmt.Errorf("Any number default must be finite")
		}
		return strconv.FormatFloat(concrete.Float(), 'g', -1, concrete.Type().Bits()), nil
	case reflect.String:
		return strconv.Quote(concrete.String()), nil
	case reflect.Array, reflect.Slice:
		items := make([]string, concrete.Len())
		for index := range items {
			items[index], err = renderAnyValue(concrete.Index(index))
			if err != nil {
				return "", err
			}
		}
		return "[]any{" + strings.Join(items, ", ") + "}", nil
	case reflect.Map:
		items := make([]string, 0, concrete.Len())
		iterator := concrete.MapRange()
		for iterator.Next() {
			key, keyErr := concreteGoValue(iterator.Key())
			if keyErr != nil || key.Kind() != reflect.String {
				return "", fmt.Errorf("Any object default key must be a string")
			}
			element, elementErr := renderAnyValue(iterator.Value())
			if elementErr != nil {
				return "", elementErr
			}
			items = append(items, strconv.Quote(key.String())+":"+element)
		}
		sort.Strings(items)
		return "map[string]any{" + strings.Join(items, ", ") + "}", nil
	default:
		return "", fmt.Errorf("Any default has unsupported Go type %s", concrete.Type())
	}
}

// goValueIsNil reports whether an authored Any value is nil after removing
// interface and pointer wrappers.
func goValueIsNil(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}
	return false
}

// authoredObjectField returns one supplied object field from a map or struct.
func authoredObjectField(value reflect.Value, name string, attribute *expr.AttributeExpr) (reflect.Value, bool, error) {
	concrete, err := concreteGoValue(value)
	if err != nil {
		return reflect.Value{}, false, err
	}
	switch concrete.Kind() {
	case reflect.Map:
		field, ok := reflectedMapValue(concrete, name)
		return field, ok, nil
	case reflect.Struct:
		field := concrete.FieldByName(GoifyAtt(attribute, name, true))
		return field, field.IsValid(), nil
	default:
		return reflect.Value{}, false, fmt.Errorf("object default has Go type %s", concrete.Type())
	}
}

// reflectedMapValue finds a string key in maps whose key type may be string or
// any, as used by evaluated design defaults.
func reflectedMapValue(value reflect.Value, key string) (reflect.Value, bool) {
	iterator := value.MapRange()
	for iterator.Next() {
		candidate, err := concreteGoValue(iterator.Key())
		if err == nil && candidate.Kind() == reflect.String && candidate.String() == key {
			return iterator.Value(), true
		}
	}
	return reflect.Value{}, false
}

// underlyingPrimitive returns the primitive beneath a named primitive type.
func underlyingPrimitive(dataType expr.DataType) expr.Primitive {
	for {
		if primitive, ok := dataType.(expr.Primitive); ok {
			return primitive
		}
		dataType = dataType.(expr.UserType).Attribute().Type
	}
}
