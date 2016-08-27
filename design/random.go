package design

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"time"

	"github.com/manveru/faker"
	regen "github.com/zach-klippenstein/goregen"
)

// RandomGenerator generates consistent random values of different types given a seed.
// The random values are consistent in that given the same seed the same random values get
// generated.
type RandomGenerator struct {
	Seed  string
	faker *faker.Faker
	rand  *rand.Rand
}

// NewRandomGenerator returns a random value generator seeded from the given string value.
func NewRandomGenerator(seed string) *RandomGenerator {
	hasher := md5.New()
	hasher.Write([]byte(seed))
	sint := int64(binary.BigEndian.Uint64(hasher.Sum(nil)))
	source := rand.NewSource(sint)
	ran := rand.New(source)
	faker := &faker.Faker{
		Language: "end",
		Dict:     faker.Dict["en"],
		Rand:     ran,
	}
	return &RandomGenerator{
		Seed:  seed,
		faker: faker,
		rand:  ran,
	}
}

// GenerateExample returns a random Go value for which t.IsCompatible returns true. The returned
// value validates any validation defined on t.
func (r *RandomGenerator) GenerateExample(t DataType) interface{} {
	return r.generateExampleRec(t, make(map[string]interface{}))
}

// generateExampleRec is the implementation for GenerateExample. It track of the types that have
// been traversed via recursion to prevent infinite loops.
func (r *RandomGenerator) generateExampleRec(t DataType, seen []string) interface{} {
	switch t {
	case Boolean:
		return r.Bool()
	case Int32:
		return r.Int32()
	case Int64:
		return r.Int64()
	case Float32:
		return r.Float32()
	case Float64:
		return r.Float64()
	case String:
		return r.String()
	case Any:
		// pick one of the primitive types for simplicity
		return r.generateExampleRec(anyPrimitive[r.Int()%len(anyPrimitive)], nil)
	default:
		switch dt := t.(type) {
		case *Array:
			ln := r.Int()%maxExampleLength + 1
			res := make([]interface{}, ln)
			for i := 0; i < ln; i++ {
				res[i] = r.generateExampleRec(dt.ElemType.Type, seen)
			}
			return dt.MakeSlice(res)
		case *Map:
			ln := r.Int()%maxExampleLength + 1
			pair := map[interface{}]interface{}{}
			for i := 0; i < ln; i++ {
				k := r.GenerateExample(dt.KeyType.Type, seen)
				v := r.GenerateExample(dt.ElemType.Type, seen)
				pair[k] = v
			}
			return dt.MakeMap(pair)
		case Object:
			// ensure fixed ordering
			keys := make([]string, 0, len(dt))
			for n := range dt {
				keys = append(keys, n)
			}
			sort.Strings(keys)

			res := make(map[string]interface{})
			for _, n := range keys {
				att := dt[n]
				res[n] = r.generateExampleRec(att.Type, seen)
			}
			return res
		case *UserTypeDefinition:
			r.generateExampleRec(dt.Type, seen)
		case *MediaTypeDefinition:
			r.generateExampleRec(dt.Type, seen)
		}
	}
}

// generateExampleAttribute returns the attribute example. If the attribute example is nil then it
// generates a value for which att.Type.IsCompatible returns true. The value also validates any
// validation defined on the attribute. The generated value is stored in the attribute example so
// that calling this method multiple times on the same attribute definition yields the same value.
func (r *RandomGenerator) generateExampleAttribute(att *AttributeExpr, seen []string) interface{} {
	if att.Example != nil {
		return att.Example
	}

	// Avoid infinite loops
	var key string
	if mt, ok := att.Type.(*MediaTypeDefinition); ok {
		key = mt.Identifier
	} else if ut, ok := att.Type.(*UserTypeDefinition); ok {
		key = ut.TypeName
	}
	if key != "" {
		count := 0
		for _, k := range seen {
			if k == key {
				count++
			}
		}
		if count > 1 {
			// Only go a couple of levels deep
			return nil
		}
		seen = append(seen, key)
	}

	switch {
	case att.Type.IsArray():
		att.Example = r.arrayExample(att.seen)

	case att.Type.IsMap():
		att.Example = r.mapExample(att.seen)

	case att.Type.IsObject():
		att.Example = r.objectExample(att.seen)

	default:
		att.Example = r.validatedExample(att.seen)
	}

	return att.Example
}

