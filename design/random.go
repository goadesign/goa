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

// generateExampleField returns the field example. If the field example is nil then it generates a
// value for which f.Type.IsCompatible returns true. The value also validates any validation defined
// on the field. The generated value is stored in the field example so that calling this method
// multiple times on the same field definition yields the same value.
func (r *RandomGenerator) generateExampleField(f *FieldDefinition, seen []string) interface{} {
	if f.Example != nil {
		return f.Example
	}

	// Avoid infinite loops
	var key string
	if mt, ok := f.Type.(*MediaTypeDefinition); ok {
		key = mt.Identifier
	} else if ut, ok := f.Type.(*UserTypeDefinition); ok {
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
	case f.Type.IsArray():
		f.Example = r.arrayExample(f, seen)

	case f.Type.IsMap():
		f.Example = r.mapExample(f, seen)

	case f.Type.IsObject():
		f.Example = r.objectExample(f, seen)

	default:
		f.Example = r.validatedExample(f, seen)
	}

	return f.Example
}

func (r *RandomGenerator) arrayExample(f *FieldDefinition, seen []string) interface{} {
	ary := f.Type.ToArray()
	ln := r.validLength(f)
	var res []interface{}
	for i := 0; i < ln; i++ {
		ex := r.generateExampleField(ary.ElemType, seen)
		if ex != nil {
			res = append(res, ex)
		}
	}
	if len(res) == 0 {
		return nil
	}
	return ary.MakeSlice(res)
}

func (r *RandomGenerator) mapExample(f *FieldDefinition, seen []string) interface{} {
	m := a.Type.ToMap()
	ln := r.validLength(f)
	res := make(map[interface{}]interface{})
	for i := 0; i < ln; i++ {
		k := r.generateExampleField(m.KeyType, seen)
		v := r.generateExampleField(m.ElemType, seen)
		if k != nil && v != nil {
			res[k] = v
		}
	}
	if len(res) == 0 {
		return nil
	}
	return m.MakeMap(res)
}

func (r *RandomGenerator) objectExample(f *FieldDefinition, seen []string) interface{} {
	// project media types
	actual := f
	if mt, ok := f.Type.(*MediaTypeDefinition); ok {
		v := f.View
		if v == "" {
			v = DefaultView
		}
		projected, _, err := mt.Project(v)
		if err != nil {
			panic(err) // bug
		}
		actual = projected.FieldDefinition
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
		if ex := r.generateExampleField(o[n], seen); ex != nil {
			res[n] = ex
		}
	}
	if len(res) > 0 {
		f.Example = res
	}

	return f.Example
}

const maxAttempts = 500      // Maximum number of retries when generating validated example.
const maxExampleLength = 3   // Maximum length for array and map examples.
const maxExampleValue = 1000 // Maximum value for integer and float examples.

// validatedExample generates a random value based on the given validations.
func (r *RandomGenerator) validatedExample(f *FieldDefinition, seen []string) interface{} {
	// Randomize array length first, since that's from higher level
	if r.hasLengthValidation(f) {
		return r.generateValidatedLengthExample(f, seen)
	}
	// Enum should dominate, because the potential "examples" are fixed
	if r.hasEnumValidation(f) {
		return r.generateValidatedEnumExample(f)
	}
	// loop until a satisified example is generated
	hasFormat, hasPattern, hasMinMax := r.hasFormatValidation(f), r.hasPatternValidation(f), r.hasMinMaxValidation(f)
	attempts := 0
	for attempts < maxAttempts {
		attempts++
		var example interface{}
		// Format comes first, since it initiates the example
		if hasFormat {
			example = r.generateFormatExample(f)
		}
		// now validate with the rest of matchers; if not satisified, redo
		if hasPattern {
			if example == nil {
				example = r.generateValidatedPatternExample(f)
			} else if !r.checkPatternValidation(f, example) {
				continue
			}
		}
		if hasMinMax {
			if example == nil {
				example = r.generateValidatedMinMaxValueExample(f)
			} else if !r.checkMinMaxValueValidation(f, example) {
				continue
			}
		}
		if example == nil {
			example = r.generateExampleRec(f.Type, seen)
		}
		return example
	}
	return r.generateExampleRec(f.Type, seen)
}

