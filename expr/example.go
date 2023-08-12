package expr

import (
	"fmt"
	"math"
	"regexp"
	"time"

	regen "github.com/AnatolyRugalev/goregen"
	googleuuid "github.com/google/uuid"
)

const (
	maxAttempts = 500 // Max number of retries to generate valid example.
	maxLength   = 3   // Max length for array and map examples.
)

// Example returns the example set on the attribute at design time. If there
// isn't such a value then Example computes a random value for the attribute
// using the given random value producer.
func (a *AttributeExpr) Example(r *ExampleGenerator) any {
	if ex := a.ExtractUserExamples(); len(ex) > 0 {
		// Return the last item in the slice so that examples can be overridden
		// in the DSL. Overridden examples are always appended to the UserExamples
		// slice.
		return ex[len(ex)-1].Value
	}

	value, ok := a.Meta.Last("openapi:example")
	if !ok {
		value, ok = a.Meta.Last("swagger:example")
	}
	if ok && value == "false" {
		return nil
	}

	// randomize array length first, since that's from higher level
	if hasLengthValidation(a) {
		return byLength(a, r)
	}
	// enum should dominate, because the potential "examples" are fixed
	if hasEnumValidation(a) {
		return byEnum(a, r)
	}
	// loop until a satisfying example is generated
	var (
		hasFormat  = hasFormatValidation(a)
		hasPattern = hasPatternValidation(a)
		hasMinMax  = hasMinMaxValidation(a)
		attempts   = 0
	)
	for attempts < maxAttempts {
		attempts++
		var example any
		// Format comes first, since it initiates the example
		if hasFormat {
			example = byFormat(a, r)
		}
		// now validate with rest of matchers; redo if not satisified
		if hasPattern {
			if example == nil {
				example = byPattern(a, r)
			} else if !checkPattern(a, example) {
				continue
			}
		}
		if hasMinMax {
			if example == nil {
				example = byMinMax(a, r)
			} else if !checkMinMaxValue(a, example) {
				continue
			}
		}
		if example == nil {
			example = a.Type.Example(r)
		}
		return example
	}
	return a.Type.Example(r)
}

// NewLength returns an int that validates the generator attribute length
// validations if any.
func NewLength(a *AttributeExpr, r *ExampleGenerator) int {
	if hasLengthValidation(a) {
		minlength, maxlength := math.Inf(1), math.Inf(-1)
		if a.Validation.MinLength != nil {
			minlength = float64(*a.Validation.MinLength)
		}
		if a.Validation.MaxLength != nil {
			maxlength = float64(*a.Validation.MaxLength)
		}
		count := 0
		if math.IsInf(minlength, 1) {
			count = int(maxlength) - (r.Int() % 3)
		} else if math.IsInf(maxlength, -1) {
			count = int(minlength) + (r.Int() % 3)
		} else if minlength < maxlength {
			diff := int(maxlength - minlength)
			if diff > maxLength {
				diff = maxLength
			}
			count = int(minlength) + (r.Int() % diff)
		} else if minlength == maxlength {
			count = int(minlength)
		} else {
			panic("Validation: MinLength > MaxLength")
		}
		if count > maxLength {
			count = maxLength
		}
		return count
	}
	return r.ArrayLength()
}

func hasLengthValidation(a *AttributeExpr) bool {
	if a.Validation == nil {
		return false
	}
	return a.Validation.MinLength != nil || a.Validation.MaxLength != nil
}

func hasEnumValidation(a *AttributeExpr) bool {
	return a.Validation != nil && len(a.Validation.Values) > 0
}

func hasFormatValidation(a *AttributeExpr) bool {
	return a.Validation != nil && a.Validation.Format != ""
}

func hasPatternValidation(a *AttributeExpr) bool {
	return a.Validation != nil && a.Validation.Pattern != ""
}

func hasMinMaxValidation(a *AttributeExpr) bool {
	if a.Validation == nil {
		return false
	}
	return a.Validation.ExclusiveMinimum != nil ||
		a.Validation.Minimum != nil ||
		a.Validation.ExclusiveMaximum != nil ||
		a.Validation.Maximum != nil
}

// byLength generates a random size array of examples based on what's given.
func byLength(a *AttributeExpr, r *ExampleGenerator) any {
	count := NewLength(a, r)
	switch a.Type.Kind() {
	case StringKind:
		return r.Characters(count)
	case BytesKind:
		return []byte(r.Characters(count))
	case MapKind:
		raw := make(map[any]any)
		m := a.Type.(*Map)
		for i := 0; i < count; i++ {
			raw[m.KeyType.Example(r)] = m.ElemType.Example(r)
		}
		return m.MakeMap(raw)
	case ArrayKind:
		raw := make([]any, count)
		ar := a.Type.(*Array)
		for i := 0; i < count; i++ {
			raw[i] = ar.ElemType.Example(r)
		}
		return ar.MakeSlice(raw)
	default:
		panic("invalid type for length validation: " + a.Type.Name())
	}
}

// byEnum returns a random selected enum value.
func byEnum(a *AttributeExpr, r *ExampleGenerator) any {
	if !hasEnumValidation(a) {
		return nil
	}
	values := a.Validation.Values
	count := len(values)
	i := r.Int() % count
	return values[i]
}

