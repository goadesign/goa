// This file validates values written with the Default DSL before code
// generation starts. It walks the design type and applies the same validation
// rules that generated boundary validators apply to request values.
package expr

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"unicode/utf8"

	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/internal/codegenname"
	goa "goa.design/goa/v3/pkg"
)

type (
	// defaultValueValidator records errors against the expression that owns the
	// authored default.
	defaultValueValidator struct {
		parent eval.Expression
		errors *eval.ValidationErrors
	}

	// defaultMapEntry keeps map validation deterministic even though Go map
	// iteration order is not stable.
	defaultMapEntry struct {
		key   reflect.Value
		value reflect.Value
		label string
	}
)

// validateDefaultValue verifies the complete value and every supplied value
// below it. Defaults are compile-time design data, so generators may assume
// this method has already proved their type and validation rules.
func (a *AttributeExpr) validateDefaultValue(ctx string, parent eval.Expression) *eval.ValidationErrors {
	if a.DefaultValue == nil {
		return nil
	}
	validator := &defaultValueValidator{
		parent: parent,
		errors: new(eval.ValidationErrors),
	}
	validator.validate(a, reflect.ValueOf(a.DefaultValue), ctx+"default value")
	return validator.errors
}

// validateDefaultValues finds every authored default below an attribute. The
// visited set stops recursive user types while still validating each declared
// attribute once.
func (a *AttributeExpr) validateDefaultValues(ctx string, parent eval.Expression) *eval.ValidationErrors {
	return a.validateDefaultValuesRecursive(ctx, parent, make(map[*AttributeExpr]struct{}))
}

// validateDefaultValuesRecursive validates one attribute and descends through
// each kind that can contain another attribute.
func (a *AttributeExpr) validateDefaultValuesRecursive(ctx string, parent eval.Expression, visited map[*AttributeExpr]struct{}) *eval.ValidationErrors {
	if _, ok := visited[a]; ok {
		return nil
	}
	visited[a] = struct{}{}
	errors := new(eval.ValidationErrors)
	errors.Merge(a.validateDefaultValue(ctx+" - ", parent))
	if _, named := a.Type.(UserType); named {
		return errors
	}
	switch {
	case IsObject(a.Type):
		for _, field := range *AsObject(a.Type) {
			errors.Merge(field.Attribute.validateDefaultValuesRecursive(
				fmt.Sprintf("%s field %q", ctx, field.Name), parent, visited,
			))
		}
	case IsArray(a.Type):
		errors.Merge(AsArray(a.Type).ElemType.validateDefaultValuesRecursive(ctx+" element", parent, visited))
	case IsMap(a.Type):
		mapped := AsMap(a.Type)
		errors.Merge(mapped.KeyType.validateDefaultValuesRecursive(ctx+" map key", parent, visited))
		errors.Merge(mapped.ElemType.validateDefaultValuesRecursive(ctx+" map value", parent, visited))
	case IsUnion(a.Type):
		for _, branch := range AsUnion(a.Type).Values {
			errors.Merge(branch.Attribute.validateDefaultValuesRecursive(
				fmt.Sprintf("%s OneOf branch %q", ctx, branch.Name), parent, visited,
			))
		}
	}
	return errors
}

