// This file defines immutable example-randomizer configuration and the
// mutable value streams owned by one code generation run.
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
	// Randomizer generates values used in generated examples. Implementations
	// must return the same sequence when constructed from the same configuration
	// and identity.
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

	// RandomizerFactory is immutable example configuration. NewRandomizer must
	// create a new mutable stream for every call. identity identifies a stable
	// design location so separate runs produce identical examples without
	// sharing consumed stream state.
	RandomizerFactory interface {
		// NewRandomizer creates an independent value stream for identity.
		NewRandomizer(identity ExampleIdentity) Randomizer
	}

	// exampleRandomizer hides the mutable stream field while promoting its
	// value methods to ExampleGenerator.
	exampleRandomizer interface {
		Randomizer
	}

	// ExampleGenerator generates examples from one run-owned value stream.
	// Derived generators use stable design identities and share only this run's
	// recursion cache, so unrelated analysis order does not change examples.
	// One planning thread owns each generator; concurrent runs use distinct
	// generators.
	ExampleGenerator struct {
		exampleRandomizer
		factory  RandomizerFactory
		identity ExampleIdentity
		// root points to the generator this one was derived from so that all
		// derived generators share the root's seen cache. It is nil on roots.
		root *ExampleGenerator
		seen map[UserType]*any
	}

	// fakerRandomizer implements Randomizer using the faker library.
	fakerRandomizer struct {
		faker *faker.Faker
		rand  *rand.Rand
	}

	// deterministicRandomizer returns fixed values for every requested kind.
	deterministicRandomizer struct{}

	// fakerRandomizerFactory retains only the seed configured by the API DSL.
	fakerRandomizerFactory struct {
		seed string
	}

	// deterministicRandomizerFactory carries no mutable run state.
	deterministicRandomizerFactory struct{}
)

// NewExampleGenerator creates an unanchored mutable run object with an empty
// recursion cache. Call At before requesting an example value.
func NewExampleGenerator(factory RandomizerFactory) *ExampleGenerator {
	return &ExampleGenerator{factory: factory}
}

// NewFakerRandomizerFactory returns immutable configuration that creates
// independent faker streams rooted at seed.
func NewFakerRandomizerFactory(seed string) RandomizerFactory {
	return fakerRandomizerFactory{seed: seed}
}

// NewDeterministicRandomizerFactory returns immutable configuration that
// creates independent streams of fixed values.
func NewDeterministicRandomizerFactory() RandomizerFactory {
	return deterministicRandomizerFactory{}
}

// At returns a generator whose stream is anchored to identity. Anchored
// generators share this run's recursion cache but never consumed stream state.
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

// Member returns a generator for the named object member below the current
// anchored identity.
func (r *ExampleGenerator) Member(name string) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.Member(name))
}

// ArrayElement returns a generator for the indexed array element below the
// current anchored identity.
func (r *ExampleGenerator) ArrayElement(index int) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.ArrayElement(index))
}

// MapKey returns a generator for the indexed map key below the current
// anchored identity.
func (r *ExampleGenerator) MapKey(index int) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.MapKey(index))
}

// MapValue returns a generator for the indexed map value below the current
// anchored identity.
func (r *ExampleGenerator) MapValue(index int) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.MapValue(index))
}

// UnionMember returns a generator for the named union member below the current
// anchored identity.
func (r *ExampleGenerator) UnionMember(name string) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	return r.structural(r.identity.UnionMember(name))
}

