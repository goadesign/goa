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
	hasFormat := eg.hasFormatValidation()
	for {
		var example interface{}
		// Format comes first, since it initiates the example
		if hasFormat {
			example = eg.generateFormatExample()
		}
		// now validate with the rest of matchers; if not satisified, redo
		failed := false
		for _, v := range eg.a.Validations {
			switch actual := v.(type) {
			case *dslengine.PatternValidationDefinition:
				if example == nil {
					example = eg.generateValidatedPatternExample(actual.Pattern)
				} else if !eg.checkPatternValidation(example, actual.Pattern) {
					failed = true
				}
			case *dslengine.MinimumValidationDefinition:
				if example == nil {
					example = eg.generateValidatedMinValueExample(actual.Min)
				} else if !eg.checkMinValueValidation(example, actual.Min) {
					failed = true
				}
			case *dslengine.MaximumValidationDefinition:
				if example == nil {
					example = eg.generateValidatedMaxValueExample(actual.Max)
				} else if !eg.checkMaxValueValidation(example, actual.Max) {
					failed = true
				}
			}
			if failed {
				break
			}
		}
		if !failed {
			return example
		}
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

// generateValidatedEnumExample returns a random selected enum value,
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

func (eg *exampleGenerator) checkPatternValidation(example interface{}, pattern string) bool {
	if re, err := regexp.Compile(pattern); err == nil {
		return re.MatchString(fmt.Sprint(example))
	}
	return false
}

func (eg *exampleGenerator) generateValidatedPatternExample(pattern string) interface{} {
	example, err := regen.Generate(pattern)
	if err != nil {
		return eg.r.faker.Name()
	}
	return example
}

func (eg *exampleGenerator) checkMinValueValidation(example interface{}, min float64) bool {
	if v, ok := example.(int); ok && float64(v) < min {
		return false
	} else if v, ok := example.(float64); ok && v < min {
		return false
	}
	return true
}

func (eg *exampleGenerator) generateValidatedMinValueExample(min float64) interface{} {
	if eg.a.Type.Kind() == IntegerKind {
		return int(min) + eg.r.Int()%int(min)
	}
	return min + eg.r.Float64()*min
}

func (eg *exampleGenerator) checkMaxValueValidation(example interface{}, max float64) bool {
	if v, ok := example.(int); ok && float64(v) > max {
		return false
	} else if v, ok := example.(float64); ok && v > max {
		return false
	}
	return true
}

func (eg *exampleGenerator) generateValidatedMaxValueExample(max float64) interface{} {
	if eg.a.Type.Kind() == IntegerKind {
		return eg.r.Int() % int(max)
	}
	return eg.r.Float64() * max
}