// validate checks one value and then visits the values selected by its design
// type. A type error stops only that branch because its validations cannot be
// applied safely.
func (v *defaultValueValidator) validate(attribute *AttributeExpr, value reflect.Value, path string) {
	if defaultPrimitive(attribute.Type) == Any {
		v.validateAny(value, path)
		return
	}
	value, ok := concreteDefaultValue(value)
	if !ok {
		v.add("%s must not be nil", path)
		return
	}
	switch {
	case IsPrimitive(attribute.Type):
		if !attribute.Type.IsCompatible(value.Interface()) {
			converted, ok := customPrimitiveDefaultValue(attribute, value)
			if !ok {
				v.addTypeError(path, value, attribute.Type)
				return
			}
			value = converted
		}
		primitive := defaultPrimitive(attribute.Type)
		if !defaultPrimitiveValueFits(primitive, value) {
			v.add("%s value %v is outside the range of %s", path, value.Interface(), primitive.Name())
			return
		}
	case IsArray(attribute.Type):
		if value.Kind() != reflect.Array && value.Kind() != reflect.Slice {
			v.addTypeError(path, value, attribute.Type)
			return
		}
	case IsMap(attribute.Type), IsUnion(attribute.Type):
		if value.Kind() != reflect.Map {
			v.addTypeError(path, value, attribute.Type)
			return
		}
	case IsObject(attribute.Type):
		if value.Kind() != reflect.Map && value.Kind() != reflect.Struct {
			v.addTypeError(path, value, attribute.Type)
			return
		}
	default:
		v.add("%s has unsupported design type %s", path, attribute.Type.Name())
		return
	}

	v.validateRules(attribute, value, path)
	switch {
	case IsArray(attribute.Type):
		v.validateArray(attribute, value, path)
	case IsMap(attribute.Type):
		v.validateMap(attribute, value, path)
	case IsObject(attribute.Type):
		v.validateObject(attribute, value, path)
	case IsUnion(attribute.Type):
		v.validateUnion(attribute, value, path)
	}
}

// validateAny accepts the JSON-shaped values that the Go value renderer can
// specialize into source. Nil is valid inside an Any collection because the
// generated service field itself has type any.
func (v *defaultValueValidator) validateAny(value reflect.Value, path string) {
	if !value.IsValid() {
		return
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.String:
		return
	case reflect.Float32, reflect.Float64:
		if math.IsInf(value.Float(), 0) || math.IsNaN(value.Float()) {
			v.add("%s number must be finite", path)
		}
	case reflect.Array, reflect.Slice:
		for index := range value.Len() {
			v.validateAny(value.Index(index), fmt.Sprintf("%s element %d", path, index))
		}
	case reflect.Map:
		for _, entry := range sortedDefaultMap(value) {
			key, ok := concreteDefaultValue(entry.key)
			if !ok || key.Kind() != reflect.String {
				v.add("%s object key must be a string", path)
				continue
			}
			v.validateAny(entry.value, fmt.Sprintf("%s field %q", path, key.String()))
		}
	default:
		v.add("%s has unsupported Go type %s", path, value.Type())
	}
}

// validateRules applies the validations inherited through primitive aliases
// as well as rules written directly on the attribute.
func (v *defaultValueValidator) validateRules(attribute *AttributeExpr, value reflect.Value, path string) {
	rules := EffectiveValidation(attribute)
	if rules == nil {
		return
	}
	if len(rules.Values) > 0 {
		valid := false
		for _, allowed := range rules.Values {
			if defaultEnumValueEqual(attribute.Type, value, allowed) {
				valid = true
				break
			}
		}
		if !valid {
			v.add("%s must be one of %#v", path, rules.Values)
		}
	}
	if value.Kind() == reflect.String {
		if rules.Format != "" {
			if err := goa.ValidateFormat(path, value.String(), goa.Format(rules.Format)); err != nil {
				v.add("%s does not match format %q", path, rules.Format)
			}
		}
		if rules.Pattern != "" {
			if err := goa.ValidatePattern(path, value.String(), rules.Pattern); err != nil {
				v.add("%s does not match pattern %q", path, rules.Pattern)
			}
		}
	}
	if number, ok := defaultNumber(value); ok {
		if rules.ExclusiveMinimum != nil && number <= *rules.ExclusiveMinimum {
			v.add("%s must be greater than %v", path, *rules.ExclusiveMinimum)
		}
		if rules.Minimum != nil && number < *rules.Minimum {
			v.add("%s must be at least %v", path, *rules.Minimum)
		}
		if rules.ExclusiveMaximum != nil && number >= *rules.ExclusiveMaximum {
			v.add("%s must be less than %v", path, *rules.ExclusiveMaximum)
		}
		if rules.Maximum != nil && number > *rules.Maximum {
			v.add("%s must be at most %v", path, *rules.Maximum)
		}
	}
	if length, ok := defaultLength(value); ok {
		if rules.MinLength != nil && length < *rules.MinLength {
			v.add("%s length must be at least %d", path, *rules.MinLength)
		}
		if rules.MaxLength != nil && length > *rules.MaxLength {
			v.add("%s length must be at most %d", path, *rules.MaxLength)
		}
	}
}

