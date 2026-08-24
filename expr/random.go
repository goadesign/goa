// This file creates the repeatable example values used by one code generation
// run. Each ExampleGenerator has its own sequence of values and shares only the
// map used while building recursive types.
package expr

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"

	"github.com/manveru/faker"
)

type (
	// Randomizer produces the primitive values used in generated examples. Two
	// Randomizer values created with the same settings and equal ExampleIdentity
	// keys must produce the same sequence.
	Randomizer interface {
		// ArrayLength decides how long an example array will be.
		ArrayLength() int
		// Int generates an integer example.
		Int() int
		// Int32 generates an int32 example.
		Int32() int32
		// Int64 generates an int64 example.
		Int64() int64
		// String generates a string example.
		String() string
		// Bool generates a bool example.
		Bool() bool
		// Float32 generates a float32 example.
		Float32() float32
		// Float64 generates a float64 example.
		Float64() float64
		// UInt generates a uint example.
		UInt() uint
		// UInt32 generates a uint32 example.
		UInt32() uint32
		// UInt64 generates a uint64 example.
		UInt64() uint64
		// Name generates a human name example.
		Name() string
		// Email generates an email address example.
		Email() string
		// Hostname generates a hostname example.
		Hostname() string
		// IPv4Address generates an IPv4 address example.
		IPv4Address() net.IP
		// IPv6Address generates an IPv6 address example.
		IPv6Address() net.IP
		// URL generates a URL example.
		URL() string
		// Characters generates a string containing n characters.
		Characters(n int) string
		// UUID generates a random version 4 UUID.
		UUID() string
	}

	// RandomizerFactory stores settings used to create Randomizer values.
	// NewRandomizer must return a new Randomizer on every call. Its
	// ExampleIdentity argument selects a repeatable sequence without sharing
	// values already consumed by another call.
	RandomizerFactory interface {
		// NewRandomizer creates an independent value sequence for the supplied
		// ExampleIdentity.
		NewRandomizer(identity ExampleIdentity) Randomizer
	}

	// exampleRandomizer lets ExampleGenerator expose Randomizer methods without
	// exposing its stored Randomizer field.
	exampleRandomizer interface {
		Randomizer
	}

	// ExampleGenerator builds examples from one value sequence. Child generators
	// use separate repeatable keys for fields and collection entries, and share
	// the map of values currently being built so recursive types can stop. One
	// planning thread uses each generator; concurrent runs use separate values.
	ExampleGenerator struct {
		exampleRandomizer
		factory  RandomizerFactory
		identity ExampleIdentity
		// root points to the first generator so child generators share its map of
		// values currently being built. It is nil on the first generator.
		root *ExampleGenerator
		seen map[UserType]*any
	}

	// FakerRandomizer produces repeatable example values with the faker library.
	FakerRandomizer struct {
		// Seed is the input used to create this value sequence.
		Seed  string
		faker *faker.Faker
		rand  *rand.Rand
	}

	// DeterministicRandomizer returns the same fixed value from every method.
	DeterministicRandomizer struct{}

	// fakerRandomizerFactory stores the seed configured by the API DSL.
	fakerRandomizerFactory struct {
		seed string
	}

	// deterministicRandomizerFactory needs no settings.
	deterministicRandomizerFactory struct{}
)

// NewExampleGenerator returns a generator with no selected example sequence and
// no values currently being built. Call At with an ExampleIdentity before
// requesting a value.
func NewExampleGenerator(factory RandomizerFactory) *ExampleGenerator {
	return &ExampleGenerator{factory: factory}
}

// NewFakerRandomizerFactory returns settings that create independent faker
// value sequences from seed.
func NewFakerRandomizerFactory(seed string) RandomizerFactory {
	return fakerRandomizerFactory{seed: seed}
}

// NewDeterministicRandomizerFactory returns settings that create independent
// Randomizer values whose methods return fixed values.
func NewDeterministicRandomizerFactory() RandomizerFactory {
	return deterministicRandomizerFactory{}
}

