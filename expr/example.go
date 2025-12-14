package expr

import (
	"fmt"
	"math"
	"regexp"
	"regexp/syntax"
	"strings"
	"time"
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

	if r.Randomizer == nil {
		return nil
	}

	value, ok := a.Meta.Last("openapi:generate")
	if !ok {
		value, ok = a.Meta.Last("swagger:generate")
	}
	if ok && value == "false" {
		return nil
	}

	value, ok = a.Meta.Last("openapi:example")
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
		switch {
		case math.IsInf(minlength, 1):
			count = int(maxlength) - (r.Int() % 3)
		case math.IsInf(maxlength, -1):
			count = int(minlength) + (r.Int() % 3)
		case minlength < maxlength:
			diff := min(int(maxlength-minlength), maxLength)
			count = int(minlength) + (r.Int() % diff)
		case minlength == maxlength:
			count = int(minlength)
		default:
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
	// Use unaliased type to handle alias types (e.g., Type("Alias", String))
	dt := unalias(a.Type)
	switch dt.Kind() {
	case StringKind:
		return r.Characters(count)
	case BytesKind:
		return []byte(r.Characters(count))
	case MapKind:
		raw := make(map[any]any)
		m := dt.(*Map)
		for range count {
			raw[m.KeyType.Example(r)] = m.ElemType.Example(r)
		}
		return m.MakeMap(raw)
	case ArrayKind:
		raw := make([]any, count)
		ar := dt.(*Array)
		for i := range count {
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
			res, err := syntax.Parse(`([0-9A-F]{2}-){5}[0-9A-F]{2}`, 0)
			if err != nil {
				return "12-34-56-78-9A-BC"
			}
			return patgen(res, r)
		}(),
		FormatCIDR:    "192.168.100.14/24",
		FormatRegexp:  r.Characters(3) + ".*",
		FormatRFC1123: time.Unix(int64(r.Int())%1454957045, 0).UTC().Format(time.RFC1123), // to obtain a "fixed" rand
		FormatUUID:    r.UUID(),
		FormatJSON:    `{"name":"example","email":"mail@example.com"}`,
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
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return r.Name()
	}
	return patgen(re.Simplify(), r)
}

func patgen(re *syntax.Regexp, r *ExampleGenerator) string {
	switch re.Op {
	case syntax.OpAlternate:
		i := r.Int() % len(re.Sub)
		return patgen(re.Sub[i], r)
	case syntax.OpCapture:
		return patgen(re.Sub[0], r)
	case syntax.OpConcat:
		var res strings.Builder
		for _, sub := range re.Sub {
			res.WriteString(patgen(sub, r))
		}
		return res.String()
	case syntax.OpLiteral:
		return string(re.Rune)
	case syntax.OpStar:
		var res strings.Builder
		count := r.Int() % 3
		for range count {
			res.WriteString(patgen(re.Sub[0], r))
		}
		return res.String()
	case syntax.OpPlus:
		var res strings.Builder
		count := r.Int()%2 + 1
		for range count {
			res.WriteString(patgen(re.Sub[0], r))
		}
		return res.String()
	case syntax.OpQuest:
		if r.Int()%2 == 0 {
			return patgen(re.Sub[0], r)
		}
		return ""
	case syntax.OpRepeat:
		var res strings.Builder
		for i := 0; i < re.Min; i++ {
			res.WriteString(patgen(re.Sub[0], r))
		}
		return res.String()
	case syntax.OpCharClass:
		var chars []rune
		for i := 0; i < len(re.Rune); i += 2 {
			start, end := re.Rune[i], re.Rune[i+1]
			for j := start; j <= end; j++ {
				chars = append(chars, j)
			}
		}
		return string(chars[r.Int()%len(chars)])
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return r.Characters(1)
	default:
		return ""
	}
}

func byMinMax(a *AttributeExpr, r *ExampleGenerator) any {
	if !hasMinMaxValidation(a) {
		return nil
	}
	var (
		minimum float64
		maximum = math.Inf(1)
		sign    = 1
	)
	if a.Validation.ExclusiveMaximum != nil {
		maximum = *a.Validation.ExclusiveMaximum
	} else if a.Validation.Maximum != nil {
		maximum = *a.Validation.Maximum
	}
	switch {
	case a.Validation.ExclusiveMinimum != nil:
		minimum = *a.Validation.ExclusiveMinimum
	case a.Validation.Minimum != nil:
		minimum = *a.Validation.Minimum
	default:
		sign = -1
		minimum = maximum
		maximum = math.Inf(1)
	}

	if math.IsInf(maximum, 1) {
		switch a.Type.Kind() {
		case IntKind:
			return sign * (r.Int() + int(minimum))
		case Int32Kind:
			return int32(sign) * (r.Int32() + int32(minimum))
		case Int64Kind:
			return int64(sign) * (r.Int64() + int64(minimum))
		case UIntKind:
			return r.UInt() + uint(minimum)
		case UInt32Kind:
			return r.UInt32() + uint32(minimum)
		case UInt64Kind:
			return r.UInt64() + uint64(minimum)
		case Float32Kind:
			return float32(sign) * (r.Float32() + float32(minimum))
		default:
			return float64(sign) * (r.Float64() + minimum)
		}
	}
	if minimum < maximum {
		delta := maximum - minimum
		switch a.Type.Kind() {
		case IntKind:
			return r.Int()%int(delta) + int(minimum)
		case Int32Kind:
			return r.Int32()%int32(delta) + int32(minimum)
		case Int64Kind:
			return r.Int64()%int64(delta) + int64(minimum)
		case UIntKind:
			return r.UInt()%uint(delta) + uint(minimum)
		case UInt32Kind:
			return r.UInt32()%uint32(delta) + uint32(minimum)
		case UInt64Kind:
			return r.UInt64()%uint64(delta) + uint64(minimum)
		case Float32Kind:
			return r.Float32()*float32(delta) + float32(minimum)
		default:
			return r.Float64()*delta + minimum
		}
	}
	switch a.Type.Kind() {
	case IntKind:
		return int(minimum)
	case Int32Kind:
		return int32(minimum)
	case Int64Kind:
		return int64(minimum)
	case UIntKind:
		return uint(minimum)
	case UInt32Kind:
		return uint32(minimum)
	case UInt64Kind:
		return uint64(minimum)
	case Float32Kind:
		return float32(minimum)
	default:
		return minimum
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
	var minimum *float64
	if a.Validation.ExclusiveMinimum != nil {
		minimum = a.Validation.ExclusiveMinimum
	}
	if a.Validation.Minimum != nil {
		minimum = a.Validation.Minimum
	}
	if minimum != nil {
		if v, ok := example.(int); ok && float64(v) < *minimum {
			return false
		} else if v, ok := example.(float64); ok && v < *minimum {
			return false
		}
	}
	var maximum *float64
	if a.Validation.ExclusiveMaximum != nil {
		maximum = a.Validation.ExclusiveMaximum
	}
	if a.Validation.Maximum != nil {
		maximum = a.Validation.Maximum
	}
	if maximum != nil {
		if v, ok := example.(int); ok && float64(v) > *maximum {
			return false
		} else if v, ok := example.(float64); ok && v > *maximum {
			return false
		}
	}
	return true
}