func (r *RandomGenerator) arrayExample(att *AttributeDefinition, seen []string) interface{} {
	ary := att.Type.ToArray()
	ln := r.validLength(att)
	var res []interface{}
	for i := 0; i < ln; i++ {
		ex := r.generateExampleAttribute(ary.ElemType, seen)
		if ex != nil {
			res = append(res, ex)
		}
	}
	if len(res) == 0 {
		return nil
	}
	return ary.MakeSlice(res)
}

func (r *RandomGenerator) mapExample(att *AttributeExpr, seen []string) interface{} {
	m := a.Type.ToMap()
	ln := r.validLength(att)
	res := make(map[interface{}]interface{})
	for i := 0; i < ln; i++ {
		k := r.generateExampleAttribute(m.KeyType, seen)
		v := r.generateExampleAttribute(m.ElemType, seen)
		if k != nil && v != nil {
			res[k] = v
		}
	}
	if len(res) == 0 {
		return nil
	}
	return m.MakeMap(res)
}

func (r *RandomGenerator) objectExample(att *AttributeExpr, seen []string) interface{} {
	// project media types
	actual := att
	if mt, ok := att.Type.(*MediaTypeDefinition); ok {
		v := att.View
		if v == "" {
			v = DefaultView
		}
		projected, _, err := mt.Project(v)
		if err != nil {
			panic(err) // bug
		}
		actual = projected.AttributeExpr
	}

	// ensure fixed ordering so random values are computed with consistent seeds
	o := actual.Type.ToObject()
	keys := make([]string, len(o))
	i := 0
	for n := range o {
		keys[i] = n
		i++
	}
	sort.Strings(keys)

	res := make(map[string]interface{})
	for _, n := range keys {
		if ex := r.generateExampleAttribute(o[n], seen); ex != nil {
			res[n] = ex
		}
	}
	if len(res) > 0 {
		att.Example = res
	}

	return att.Example
}

const maxAttempts = 500      // Maximum number of retries when generating validated example.
const maxExampleLength = 3   // Maximum length for array and map examples.
const maxExampleValue = 1000 // Maximum value for integer and float examples.

// validatedExample generates a random value based on the given validations.
func (r *RandomGenerator) validatedExample(att *AttributeExpr, seen []string) interface{} {
	// Randomize array length first, since that's from higher level
	if r.hasLengthValidation(att) {
		return r.generateValidatedLengthExample(att, seen)
	}
	// Enum should dominate, because the potential "examples" are fixed
	if r.hasEnumValidation(att) {
		return r.generateValidatedEnumExample(att)
	}
	// loop until a satisified example is generated
	hasFormat, hasPattern, hasMinMax := r.hasFormatValidation(att), r.hasPatternValidation(att), r.hasMinMaxValidation(att)
	attempts := 0
	for attempts < maxAttempts {
		attempts++
		var example interface{}
		// Format comes first, since it initiates the example
		if hasFormat {
			example = r.generateFormatExample(att)
		}
		// now validate with the rest of matchers; if not satisified, redo
		if hasPattern {
			if example == nil {
				example = r.generateValidatedPatternExample(att)
			} else if !r.checkPatternValidation(att, example) {
				continue
			}
		}
		if hasMinMax {
			if example == nil {
				example = r.generateValidatedMinMaxValueExample(att)
			} else if !r.checkMinMaxValueValidation(att, example) {
				continue
			}
		}
		if example == nil {
			example = r.generateExampleRec(att.Type, seen)
		}
		return example
	}
	return r.generateExampleRec(att.Type, seen)
}