// defaultEnumValueEqual compares numeric enum values after converting both to
// the primitive selected by the design. Generated Go comparisons apply this
// same conversion to untyped enum constants.
func defaultEnumValueEqual(dataType DataType, value reflect.Value, allowed any) bool {
	primitive := defaultPrimitive(dataType)
	if primitive < Int || primitive > Float64 {
		return reflect.DeepEqual(value.Interface(), allowed)
	}
	allowedValue, ok := concreteDefaultValue(reflect.ValueOf(allowed))
	if !ok || !primitive.IsCompatible(allowedValue.Interface()) {
		return false
	}
	target := defaultPrimitiveReflectType(primitive)
	if !value.Type().ConvertibleTo(target) || !allowedValue.Type().ConvertibleTo(target) {
		return false
	}
	return reflect.DeepEqual(value.Convert(target).Interface(), allowedValue.Convert(target).Interface())
}

// validateArray applies the element contract to every authored element.
func (v *defaultValueValidator) validateArray(attribute *AttributeExpr, value reflect.Value, path string) {
	element := AsArray(attribute.Type).ElemType
	for index := range value.Len() {
		v.validate(element, value.Index(index), fmt.Sprintf("%s element %d", path, index))
	}
}

// validateMap applies both key and value contracts to every authored entry.
func (v *defaultValueValidator) validateMap(attribute *AttributeExpr, value reflect.Value, path string) {
	mapped := AsMap(attribute.Type)
	for _, entry := range sortedDefaultMap(value) {
		v.validate(mapped.KeyType, entry.key, path+" map key")
		v.validate(mapped.ElemType, entry.value, fmt.Sprintf("%s map value for %s", path, entry.label))
	}
}

// validateObject checks required fields, rejects unknown fields, and applies
// each supplied field's own contract.
func (v *defaultValueValidator) validateObject(attribute *AttributeExpr, value reflect.Value, path string) {
	object := AsObject(attribute.Type)
	fields := make(map[string]*AttributeExpr, len(*object))
	for _, field := range *object {
		fields[field.Name] = field.Attribute
	}
	values, invalidKeys := defaultObjectFields(value, object)
	for _, keyType := range invalidKeys {
		v.add("%s object key has type %s, expected string", path, keyType)
	}
	for _, required := range attribute.AllRequired() {
		if _, ok := values[required]; !ok {
			v.add("%s is missing required field %q", path, required)
		}
	}
	for _, name := range sortedDefaultNames(values) {
		field := fields[name]
		if field == nil {
			v.add("%s contains unknown field %q", path, name)
			continue
		}
		v.validate(field, values[name], fmt.Sprintf("%s field %q", path, name))
	}
}

// defaultObjectFields returns values using their design names. Struct fields
// use the same Go names as generated object literals.
func defaultObjectFields(value reflect.Value, object *Object) (map[string]reflect.Value, []string) {
	if value.Kind() == reflect.Map {
		return defaultStringMap(value)
	}
	values := make(map[string]reflect.Value, len(*object))
	for _, field := range *object {
		name := codegenname.AttributeName(field.Name, field.Attribute.Meta["struct:field:name"])
		fieldValue := value.FieldByName(codegenname.Goify(name, true))
		if fieldValue.IsValid() {
			values[field.Name] = fieldValue
		}
	}
	return values, nil
}

