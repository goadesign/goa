package design

import (
	"fmt"
	"math"
	"regexp"
	"time"

	regen "github.com/zach-klippenstein/goregen"
)

const (
	maxAttempts = 500  // Maximum number of retries when generating validated example.
	maxLength   = 3    // Maximum length for array and map examples.
	maxValue    = 1000 // Maximum value for integer and float examples.
)

// Example returns a random value for the generator attribute using the generator random value
// producer.
func (a *AttributeExpr) Example(r *Random) interface{} {
	// Randomize array length first, since that's from higher level
	if hasLengthValidation(a) {
		return byLength(a, r)
	}
	// Enum should dominate, because the potential "examples" are fixed
	if hasEnumValidation(a) {
		return byEnum(a, r)
	}
	// loop until a satisified example is generated
	hasFormat, hasPattern, hasMinMax := hasFormatValidation(a), hasPatternValidation(a), hasMinMaxValidation(a)
	attempts := 0
	for attempts < maxAttempts {
		attempts++
		var example interface{}
		// Format comes first, since it initiates the example
		if hasFormat {
			example = byFormat(a, r)
		}
		// now validate with the rest of matchers; if not satisified, redo
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

// NewLength returns an int that validates the generator attribute length validations if any.
func NewLength(a *AttributeExpr, r *Random) int {
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
	return r.Int()%3 + 1
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
	return a.Validation.Minimum != nil || a.Validation.Maximum != nil
}

// byLength generates a random size array of examples based on what's given.
func byLength(a *AttributeExpr, r *Random) interface{} {
	count := NewLength(a, r)
	if a.Type.Kind() == StringKind {
		return r.faker.Characters(count)
	}
	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		res[i] = a.Type.(*Array).Example(r)
	}
	return res
}

// byEnum returns a random selected enum value.
func byEnum(a *AttributeExpr, r *Random) interface{} {
	if !hasEnumValidation(a) {
		return nil
	}
	values := a.Validation.Values
	count := len(values)
	i := r.Int() % count
	return values[i]
}

// byFormat returns a random example based on the format the user asks.
func byFormat(a *AttributeExpr, r *Random) interface{} {
	if !hasFormatValidation(a) {
		return nil
	}
	format := a.Validation.Format
	if res, ok := map[string]interface{}{
		"email":     r.faker.Email(),
		"hostname":  r.faker.DomainName() + "." + r.faker.DomainSuffix(),
		"date-time": time.Unix(int64(r.Int())%1454957045, 0).Format(time.RFC3339), // to obtain a "fixed" rand
		"ipv4":      r.faker.IPv4Address().String(),
		"ipv6":      r.faker.IPv6Address().String(),
		"ip":        r.faker.IPv4Address().String(),
		"uri":       r.faker.URL(),
		"mac": func() string {
			res, err := regen.Generate(`([0-9A-F]{2}-){5}[0-9A-F]{2}`)
			if err != nil {
				return "12-34-56-78-9A-BC"
			}
			return res
		}(),
		"cidr":   "192.168.100.14/24",
		"regexp": r.faker.Characters(3) + ".*",
	}[format]; ok {
		return res
	}
	panic("Validation: unknown format '" + format + "'") // bug
}

// byPattern generates a random value that satisifies the pattern. Note: if multiple patterns are
// given, only one of them is used. currently, it doesn't support multiple.
func byPattern(a *AttributeExpr, r *Random) interface{} {
	if !hasPatternValidation(a) {
		return false
	}
	pattern := a.Validation.Pattern
	example, err := regen.Generate(pattern)
	if err != nil {
		return r.faker.Name()
	}
	return example
}

func byMinMax(a *AttributeExpr, r *Random) interface{} {
	if !hasMinMaxValidation(a) {
		return nil
	}
	var (
		i32  = a.Type.Kind() == Int32Kind || a.Type.Kind() == UInt32Kind
		i64  = a.Type.Kind() == Int64Kind || a.Type.Kind() == UInt64Kind
		f32  = a.Type.Kind() == Float32Kind
		min  = math.Inf(-1)
		max  = math.Inf(1)
		sign = 1
	)
	if a.Validation.Maximum != nil {
		max = *a.Validation.Maximum
	}
	if a.Validation.Minimum != nil {
		min = *a.Validation.Minimum
	} else {
		sign = -1
		min = max
		max = math.Inf(1)
	}

	if math.IsInf(max, 1) {
		switch {
		case i32:
			return int32(sign) * (r.Int32() + int32(min))
		case i64:
			return int64(sign) * (r.Int64() + int64(min))
		case f32:
			return float32(sign) * (r.Float32() + float32(min))
		default:
			return float64(sign) * (r.Float64() + min)
		}
	}
	if min < max {
		delta := max - min
		switch {
		case i32:
			return r.Int32()%int32(delta) + int32(min)
		case i64:
			return r.Int64()%int64(delta) + int64(min)
		case f32:
			return r.Float32()*float32(delta) + float32(min)
		default:
			return r.Float64()*delta + min
		}
	}
	switch {
	case i32:
		return int32(min)
	case i64:
		return int64(min)
	case f32:
		return float32(min)
	default:
		return min
	}
}

func checkPattern(a *AttributeExpr, example interface{}) bool {
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

func checkMinMaxValue(a *AttributeExpr, example interface{}) bool {
	if !hasMinMaxValidation(a) {
		return true
	}
	if min := a.Validation.Minimum; min != nil {
		if v, ok := example.(int); ok && float64(v) < *min {
			return false
		} else if v, ok := example.(float64); ok && v < *min {
			return false
		}
	}
	if max := a.Validation.Maximum; max != nil {
		if v, ok := example.(int); ok && float64(v) > *max {
			return false
		} else if v, ok := example.(float64); ok && v > *max {
			return false
		}
	}
	return true
}
