package design

import (
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/goadesign/goa/dslengine"
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

// generate generates a random value based on the given validations.
func (eg *exampleGenerator) generate() interface{} {
	// Randomize array length first, since that's from higher level
	if eg.hasLengthValidation() {
		return eg.generateValidatedLengthExample()
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
		return example
	}
	return nil
}

func (eg *exampleGenerator) hasLengthValidation() bool {
	for _, v := range eg.a.Validations {
		switch v.(type) {
		case *dslengine.MinLengthValidationDefinition:
			return true
		case *dslengine.MaxLengthValidationDefinition:
			return true
		}
	}
	return false
}

// generateValidatedLengthExample generates a random size array of examples based on what's given.
func (eg *exampleGenerator) generateValidatedLengthExample() interface{} {
	minlength, maxlength := math.Inf(1), math.Inf(-1)
	for _, v := range eg.a.Validations {
		switch actual := v.(type) {
		case *dslengine.MinLengthValidationDefinition:
			minlength = math.Min(minlength, float64(actual.MinLength))
			maxlength = math.Max(maxlength, float64(actual.MinLength))
		case *dslengine.MaxLengthValidationDefinition:
			minlength = math.Min(minlength, float64(actual.MaxLength))
			maxlength = math.Max(maxlength, float64(actual.MaxLength))
		}
	}
	count := 0
	if math.IsInf(minlength, 1) {
		count = int(maxlength) - (eg.r.Int() % 3)
	} else if math.IsInf(maxlength, -1) {
		count = int(minlength) + (eg.r.Int() % 3)
	} else if minlength < maxlength {
		count = int(minlength) + (eg.r.Int() % int(maxlength-minlength))
	} else if minlength == maxlength {
		count = int(minlength)
	} else {
		panic("Validation: MinLength > MaxLength")
	}
	if !eg.a.Type.IsArray() {
		return eg.r.faker.Characters(count)
	}
	res := make([]interface{}, count)
	for i := 0; i < count; i++ {
		res[i] = eg.a.Type.ToArray().ElemType.GenerateExample(eg.r)
	}
	return res
}

func (eg *exampleGenerator) hasEnumValidation() bool {
	for _, v := range eg.a.Validations {
		if _, ok := v.(*dslengine.EnumValidationDefinition); ok {
			return true
		}
	}
	return false
}

// generateValidatedEnumExample returns a random selected enum value.
func (eg *exampleGenerator) generateValidatedEnumExample() interface{} {
	for _, v := range eg.a.Validations {
		if actual, ok := v.(*dslengine.EnumValidationDefinition); ok {
			count := len(actual.Values)
			i := eg.r.Int() % count
			return actual.Values[i]
		}
	}
	return nil
}

func (eg *exampleGenerator) hasFormatValidation() bool {
	for _, v := range eg.a.Validations {
		if _, ok := v.(*dslengine.FormatValidationDefinition); ok {
			return true
		}
	}
	return false
}

// generateFormatExample returns a random example based on the format the user asks.
func (eg *exampleGenerator) generateFormatExample() interface{} {
	for _, v := range eg.a.Validations {
		if actual, ok := v.(*dslengine.FormatValidationDefinition); ok {
			if res, ok := map[string]interface{}{
				"email":     eg.r.faker.Email(),
				"hostname":  eg.r.faker.DomainName() + "." + eg.r.faker.DomainSuffix(),
				"date-time": time.Now().Format(time.RFC3339),
				"ipv4":      eg.r.faker.IPv4Address().String(),
				"ipv6":      eg.r.faker.IPv6Address().String(),
				"uri":       eg.r.faker.URL(),
				"mac": func() string {
					res, err := regen.Generate(`([0-9A-F]{2}-){5}[0-9A-F]{2}`)
					if err != nil {
						return "12-34-56-78-9A-BC"
					}
					return res
				}(),
				"cidr":   "192.168.100.14/24",
				"regexp": eg.r.faker.Characters(3) + ".*",
			}[actual.Format]; ok {
				return res
			}
			panic("Validation: unknown format '" + actual.Format + "'") // bug
		}
	}
	return nil
}

func (eg *exampleGenerator) hasPatternValidation() bool {
	for _, v := range eg.a.Validations {
		if _, ok := v.(*dslengine.PatternValidationDefinition); ok {
			return true
		}
	}
	return false
}

func (eg *exampleGenerator) checkPatternValidation(example interface{}) bool {
	for _, v := range eg.a.Validations {
		if actual, ok := v.(*dslengine.PatternValidationDefinition); ok {
			re, err := regexp.Compile(actual.Pattern)
			if err != nil {
				panic("Validation: invalid pattern '" + actual.Pattern + "'")
			}
			if !re.MatchString(fmt.Sprint(example)) {
				return false
			}
		}
	}
	return true
}

// generateValidatedPatternExample generates a random value that satisifies the pattern. Note: if
// multiple patterns are given, only one of them is used. currently, it doesn't support multiple.
func (eg *exampleGenerator) generateValidatedPatternExample() interface{} {
	for _, v := range eg.a.Validations {
		if actual, ok := v.(*dslengine.PatternValidationDefinition); ok {
			example, err := regen.Generate(actual.Pattern)
			if err != nil {
				return eg.r.faker.Name()
			}
			return example
		}
	}
	return nil
}

func (eg *exampleGenerator) hasMinMaxValidation() bool {
	for _, v := range eg.a.Validations {
		if _, ok := v.(*dslengine.MinimumValidationDefinition); ok {
			return true
		}
		if _, ok := v.(*dslengine.MaximumValidationDefinition); ok {
			return true
		}
	}
	return false
}

func (eg *exampleGenerator) checkMinMaxValueValidation(example interface{}) bool {
	for _, v := range eg.a.Validations {
		switch actual := v.(type) {
		case *dslengine.MinimumValidationDefinition:
			if v, ok := example.(int); ok && float64(v) < actual.Min {
				return false
			} else if v, ok := example.(float64); ok && v < actual.Min {
				return false
			}
		case *dslengine.MaximumValidationDefinition:
			if v, ok := example.(int); ok && float64(v) > actual.Max {
				return false
			} else if v, ok := example.(float64); ok && v > actual.Max {
				return false
			}
		}
	}
	return true
}

func (eg *exampleGenerator) generateValidatedMinMaxValueExample() interface{} {
	min, max := math.Inf(1), math.Inf(-1)
	for _, v := range eg.a.Validations {
		switch actual := v.(type) {
		case *dslengine.MinimumValidationDefinition:
			min = math.Min(min, float64(actual.Min))
		case *dslengine.MaximumValidationDefinition:
			max = math.Max(max, float64(actual.Max))
		}
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
	} else {
		panic("Validation: Min > Max")
	}
	return nil
}