func (r *RandomGenerator) validLength(att *AttributeExpr) int {
	if r.hasLengthValidation(att) {
		minlength, maxlength := math.Inf(1), math.Inf(-1)
		if att.Validation.MinLength != nil {
			minlength = float64(*att.Validation.MinLength)
		}
		if att.Validation.MaxLength != nil {
			maxlength = float64(*att.Validation.MaxLength)
		}
		count := 0
		if math.IsInf(minlength, 1) {
			count = int(maxlength) - (r.Int() % 3)
		} else if math.IsInf(maxlength, -1) {
			count = int(minlength) + (r.Int() % 3)
		} else if minlength < maxlength {
			diff := int(maxlength - minlength)
			if diff > maxExampleLength {
				diff = maxExampleLength
			}
			count = int(minlength) + (r.Int() % diff)
		} else if minlength == maxlength {
			count = int(minlength)
		} else {
			panic("Validation: MinLength > MaxLength")
		}
		if count > maxExampleLength {
			count = maxExampleLength
		}
		return count
	}
	return r.Int()%maxExampleLength + 1
}

func (r *RandomGenerator) hasLengthValidation(att *AttributeExpr) bool {
	if att.Validation == nil {
		return false
	}
	return att.Validation.MinLength != nil || att.Validation.MaxLength != nil
}

func (r *RandomGenerator) hasEnumValidation(att *AttributeExpr) bool {
	return att.Validation != nil && len(att.Validation.Values) > 0
}

func (r *RandomGenerator) hasFormatValidation(att *AttributeExpr) bool {
	return att.Validation != nil && att.Validation.Format != ""
}

func (r *RandomGenerator) hasPatternValidation(att *AttributeExpr) bool {
	return att.Validation != nil && att.Validation.Pattern != ""
}

// generateValidatedLengthExample generates a random size array of examples based on what's given.
func (r *RandomGenerator) generateValidatedLengthExample(att *AttributeExpr, seen []string) interface{} {
	ln := r.validLength(att)
	if !att.Type.IsArray() {
		return r.faker.Characters(ln)
	}
	res := make([]interface{}, ln)
	for i := 0; i < ln; i++ {
		res[i] = r.generateExampleAttribute(att.Type.ToArray().ElemType, seen)
	}
	return res
}

// generateValidatedEnumExample returns a random selected enum value.
func (r *RandomGenerator) generateValidatedEnumExample(att *AttributeExpr) interface{} {
	values := att.Validation.Values
	ln := len(values)
	i := r.Int() % ln
	return values[i]
}

