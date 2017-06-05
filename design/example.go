package design

import (
	"fmt"
	"math"
	"regexp"
	"time"

	regen "github.com/zach-klippenstein/goregen"
)

// exampleGenerator generates a random example based on the given validations on the definition.
type exampleGenerator struct {
	a *AttributeDefinition
	r *RandomGenerator
}

// newExampleGenerator returns an example generator that uses the given random generator.
func newExampleGenerator(a *AttributeDefinition, r *RandomGenerator) *exampleGenerator {
	return &exampleGenerator{a, r}
}

// Maximum number of tries for generating example.
const maxAttempts = 500

// Generate generates a random value based on the given validations.
func (eg *exampleGenerator) Generate(seen []string) interface{} {
	// Randomize array length first, since that's from higher level
	if eg.hasLengthValidation() {
		return eg.generateValidatedLengthExample(seen)
	}
	// Enum should dominate, because the potential "examples" are fixed
	if eg.hasEnumValidation() {
		return eg.generateValidatedEnumExample()
	}
	// loop until a satisified example is generated
	hasFormat, hasPattern, hasMinMax := eg.hasFormatValidation(), eg.hasPatternValidation(), eg.hasMinMaxValidation()
	attempts := 0
	for attempts < maxAttempts {
		attempts++
		var example interface{}
		// Format comes first, since it initiates the example
		if hasFormat {
			example = eg.generateFormatExample()
		}
		// now validate with the rest of matchers; if not satisified, redo
		if hasPattern {
			if example == nil {
				example = eg.generateValidatedPatternExample()
			} else if !eg.checkPatternValidation(example) {
				continue
			}
		}
		if hasMinMax {
			if example == nil {
				example = eg.generateValidatedMinMaxValueExample()
			} else if !eg.checkMinMaxValueValidation(example) {
				continue
			}
		}
		if example == nil {
			example = eg.a.Type.GenerateExample(eg.r, seen)
		}
		return example
	}
	return eg.a.Type.GenerateExample(eg.r, seen)
}

func (eg *exampleGenerator) ExampleLength() int {
	if eg.hasLengthValidation() {
		minlength, maxlength := math.Inf(1), math.Inf(-1)
		if eg.a.Validation.MinLength != nil {
			minlength = float64(*eg.a.Validation.MinLength)
		}
		if eg.a.Validation.MaxLength != nil {
			maxlength = float64(*eg.a.Validation.MaxLength)
		}
		count := 0
		if math.IsInf(minlength, 1) {
			count = int(maxlength) - (eg.r.Int() % 3)
		} else if math.IsInf(maxlength, -1) {
			count = int(minlength) + (eg.r.Int() % 3)
		} else if minlength < maxlength {
			diff := int(maxlength - minlength)
			if diff > maxExampleLength {
				diff = maxExampleLength
			}
			count = int(minlength) + (eg.r.Int() % diff)
		} else if minlength == maxlength {
			count = int(minlength)
		} else {
			panic("Validation: MinLength > MaxLength")
		}
		if count > maxExampleLength {
			count = maxExampleLength
		}
		if count <= 0 && maxlength != 0 {
			count = 1
		}
		return count
	}
	return eg.r.Int()%3 + 1
}

func (eg *exampleGenerator) hasLengthValidation() bool {
	if eg.a.Validation == nil {
		return false
	}
	return eg.a.Validation.MinLength != nil || eg.a.Validation.MaxLength != nil
}

const maxExampleLength = 10

// generateValidatedLengthExample generates a random size array of examples based on what's given.
func (eg *exampleGenerator) generateValidatedLengthExample(seen []string) interface{} {
	count := eg.ExampleLength()
	if !eg.a.Type.IsArray() {
		return eg.r.faker.Characters(count)
	}
	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		res[i] = eg.a.Type.ToArray().ElemType.GenerateExample(eg.r, seen)
	}
	return res
}

func (eg *exampleGenerator) hasEnumValidation() bool {
	return eg.a.Validation != nil && len(eg.a.Validation.Values) > 0
}

// generateValidatedEnumExample returns a random selected enum value.
func (eg *exampleGenerator) generateValidatedEnumExample() interface{} {
	if !eg.hasEnumValidation() {
		return nil
	}
	values := eg.a.Validation.Values
	count := len(values)
	i := eg.r.Int() % count
	return values[i]
}

func (eg *exampleGenerator) hasFormatValidation() bool {
	return eg.a.Validation != nil && eg.a.Validation.Format != ""
}