// byFormat returns a random example based on the format the user asks.
func byFormat(a *AttributeExpr, r *ExampleGenerator) any {
	if !hasFormatValidation(a) {
		return nil
	}
	format := a.Validation.Format
	if res, ok := map[ValidationFormat]any{
		FormatEmail:    r.Email(),
		FormatHostname: r.Hostname(),
		FormatDate:     time.Unix(int64(r.Int())%1454957045, 0).UTC().Format(time.DateOnly), // to obtain a "fixed" rand
		FormatDateTime: time.Unix(int64(r.Int())%1454957045, 0).UTC().Format(time.RFC3339),  // to obtain a "fixed" rand
		FormatIPv4:     r.IPv4Address().String(),
		FormatIPv6:     r.IPv6Address().String(),
		FormatIP:       r.IPv4Address().String(),
		FormatURI:      r.URL(),
		FormatMAC: func() string {
			res, err := regen.Generate(`([0-9A-F]{2}-){5}[0-9A-F]{2}`)
			if err != nil {
				return "12-34-56-78-9A-BC"
			}
			return res
		}(),
		FormatCIDR:    "192.168.100.14/24",
		FormatRegexp:  r.Characters(3) + ".*",
		FormatRFC1123: time.Unix(int64(r.Int())%1454957045, 0).UTC().Format(time.RFC1123), // to obtain a "fixed" rand
		FormatUUID: func() string {
			uuid, err := googleuuid.NewUUID()
			if err != nil {
				return "12345678-1234-1234-9232-123456789ABC"
			}
			return uuid.String()
		}(),
		FormatJSON: `{"name":"example","email":"mail@example.com"}`,
	}[format]; ok {
		return res
	}
	panic("Validation: unknown format '" + format + "'") // bug
}

// byPattern generates a random value that satisfies the pattern.
//
// Note: if multiple patterns are given, only one of them is used.
func byPattern(a *AttributeExpr, r *ExampleGenerator) any {
	if !hasPatternValidation(a) {
		return false
	}
	pattern := a.Validation.Pattern
	gen, err := regen.NewGenerator(pattern, &regen.GeneratorArgs{MaxUnboundedRepeatCount: 6})
	if err != nil {
		return r.Name()
	}
	return gen.Generate()
}

func byMinMax(a *AttributeExpr, r *ExampleGenerator) any {
	if !hasMinMaxValidation(a) {
		return nil
	}
	var (
		min  = math.Inf(-1)
		max  = math.Inf(1)
		sign = 1
	)
	if a.Validation.ExclusiveMaximum != nil {
		max = *a.Validation.ExclusiveMaximum
	} else if a.Validation.Maximum != nil {
		max = *a.Validation.Maximum
	}
	if a.Validation.ExclusiveMinimum != nil {
		min = *a.Validation.ExclusiveMinimum
	} else if a.Validation.Minimum != nil {
		min = *a.Validation.Minimum
	} else {
		sign = -1
		min = max
		max = math.Inf(1)
	}

	if math.IsInf(max, 1) {
		switch a.Type.Kind() {
		case IntKind:
			return sign * (r.Int() + int(min))
		case Int32Kind:
			return int32(sign) * (r.Int32() + int32(min))
		case Int64Kind:
			return int64(sign) * (r.Int64() + int64(min))
		case UIntKind:
			return r.UInt() + uint(min)
		case UInt32Kind:
			return r.UInt32() + uint32(min)
		case UInt64Kind:
			return r.UInt64() + uint64(min)
		case Float32Kind:
			return float32(sign) * (r.Float32() + float32(min))
		default:
			return float64(sign) * (r.Float64() + min)
		}
	}
	if min < max {
		delta := max - min
		switch a.Type.Kind() {
		case IntKind:
			return r.Int()%int(delta) + int(min)
		case Int32Kind:
			return r.Int32()%int32(delta) + int32(min)
		case Int64Kind:
			return r.Int64()%int64(delta) + int64(min)
		case UIntKind:
			return r.UInt()%uint(delta) + uint(min)
		case UInt32Kind:
			return r.UInt32()%uint32(delta) + uint32(min)
		case UInt64Kind:
			return r.UInt64()%uint64(delta) + uint64(min)
		case Float32Kind:
			return r.Float32()*float32(delta) + float32(min)
		default:
			return r.Float64()*delta + min
		}
	}
	switch a.Type.Kind() {
	case IntKind:
		return int(min)
	case Int32Kind:
		return int32(min)
	case Int64Kind:
		return int64(min)
	case UIntKind:
		return uint(min)
	case UInt32Kind:
		return uint32(min)
	case UInt64Kind:
		return uint64(min)
	case Float32Kind:
		return float32(min)
	default:
		return min
	}
}

func checkPattern(a *AttributeExpr, example any) bool {
	if !hasPatternValidation(a) {
		return true
	}
	pattern := a.Validation.Pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("Validation: invalid pattern '" + pattern + "'")
	}
	if !re.MatchString(fmt.Sprint(example)) {
		return false
	}
	return true
}

func checkMinMaxValue(a *AttributeExpr, example any) bool {
	if !hasMinMaxValidation(a) {
		return true
	}
	var min *float64
	if a.Validation.ExclusiveMinimum != nil {
		min = a.Validation.ExclusiveMinimum
	}
	if a.Validation.Minimum != nil {
		min = a.Validation.Minimum
	}
	if min != nil {
		if v, ok := example.(int); ok && float64(v) < *min {
			return false
		} else if v, ok := example.(float64); ok && v < *min {
			return false
		}
	}
	var max *float64
	if a.Validation.ExclusiveMaximum != nil {
		max = a.Validation.ExclusiveMaximum
	}
	if a.Validation.Maximum != nil {
		max = a.Validation.Maximum
	}
	if max != nil {
		if v, ok := example.(int); ok && float64(v) > *max {
			return false
		} else if v, ok := example.(float64); ok && v > *max {
			return false
		}
	}
	return true
}