// validateUnion requires the canonical tagged envelope and validates only the
// branch named by its discriminator.
func (v *defaultValueValidator) validateUnion(attribute *AttributeExpr, value reflect.Value, path string) {
	union := AsUnion(attribute.Type)
	values, invalidKeys := defaultStringMap(value)
	if len(invalidKeys) > 0 {
		for _, keyType := range invalidKeys {
			v.add("%s OneOf key has type %s, expected string", path, keyType)
		}
		return
	}
	tagValue, hasTag := values[union.GetTypeKey()]
	branchValue, hasValue := values[union.GetValueKey()]
	if !hasTag {
		v.add("%s is missing OneOf discriminator %q", path, union.GetTypeKey())
		return
	}
	tagValue, ok := concreteDefaultValue(tagValue)
	if !ok || tagValue.Kind() != reflect.String {
		v.add("%s OneOf discriminator %q must be a string", path, union.GetTypeKey())
		return
	}
	tag := tagValue.String()
	var branch *AttributeExpr
	for _, candidate := range union.Values {
		if candidate.Name == tag {
			branch = candidate.Attribute
			break
		}
	}
	if branch == nil {
		v.add("%s selects unknown OneOf branch %q", path, tag)
		return
	}
	if !hasValue {
		v.add("%s is missing the value for OneOf branch %q", path, tag)
		return
	}
	if len(values) != 2 {
		v.add("%s OneOf value must contain only %q and %q", path, union.GetTypeKey(), union.GetValueKey())
	}
	v.validate(branch, branchValue, fmt.Sprintf("%s OneOf branch %q", path, tag))
}

// add records one design error with the expression that owns the default.
func (v *defaultValueValidator) add(format string, args ...any) {
	v.errors.Add(v.parent, format, args...)
}

// addTypeError reports the concrete authored Go type and expected Goa type.
func (v *defaultValueValidator) addTypeError(path string, value reflect.Value, dataType DataType) {
	v.add("%s has type %s, expected %s", path, value.Type(), dataType.Name())
}

// concreteDefaultValue removes interfaces and pointers while preserving the
// concrete value used for type and validation checks.
func concreteDefaultValue(value reflect.Value) (reflect.Value, bool) {
	if !value.IsValid() {
		return reflect.Value{}, false
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return reflect.Value{}, false
		}
		value = value.Elem()
	}
	return value, true
}

// defaultPrimitive returns the primitive below a chain of named types.
func defaultPrimitive(dataType DataType) Primitive {
	for {
		if primitive, ok := dataType.(Primitive); ok {
			return primitive
		}
		userType, ok := dataType.(UserType)
		if !ok {
			return 0
		}
		dataType = userType.Attribute().Type
	}
}

// customPrimitiveDefaultValue converts a value already declared with the exact
// custom Go type into the primitive value used by design validations. The
// metadata contains the package path and type name, so this check never needs
// to load or inspect a Go package.
func customPrimitiveDefaultValue(attribute *AttributeExpr, value reflect.Value) (reflect.Value, bool) {
	metadata := attribute.Meta["struct:field:type"]
	if len(metadata) < 2 || value.Type().PkgPath() != metadata[1] {
		return reflect.Value{}, false
	}
	name := metadata[0]
	for index := len(name) - 1; index >= 0; index-- {
		if name[index] == '.' {
			name = name[index+1:]
			break
		}
	}
	if value.Type().Name() != name {
		return reflect.Value{}, false
	}
	target := defaultPrimitiveReflectType(defaultPrimitive(attribute.Type))
	if target == nil || !value.Type().ConvertibleTo(target) {
		return reflect.Value{}, false
	}
	return value.Convert(target), true
}

// defaultPrimitiveReflectType returns the native Go type used to evaluate a
// primitive's design validations.
func defaultPrimitiveReflectType(primitive Primitive) reflect.Type {
	switch primitive {
	case Boolean:
		return reflect.TypeOf(false)
	case Int:
		return reflect.TypeOf(int(0))
	case Int32:
		return reflect.TypeOf(int32(0))
	case Int64:
		return reflect.TypeOf(int64(0))
	case UInt:
		return reflect.TypeOf(uint(0))
	case UInt32:
		return reflect.TypeOf(uint32(0))
	case UInt64:
		return reflect.TypeOf(uint64(0))
	case Float32:
		return reflect.TypeOf(float32(0))
	case Float64:
		return reflect.TypeOf(float64(0))
	case String:
		return reflect.TypeOf("")
	case Bytes:
		return reflect.TypeOf([]byte(nil))
	default:
		return nil
	}
}