// NewFakerRandomizer returns a repeatable faker value sequence created from
// seed.
func NewFakerRandomizer(seed string) Randomizer {
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

	return &FakerRandomizer{
		Seed:  seed,
		faker: faker,
		rand:  ran,
	}
}

// NewDeterministicRandomizer returns a value sequence whose methods return
// fixed values.
func NewDeterministicRandomizer() Randomizer {
	return &DeterministicRandomizer{}
}

// At returns a generator whose value sequence is selected by the supplied
// ExampleIdentity. The result shares the map of values currently being built
// in this run, but gets a new Randomizer with no consumed values.
func (r *ExampleGenerator) At(identity ExampleIdentity) *ExampleGenerator {
	root := r.store()
	if root.factory == nil {
		return r
	}
	if identity.seed == "" {
		panic("example identity is not initialized")
	}
	return &ExampleGenerator{
		exampleRandomizer: root.factory.NewRandomizer(identity),
		factory:           root.factory,
		identity:          identity,
		root:              root,
	}
}

// Member returns a generator whose repeatable sequence is selected by the
// current ExampleIdentity plus the named field.
func (r *ExampleGenerator) Member(name string) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.Member(name))
}

// ArrayElement returns a generator whose repeatable sequence is selected by
// the current ExampleIdentity plus the array index.
func (r *ExampleGenerator) ArrayElement(index int) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.ArrayElement(index))
}

// MapKey returns a generator whose repeatable sequence is selected by the
// current ExampleIdentity plus the map key index.
func (r *ExampleGenerator) MapKey(index int) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.MapKey(index))
}

// MapValue returns a generator whose repeatable sequence is selected by the
// current ExampleIdentity plus the map value index.
func (r *ExampleGenerator) MapValue(index int) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.MapValue(index))
}

// UnionMember returns a generator whose repeatable sequence is selected by the
// current ExampleIdentity plus the union branch name.
func (r *ExampleGenerator) UnionMember(name string) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.UnionMember(name))
}

// ArrayLength returns a small positive array length.
func (r *FakerRandomizer) ArrayLength() int {
	return r.Int()%3 + 2
}

// Int returns the next int value.
func (r *FakerRandomizer) Int() int {
	return r.rand.Int()
}

// Int32 returns the next int32 value.
func (r *FakerRandomizer) Int32() int32 {
	return r.rand.Int31()
}

// Int64 returns the next int64 value.
func (r *FakerRandomizer) Int64() int64 {
	return r.rand.Int63()
}

// String returns the next short sentence.
func (r *FakerRandomizer) String() string {
	return r.faker.Sentence(2, false)
}

// Bool returns the next boolean value.
func (r *FakerRandomizer) Bool() bool {
	return r.rand.Int()%2 == 0
}

// Float32 returns the next float32 value.
func (r *FakerRandomizer) Float32() float32 {
	return r.rand.Float32()
}

// Float64 returns the next float64 value.
func (r *FakerRandomizer) Float64() float64 {
	return r.rand.Float64()
}

// UInt returns the next uint value.
func (r *FakerRandomizer) UInt() uint {
	return uint(r.UInt64())
}

// UInt32 returns the next uint32 value.
func (r *FakerRandomizer) UInt32() uint32 {
	return r.rand.Uint32()
}

// UInt64 returns the next uint64 value.
func (r *FakerRandomizer) UInt64() uint64 {
	return r.rand.Uint64()
}

// Email returns the next email address.
func (r *FakerRandomizer) Email() string {
	return r.faker.Email()
}

// Hostname returns the next hostname.
func (r *FakerRandomizer) Hostname() string {
	return r.faker.DomainName() + "." + r.faker.DomainSuffix()
}

// IPv4Address returns the next IPv4 address.
func (r *FakerRandomizer) IPv4Address() net.IP {
	return r.faker.IPv4Address()
}

// IPv6Address returns the next IPv6 address.
func (r *FakerRandomizer) IPv6Address() net.IP {
	return r.faker.IPv6Address()
}

// URL returns the next URL.
func (r *FakerRandomizer) URL() string {
	return r.faker.URL()
}

// Characters returns the next string containing n characters.
func (r *FakerRandomizer) Characters(n int) string {
	return r.faker.Characters(n)
}