func (r *fakerRandomizer) ArrayLength() int {
	return r.Int()%3 + 2
}
func (r *fakerRandomizer) Int() int {
	return r.rand.Int()
}
func (r *fakerRandomizer) Int32() int32 {
	return r.rand.Int31()
}
func (r *fakerRandomizer) Int64() int64 {
	return r.rand.Int63()
}
func (r *fakerRandomizer) String() string {
	return r.faker.Sentence(2, false)
}
func (r *fakerRandomizer) Bool() bool {
	return r.rand.Int()%2 == 0
}
func (r *fakerRandomizer) Float32() float32 {
	return r.rand.Float32()
}
func (r *fakerRandomizer) Float64() float64 {
	return r.rand.Float64()
}
func (r *fakerRandomizer) UInt() uint {
	return uint(r.UInt64())
}
func (r *fakerRandomizer) UInt32() uint32 {
	return r.rand.Uint32()
}
func (r *fakerRandomizer) UInt64() uint64 {
	return r.rand.Uint64()
}
func (r *fakerRandomizer) Email() string {
	return r.faker.Email()
}
func (r *fakerRandomizer) Hostname() string {
	return r.faker.DomainName() + "." + r.faker.DomainSuffix()
}
func (r *fakerRandomizer) IPv4Address() net.IP {
	return r.faker.IPv4Address()
}
func (r *fakerRandomizer) IPv6Address() net.IP {
	return r.faker.IPv6Address()
}
func (r *fakerRandomizer) URL() string {
	return r.faker.URL()
}
func (r *fakerRandomizer) Characters(n int) string {
	return r.faker.Characters(n)
}
func (r *fakerRandomizer) UUID() string {
	uuid := make([]byte, 16)
	r.rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
func (r *fakerRandomizer) Name() string {
	return r.faker.Name()
}

func (deterministicRandomizer) ArrayLength() int        { return 1 }
func (deterministicRandomizer) Int() int                { return 1 }
func (deterministicRandomizer) Int32() int32            { return 1 }
func (deterministicRandomizer) Int64() int64            { return 1 }
func (deterministicRandomizer) String() string          { return "abc123" }
func (deterministicRandomizer) Bool() bool              { return false }
func (deterministicRandomizer) Float32() float32        { return 1 }
func (deterministicRandomizer) Float64() float64        { return 1 }
func (deterministicRandomizer) UInt() uint              { return 1 }
func (deterministicRandomizer) UInt32() uint32          { return 1 }
func (deterministicRandomizer) UInt64() uint64          { return 1 }
func (deterministicRandomizer) Name() string            { return "Alice" }
func (deterministicRandomizer) Email() string           { return "alice@example.com" }
func (deterministicRandomizer) Hostname() string        { return "example.com" }
func (deterministicRandomizer) IPv4Address() net.IP     { return net.IPv4zero }
func (deterministicRandomizer) IPv6Address() net.IP     { return net.IPv6zero }
func (deterministicRandomizer) URL() string             { return "https://example.com/foo" }
func (deterministicRandomizer) Characters(n int) string { return strings.Repeat("a", n) }
func (deterministicRandomizer) UUID() string            { return "550e8400-e29b-41d4-a716-446655440000" }

// NewRandomizer creates an independent faker stream for identity.
func (f fakerRandomizerFactory) NewRandomizer(identity ExampleIdentity) Randomizer {
	return newFakerRandomizer(f.seed + identity.Seed())
}

// NewRandomizer creates an independent deterministic stream. identity does
// not affect fixed deterministic values.
func (deterministicRandomizerFactory) NewRandomizer(ExampleIdentity) Randomizer {
	return newDeterministicRandomizer()
}

// newFakerRandomizer creates a mutable faker stream from exact seed material.
func newFakerRandomizer(seed string) Randomizer {
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

	return &fakerRandomizer{
		faker: faker,
		rand:  ran,
	}
}

// newDeterministicRandomizer builds a stream that returns fixed values.
func newDeterministicRandomizer() Randomizer {
	return &deterministicRandomizer{}
}

// previouslySeen returns the value already being built for typ in this run.
// Declaration origins, rather than authored string IDs, distinguish graph
// nodes while still breaking recursive cycles through copied types.
func (r *ExampleGenerator) previouslySeen(typ UserType) (*any, bool) {
	s := r.store()
	if s.seen == nil {
		return nil, false
	}
	val, haveSeen := s.seen[typ.Origin()]
	return val, haveSeen
}

// haveSeen records the value being built for typ so recursive descent can
// reuse it before construction finishes.
func (r *ExampleGenerator) haveSeen(typ UserType, val *any) {
	s := r.store()
	if s.seen == nil {
		s.seen = make(map[UserType]*any)
	}

	s.seen[typ.Origin()] = val
}

// store returns the generator owning the seen cache and factory: the generator
// this one was derived from, or the generator itself when it is a root.
func (r *ExampleGenerator) store() *ExampleGenerator {
	if r.root != nil {
		return r.root
	}
	return r
}

// structural returns a generator drawing from the structural identity and
// sharing this generator's run-local recursion cache.
func (r *ExampleGenerator) structural(identity ExampleIdentity) *ExampleGenerator {
	if r.factory == nil {
		return r
	}
	if r.exampleRandomizer == nil {
		panic("example generator must be anchored before structural descent")
	}
	return r.At(identity)
}