// defaultPrimitiveValueFits reports whether a compatible numeric value can be
// represented by the generated primitive type without overflow.
func defaultPrimitiveValueFits(primitive Primitive, value reflect.Value) bool {
	switch primitive {
	case Int:
		return defaultSignedValueFits(value, 0)
	case Int32:
		return defaultSignedValueFits(value, 32)
	case Int64:
		return defaultSignedValueFits(value, 64)
	case UInt:
		return defaultUnsignedValueFits(value, 0)
	case UInt32:
		return defaultUnsignedValueFits(value, 32)
	case UInt64:
		return defaultUnsignedValueFits(value, 64)
	case Float32:
		number, ok := defaultNumber(value)
		return ok && !math.IsInf(number, 0) && !math.IsNaN(number) && math.Abs(number) <= math.MaxFloat32
	case Float64:
		number, ok := defaultNumber(value)
		return ok && !math.IsInf(number, 0) && !math.IsNaN(number)
	default:
		return true
	}
}

// defaultSignedValueFits checks signed bounds. A zero bit count means the
// generated Go int width on the current generation platform.
func defaultSignedValueFits(value reflect.Value, bits int) bool {
	if bits == 0 {
		bits = reflect.TypeOf(int(0)).Bits()
	}
	var maximum uint64
	if bits == 64 {
		maximum = math.MaxInt64
	} else {
		maximum = uint64(1)<<(bits-1) - 1
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number := value.Int()
		minimum := -int64(maximum) - 1
		return number >= minimum && (number < 0 || uint64(number) <= maximum)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() <= maximum
	default:
		return false
	}
}

// defaultUnsignedValueFits checks unsigned bounds. A zero bit count means the
// generated Go uint width on the current generation platform.
func defaultUnsignedValueFits(value reflect.Value, bits int) bool {
	if bits == 0 {
		bits = reflect.TypeOf(uint(0)).Bits()
	}
	maximum := uint64(math.MaxUint64)
	if bits < 64 {
		maximum = uint64(1)<<bits - 1
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() >= 0 && uint64(value.Int()) <= maximum
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() <= maximum
	default:
		return false
	}
}

// defaultNumber converts a numeric design value to the representation used by
// numeric validation expressions.
func defaultNumber(value reflect.Value) (float64, bool) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return value.Float(), true
	default:
		return 0, false
	}
}

// defaultLength returns the length measured by generated validators. String
// lengths count Unicode characters; collections and bytes count elements.
func defaultLength(value reflect.Value) (int, bool) {
	switch value.Kind() {
	case reflect.String:
		return utf8.RuneCountInString(value.String()), true
	case reflect.Array, reflect.Slice, reflect.Map:
		return value.Len(), true
	default:
		return 0, false
	}
}

// defaultStringMap returns an object's or OneOf's authored fields. Callers
// already proved that value is a map.
func defaultStringMap(value reflect.Value) (map[string]reflect.Value, []string) {
	values := make(map[string]reflect.Value, value.Len())
	var invalid []string
	iterator := value.MapRange()
	for iterator.Next() {
		key, ok := concreteDefaultValue(iterator.Key())
		if !ok || key.Kind() != reflect.String {
			keyType := "nil"
			if key.IsValid() {
				keyType = key.Type().String()
			}
			invalid = append(invalid, keyType)
			continue
		}
		values[key.String()] = iterator.Value()
	}
	sort.Strings(invalid)
	return values, invalid
}

// sortedDefaultMap returns map entries ordered by their printed key so error
// order does not change between evaluations.
func sortedDefaultMap(value reflect.Value) []defaultMapEntry {
	entries := make([]defaultMapEntry, 0, value.Len())
	iterator := value.MapRange()
	for iterator.Next() {
		key := iterator.Key()
		entries = append(entries, defaultMapEntry{
			key:   key,
			value: iterator.Value(),
			label: fmt.Sprintf("%#v", key.Interface()),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].label < entries[j].label
	})
	return entries
}

// sortedDefaultNames returns object field names in stable order.
func sortedDefaultNames(values map[string]reflect.Value) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