// UUID returns the next random version 4 UUID.
func (r *FakerRandomizer) UUID() string {
	uuid := make([]byte, 16)
	r.rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// Name returns the next human name.
func (r *FakerRandomizer) Name() string {
	return r.faker.Name()
}

// ArrayLength returns one.
func (DeterministicRandomizer) ArrayLength() int { return 1 }

// Int returns one.
func (DeterministicRandomizer) Int() int { return 1 }

// Int32 returns one.
func (DeterministicRandomizer) Int32() int32 { return 1 }

// Int64 returns one.
func (DeterministicRandomizer) Int64() int64 { return 1 }

// String returns a fixed string.
func (DeterministicRandomizer) String() string { return "abc123" }

// Bool returns false.
func (DeterministicRandomizer) Bool() bool { return false }

// Float32 returns one.
func (DeterministicRandomizer) Float32() float32 { return 1 }

// Float64 returns one.
func (DeterministicRandomizer) Float64() float64 { return 1 }

// UInt returns one.
func (DeterministicRandomizer) UInt() uint { return 1 }

// UInt32 returns one.
func (DeterministicRandomizer) UInt32() uint32 { return 1 }

// UInt64 returns one.
func (DeterministicRandomizer) UInt64() uint64 { return 1 }

// Name returns a fixed human name.
func (DeterministicRandomizer) Name() string { return "Alice" }

// Email returns a fixed email address.
func (DeterministicRandomizer) Email() string { return "alice@example.com" }

// Hostname returns a fixed hostname.
func (DeterministicRandomizer) Hostname() string { return "example.com" }

// IPv4Address returns the unspecified IPv4 address.
func (DeterministicRandomizer) IPv4Address() net.IP { return net.IPv4zero }

// IPv6Address returns the unspecified IPv6 address.
func (DeterministicRandomizer) IPv6Address() net.IP { return net.IPv6zero }

// URL returns a fixed URL.
func (DeterministicRandomizer) URL() string { return "https://example.com/foo" }

// Characters returns n copies of "a".
func (DeterministicRandomizer) Characters(n int) string { return strings.Repeat("a", n) }

// UUID returns a fixed version 4 UUID.
func (DeterministicRandomizer) UUID() string { return "550e8400-e29b-41d4-a716-446655440000" }

// NewRandomizer creates an independent faker value sequence selected by the
// supplied ExampleIdentity key.
func (f fakerRandomizerFactory) NewRandomizer(identity ExampleIdentity) Randomizer {
	return NewFakerRandomizer(f.seed + identity.Seed())
}

// NewRandomizer creates an independent Randomizer whose methods return fixed
// values. The supplied ExampleIdentity does not change those values.
func (deterministicRandomizerFactory) NewRandomizer(ExampleIdentity) Randomizer {
	return NewDeterministicRandomizer()
}

// previouslySeen returns the value already being built for typ in this run. It
// uses the original type declaration so copied types find the same in-progress
// value and recursive definitions stop.
func (r *ExampleGenerator) previouslySeen(typ UserType) (*any, bool) {
	s := r.store()
	if s.seen == nil {
		return nil, false
	}
	val, haveSeen := s.seen[typ.Origin()]
	return val, haveSeen
}

// haveSeen records the value currently being built for typ so a recursive use
// can return it before construction finishes.
func (r *ExampleGenerator) haveSeen(typ UserType, val *any) {
	s := r.store()
	if s.seen == nil {
		s.seen = make(map[UserType]*any)
	}

	s.seen[typ.Origin()] = val
}

// store returns the first generator, which stores the RandomizerFactory and the
// map of values currently being built. It returns r when r has no parent.
func (r *ExampleGenerator) store() *ExampleGenerator {
	if r.root != nil {
		return r.root
	}
	return r
}

// structural returns a generator for the sequence selected by the supplied
// ExampleIdentity. It shares this run's map of values currently being built.
func (r *ExampleGenerator) structural(identity ExampleIdentity) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	if r.exampleRandomizer == nil {
		panic("example generator must be anchored before structural descent")
	}
	return r.At(identity)
}