func (r *RandomGenerator) validLength(f *FieldDefinition) int {
	if r.hasLengthValidation(f) {
		minlength, maxlength := math.Inf(1), math.Inf(-1)
		if f.Validation.MinLength != nil {
			minlength = float64(*f.Validation.MinLength)
		}
		if f.Validation.MaxLength != nil {
			maxlength = float64(*f.Validation.MaxLength)
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

func (r *RandomGenerator) hasLengthValidation(f *FieldDefinition) bool {
	if f.Validation == nil {
		return false
	}
	return f.Validation.MinLength != nil || f.Validation.MaxLength != nil
}

func (r *RandomGenerator) hasEnumValidation(f) bool {
	return f.Validation != nil && len(f.Validation.Values) > 0
}

func (r *RandomGenerator) hasFormatValidation(f *FieldDefinition) bool {
	return f.Validation != nil && f.Validation.Format != ""
}

func (r *RandomGenerator) hasPatternValidation(f *FieldDefinition) bool {
	return f.Validation != nil && f.Validation.Pattern != ""
}

// generateValidatedLengthExample generates a random size array of examples based on what's given.
func (r *RandomGenerator) generateValidatedLengthExample(f *FieldDefinition, seen []string) interface{} {
	ln := r.validLength(f)
	if !f.Type.IsArray() {
		return r.faker.Characters(ln)
	}
	res := make([]interface{}, ln)
	for i := 0; i < ln; i++ {
		res[i] = r.generateExampleField(f.Type.ToArray().ElemType, seen)
	}
	return res
}

// generateValidatedEnumExample returns a random selected enum value.
func (r *RandomGenerator) generateValidatedEnumExample(f *FieldDefinition) interface{} {
	values := f.Validation.Values
	ln := len(values)
	i := r.Int() % ln
	return values[i]
}

// generateFormatExample returns a random example based on the format the user asks.
func (r *RandomGenerator) generateFormatExample(f *FieldDefinition) interface{} {
	if !r.hasFormatValidation(f) {
		return nil
	}
	format := f.Validation.Format
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

func (r *RandomGenerator) checkPatternValidation(f *FieldDefinition, example interface{}) bool {
	if !r.hasPatternValidation(f) {
		return true
	}
	pattern := f.Validation.Pattern
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
func (r *RandomGenerator) generateValidatedPatternExample(f *FieldDefinition) interface{} {
	if !r.hasPatternValidation(f) {
		return false
	}
	pattern := f.Validation.Pattern
	example, err := regen.Generate(pattern)
	if err != nil {
		return r.faker.Name()
	}
	return example
}

func (r *RandomGenerator) hasMinMaxValidation(f *FieldDefinition) bool {
	if f.Validation == nil {
		return false
	}
	return f.Validation.Minimum != nil || f.Validation.Maximum != nil
}

func (r *RandomGenerator) checkMinMaxValueValidation(f *FieldDefinition, example interface{}) bool {
	if !r.hasMinMaxValidation(f) {
		return true
	}
	valid := true
	if min := f.Validation.Minimum; min != nil {
		if v, ok := example.(int); ok && float64(v) < *min {
			valid = false
		} else if v, ok := example.(float64); ok && v < *min {
			valid = false
		}
	}
	if !valid {
		return false
	}
	if max := f.Validation.Maximum; max != nil {
		if v, ok := example.(int); ok && float64(v) > *max {
			return false
		} else if v, ok := example.(float64); ok && v > *max {
			return false
		}
	}
	return true
}

func (r *RandomGenerator) generateValidatedMinMaxValueExample(f *FieldDefinition) interface{} {
	if !r.hasMinMaxValidation(f) {
		return nil
	}
	min, max := math.Inf(1), math.Inf(-1)
	if f.Validation.Minimum != nil {
		min = *f.Validation.Minimum
	}
	if f.Validation.Maximum != nil {
		max = *f.Validation.Maximum
	}
	if math.IsInf(min, 1) {
		// No minimum therefore there is a maximum
		if f.Type.Kind() == Int32Kind {
			if max <= 0 {
				return -((int32(-max) + r.Int32()) % int32(maxExampleValue))
			}
			return r.Int32() % int32(max) // Do not limit with maxExampleValue as there is an explicit max
		}
		if f.Type.Kind() == Int64Kind {
			if max <= 0 {
				return -((int64(-max) + r.Int64()) % int64(maxExampleValue))
			}
			return r.Int64() % int64(max) // Do not limit with maxExampleValue as there is an explicit max
		}
		if f.Type.Kind() == Float32Kind {
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
		if f.Type.Kind() == Int32Kind {
			if min <= 0 {
				return int32(min) + (r.Int32() % int32(maxExampleValue))
			}
			return (int32(min) + r.Int32()) % int32(maxExampleValue)
		}
		if f.Type.Kind() == Int64Kind {
			if min <= 0 {
				return int64(min) + (r.Int64() % int64(maxExampleValue))
			}
			return (int64(min) + r.Int64()) % int64(maxExampleValue)
		}
		if f.Type.Kind() == Float32Kind {
			return r.Float32()*maxExampleValue + float32(min)
		}
		return r.Float64()*maxExampleValue + float64(min)
	} else if min < max {
		if f.Type.Kind() == Int32Kind {
			return int32(min) + r.Int32()%int32(max-min)
		}
		if f.Type.Kind() == Int64Kind {
			return int64(min) + r.Int64()%int64(max-min)
		}
		if f.Type.Kind() == Int32Kind {
			return min + r.Float32()*(float32(max-min))
		}
		return min + r.Float64()*(max-min)
	} else if min == max {
		if f.Type.Kind() == Int32Kind {
			return int32(min)
		}
		if f.Type.Kind() == Int64Kind {
			return int64(min)
		}
		if f.Type.Kind() == Float32Kind {
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