// generateFormatExample returns a random example based on the format the user asks.
func (eg *exampleGenerator) generateFormatExample() interface{} {
	if !eg.hasFormatValidation() {
		return nil
	}
	format := eg.a.Validation.Format
	if res, ok := map[string]interface{}{
		"email":     eg.r.faker.Email(),
		"hostname":  eg.r.faker.DomainName() + "." + eg.r.faker.DomainSuffix(),
		"date-time": time.Unix(int64(eg.r.Int())%1454957045, 0).Format(time.RFC3339), // to obtain a "fixed" rand
		"ipv4":      eg.r.faker.IPv4Address().String(),
		"ipv6":      eg.r.faker.IPv6Address().String(),
		"ip":        eg.r.faker.IPv4Address().String(),
		"uri":       eg.r.faker.URL(),
		"mac": func() string {
			res, err := regen.Generate(`([0-9A-F]{2}-){5}[0-9A-F]{2}`)
			if err != nil {
				return "12-34-56-78-9A-BC"
			}
			return res
		}(),
		"cidr":    "192.168.100.14/24",
		"regexp":  eg.r.faker.Characters(3) + ".*",
		"rfc1123": time.Unix(int64(eg.r.Int())%1454957045, 0).Format(time.RFC1123), // to obtain a "fixed" rand
	}[format]; ok {
		return res
	}
	panic("Validation: unknown format '" + format + "'") // bug
}

func (eg *exampleGenerator) hasPatternValidation() bool {
	return eg.a.Validation != nil && eg.a.Validation.Pattern != ""
}

func (eg *exampleGenerator) checkPatternValidation(example interface{}) bool {
	if !eg.hasPatternValidation() {
		return true
	}
	pattern := eg.a.Validation.Pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic("Validation: invalid pattern '" + pattern + "'")
	}
	if !re.MatchString(fmt.Sprint(example)) {
		return false
	}
	return true
}

// generateValidatedPatternExample generates a random value that satisifies the pattern. Note: if
// multiple patterns are given, only one of them is used. currently, it doesn't support multiple.
func (eg *exampleGenerator) generateValidatedPatternExample() interface{} {
	if !eg.hasPatternValidation() {
		return false
	}
	pattern := eg.a.Validation.Pattern
	example, err := regen.Generate(pattern)
	if err != nil {
		return eg.r.faker.Name()
	}
	return example
}

func (eg *exampleGenerator) hasMinMaxValidation() bool {
	if eg.a.Validation == nil {
		return false
	}
	return eg.a.Validation.Minimum != nil || eg.a.Validation.Maximum != nil
}

func (eg *exampleGenerator) checkMinMaxValueValidation(example interface{}) bool {
	if !eg.hasMinMaxValidation() {
		return true
	}
	valid := true
	if min := eg.a.Validation.Minimum; min != nil {
		if v, ok := example.(int); ok && float64(v) < *min {
			valid = false
		} else if v, ok := example.(float64); ok && v < *min {
			valid = false
		}
	}
	if !valid {
		return false
	}
	if max := eg.a.Validation.Maximum; max != nil {
		if v, ok := example.(int); ok && float64(v) > *max {
			return false
		} else if v, ok := example.(float64); ok && v > *max {
			return false
		}
	}
	return true
}

func (eg *exampleGenerator) generateValidatedMinMaxValueExample() interface{} {
	if !eg.hasMinMaxValidation() {
		return nil
	}
	min, max := math.Inf(1), math.Inf(-1)
	if eg.a.Validation.Minimum != nil {
		min = *eg.a.Validation.Minimum
	}
	if eg.a.Validation.Maximum != nil {
		max = *eg.a.Validation.Maximum
	}
	if math.IsInf(min, 1) {
		if eg.a.Type.Kind() == IntegerKind {
			if max == 0 {
				return int(max) - eg.r.Int()%3
			}
			return eg.r.Int() % int(max)
		}
		return eg.r.Float64() * max
	} else if math.IsInf(max, -1) {
		if eg.a.Type.Kind() == IntegerKind {
			if min == 0 {
				return int(min) + eg.r.Int()%3
			}
			return int(min) + eg.r.Int()%int(min)
		}
		return min + eg.r.Float64()*min
	} else if min < max {
		if eg.a.Type.Kind() == IntegerKind {
			return int(min) + eg.r.Int()%int(max-min)
		}
		return min + eg.r.Float64()*(max-min)
	} else if min == max {
		if eg.a.Type.Kind() == IntegerKind {
			return int(min)
		}
		return min
	}
	panic("Validation: Min > Max")
}