// generateFormatExample returns a random example based on the format the user asks.
func (r *RandomGenerator) generateFormatExample(att *AttributeExpr) interface{} {
	if !r.hasFormatValidation(att) {
		return nil
	}
	format := att.Validation.Format
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

func (r *RandomGenerator) checkPatternValidation(att *AttributeExpr, example interface{}) bool {
	if !r.hasPatternValidation(att) {
		return true
	}
	pattern := att.Validation.Pattern
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
func (r *RandomGenerator) generateValidatedPatternExample(att *AttributeExpr) interface{} {
	if !r.hasPatternValidation(att) {
		return false
	}
	pattern := att.Validation.Pattern
	example, err := regen.Generate(pattern)
	if err != nil {
		return r.faker.Name()
	}
	return example
}

func (r *RandomGenerator) hasMinMaxValidation(att *AttributeExpr) bool {
	if att.Validation == nil {
		return false
	}
	return att.Validation.Minimum != nil || att.Validation.Maximum != nil
}

func (r *RandomGenerator) checkMinMaxValueValidation(att *AttributeExpr, example interface{}) bool {
	if !r.hasMinMaxValidation(att) {
		return true
	}
	valid := true
	if min := att.Validation.Minimum; min != nil {
		if v, ok := example.(int); ok && float64(v) < *min {
			valid = false
		} else if v, ok := example.(float64); ok && v < *min {
			valid = false
		}
	}
	if !valid {
		return false
	}
	if max := att.Validation.Maximum; max != nil {
		if v, ok := example.(int); ok && float64(v) > *max {
			return false
		} else if v, ok := example.(float64); ok && v > *max {
			return false
		}
	}
	return true
}

func (r *RandomGenerator) generateValidatedMinMaxValueExample(att *AttributeExpr) interface{} {
	if !r.hasMinMaxValidation(att) {
		return nil
	}
	min, max := math.Inf(1), math.Inf(-1)
	if att.Validation.Minimum != nil {
		min = *att.Validation.Minimum
	}
	if att.Validation.Maximum != nil {
		max = *att.Validation.Maximum
	}
	if math.IsInf(min, 1) {
		// No minimum therefore there is a maximum
		if att.Type.Kind() == Int32Kind {
			if max <= 0 {
				return -((int32(-max) + r.Int32()) % int32(maxExampleValue))
			}
			return r.Int32() % int32(max) // Do not limit with maxExampleValue as there is an explicit max
		}
		if att.Type.Kind() == Int64Kind {
			if max <= 0 {
				return -((int64(-max) + r.Int64()) % int64(maxExampleValue))
			}
			return r.Int64() % int64(max) // Do not limit with maxExampleValue as there is an explicit max
		}
		if att.Type.Kind() == Float32Kind {
			if max <= 0 {
				return r.Float32()*float32(max) + float32(max)
			}
			return r.Float32() * float32(max)
		}
		if max <= 0 {
			return r.Float64()*float64(max) + float64(max)
		}
		return r.Float64() * float64(max)
	} else if math.IsInf(max, -1) {
		// Minimum and no maximum
		if att.Type.Kind() == Int32Kind {
			if min <= 0 {
				return int32(min) + (r.Int32() % int32(maxExampleValue))
			}
			return (int32(min) + r.Int32()) % int32(maxExampleValue)
		}
		if att.Type.Kind() == Int64Kind {
			if min <= 0 {
				return int64(min) + (r.Int64() % int64(maxExampleValue))
			}
			return (int64(min) + r.Int64()) % int64(maxExampleValue)
		}
		if att.Type.Kind() == Float32Kind {
			return r.Float32()*maxExampleValue + float32(min)
		}
		return r.Float64()*maxExampleValue + float64(min)
	} else if min < max {
		if att.Type.Kind() == Int32Kind {
			return int32(min) + r.Int32()%int32(max-min)
		}
		if att.Type.Kind() == Int64Kind {
			return int64(min) + r.Int64()%int64(max-min)
		}
		if att.Type.Kind() == Int32Kind {
			return min + r.Float32()*(float32(max-min))
		}
		return min + r.Float64()*(max-min)
	} else if min == max {
		if att.Type.Kind() == Int32Kind {
			return int32(min)
		}
		if att.Type.Kind() == Int64Kind {
			return int64(min)
		}
		if att.Type.Kind() == Float32Kind {
			return float32(min)
		}
		return min
	}
	panic("Validation: Min > Max") // bug
}

// Int produces a random integer.
func (r *RandomGenerator) Int() int {
	return r.rand.Int()
}

// Int32 produces a random 32-bit integer.
func (r *RandomGenerator) Int32() int32 {
	return r.rand.Int31()
}

// Int64 produces a random 64-bit integer.
func (r *RandomGenerator) Int64() int64 {
	return r.rand.Int63()
}

// String produces a random string.
func (r *RandomGenerator) String() string {
	return r.faker.Sentence(2, false)

}

// Bool produces a random boolean.
func (r *RandomGenerator) Bool() bool {
	return r.rand.Int()%2 == 0
}

// Float32 produces a random float32 value.
func (r *RandomGenerator) Float32() float32 {
	return r.rand.Float32()
}

// Float64 produces a random float64 value.
func (r *RandomGenerator) Float64() float64 {
	return r.rand.Float64()
}
