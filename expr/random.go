package expr

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"

	"github.com/manveru/faker"
)

// Randomizer generates consistent random values of different types given a seed.
//
// The random values should be consistent in that given the same seed the same
// random values get generated.
//
// Setting the randomizer to nil disables example generation.
type Randomizer interface {
	// ArrayLength decides how long an example array will be
	ArrayLength() int
	// Int generates an integer example
	Int() int
	// Int32 generates an int32 example
	Int32() int32
	// Int64 generates an int64 example
	Int64() int64
	// String generates a string example
	String() string
	// Bool generates a bool example
	Bool() bool
	// Float32 generates a float32 example
	Float32() float32
	// Float64 generates a float64 example
	Float64() float64
	// UInt generates a uint example
	UInt() uint
	// UInt32 generates a uint example
	UInt32() uint32
	// UInt64 generates a uint example
	UInt64() uint64
	// Name generates a human name example
	Name() string
	// Email generates an example email address
	Email() string
	// Hostname generates an example hostname
	Hostname() string
	// IPv4Address generates an example IPv4 address
	IPv4Address() net.IP
	// IPv6Address generates an example IPv6 address
	IPv6Address() net.IP
	// URL generates an example URL
	URL() string
	// Characters generates a n-character string example
	Characters(n int) string
	// UUID generates a random v4 UUID
	UUID() string
}

// NewRandom returns a random value generator seeded from the given string
// value, using the faker library to generate random but realistic values.
func NewRandom(seed string) *ExampleGenerator {
	return &ExampleGenerator{
		Randomizer: NewFakerRandomizer(seed),
		seed:       seed,
	}
}

// ExampleGenerator generates examples from a value stream seeded by a design
// identity. Example computations derive child generators at stable design
// boundaries (user type IDs, object field names, array indices) via Derived
// so that an example is a pure function of the design: it does not change
// when unrelated parts of the design change or when code generators evaluate
// attributes in a different order.
type ExampleGenerator struct {
	Randomizer
	// seed identifies the design element this generator draws values for;
	// generators derived from it extend the seed via Derived. It is empty
	// for generators built around a caller-supplied Randomizer, which
	// cannot re-seed and therefore never derive.
	seed string
	// root points to the generator this one was derived from so that all
	// derived generators share the root's seen cache. It is nil on roots.
	root *ExampleGenerator
	seen map[string]*any
	mu   sync.RWMutex
}

// Derived returns a generator whose value stream is seeded from this
// generator's seed extended with the given identity, independent of how many
// values were drawn so far. Derived generators share the root generator's
// seen values so a user type keeps a single example wherever it appears.
// Generators that cannot re-seed (disabled example generation or a
// caller-supplied Randomizer) return themselves.
func (r *ExampleGenerator) Derived(id string) *ExampleGenerator {
	return r.reseeded(r.seed + "/" + id)
}

// Rebased returns a generator whose value stream is seeded from the root
// design seed and the given absolute identity, discarding the current
// derivation path. It anchors examples of design elements that own a global
// identity — user type IDs in particular — so the computed value is the same
// no matter where in the design the element is reached from. Generators that
// cannot re-seed (disabled example generation or a caller-supplied
// Randomizer) return themselves.
func (r *ExampleGenerator) Rebased(id string) *ExampleGenerator {
	return r.reseeded(r.store().seed + ":" + id)
}

// PreviouslySeen returns the previously seen value for a given ID
func (r *ExampleGenerator) PreviouslySeen(typeID string) (*any, bool) {
	s := r.store()
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.seen == nil {
		return nil, false
	}
	val, haveSeen := s.seen[typeID]
	return val, haveSeen
}

// HaveSeen stores the seen value in the randomizer, for reuse later
func (r *ExampleGenerator) HaveSeen(typeID string, val *any) {
	s := r.store()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.seen == nil {
		s.seen = make(map[string]*any)
	}

	s.seen[typeID] = val
}

// Field returns a generator anchored to the identity of the named field of
// the given parent attribute: the parent type identity extended with the
// field name when the parent is a user type, the field name alone otherwise.
// Code generators use it when they compute the example of one element
// extracted from a payload or result (transport params, headers, cookies,
// metadata) so the standalone example matches the corresponding field value
// in the parent type's composite example and stays stable across generator
// changes.
func (r *ExampleGenerator) Field(parent *AttributeExpr, name string) *ExampleGenerator {
	if ut, ok := parent.Type.(UserType); ok {
		return r.Rebased(ut.ID()).Derived(name)
	}
	return r.Rebased(name)
}

// store returns the generator owning the seen cache and the root design
// seed: the generator this one was derived from, or the generator itself
// when it is a root.
func (r *ExampleGenerator) store() *ExampleGenerator {
	if r.root != nil {
		return r.root
	}
	return r
}

// reseeded returns a generator drawing from a fresh value stream seeded with
// the given seed and sharing this generator's root state.
func (r *ExampleGenerator) reseeded(seed string) *ExampleGenerator {
	if r.Randomizer == nil || r.store().seed == "" {
		return r
	}
	return &ExampleGenerator{
		Randomizer: NewFakerRandomizer(seed),
		seed:       seed,
		root:       r.store(),
	}
}

// NewFakerRandomizer creates a randomizer that uses the faker library to
// generate fake but reasonable values.
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

// FakerRandomizer implements the Random interface, using the Faker library.
type FakerRandomizer struct {
	Seed  string
	faker *faker.Faker
	rand  *rand.Rand
}

func (r *FakerRandomizer) ArrayLength() int {
	return r.Int()%3 + 2
}
func (r *FakerRandomizer) Int() int {
	return r.rand.Int()
}
func (r *FakerRandomizer) Int32() int32 {
	return r.rand.Int31()
}
func (r *FakerRandomizer) Int64() int64 {
	return r.rand.Int63()
}
func (r *FakerRandomizer) String() string {
	return r.faker.Sentence(2, false)
}
func (r *FakerRandomizer) Bool() bool {
	return r.rand.Int()%2 == 0
}
func (r *FakerRandomizer) Float32() float32 {
	return r.rand.Float32()
}
func (r *FakerRandomizer) Float64() float64 {
	return r.rand.Float64()
}
func (r *FakerRandomizer) UInt() uint {
	return uint(r.UInt64())
}
func (r *FakerRandomizer) UInt32() uint32 {
	return r.rand.Uint32()
}
func (r *FakerRandomizer) UInt64() uint64 {
	return r.rand.Uint64()
}
func (r *FakerRandomizer) Email() string {
	return r.faker.Email()
}
func (r *FakerRandomizer) Hostname() string {
	return r.faker.DomainName() + "." + r.faker.DomainSuffix()
}
func (r *FakerRandomizer) IPv4Address() net.IP {
	return r.faker.IPv4Address()
}
func (r *FakerRandomizer) IPv6Address() net.IP {
	return r.faker.IPv6Address()
}
func (r *FakerRandomizer) URL() string {
	return r.faker.URL()
}
func (r *FakerRandomizer) Characters(n int) string {
	return r.faker.Characters(n)
}
func (r *FakerRandomizer) UUID() string {
	uuid := make([]byte, 16)
	r.rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
func (r *FakerRandomizer) Name() string {
	return r.faker.Name()
}

// NewDeterministicRandomizer builds a Randomizer that will return hard-coded
// values, removing all randomness from example generation.
func NewDeterministicRandomizer() Randomizer {
	return &DeterministicRandomizer{}
}

// DeterministicRandomizer returns hard-coded values, removing all randomness
// from example generation
type DeterministicRandomizer struct{}

func (DeterministicRandomizer) ArrayLength() int        { return 1 }
func (DeterministicRandomizer) Int() int                { return 1 }
func (DeterministicRandomizer) Int32() int32            { return 1 }
func (DeterministicRandomizer) Int64() int64            { return 1 }
func (DeterministicRandomizer) String() string          { return "abc123" }
func (DeterministicRandomizer) Bool() bool              { return false }
func (DeterministicRandomizer) Float32() float32        { return 1 }
func (DeterministicRandomizer) Float64() float64        { return 1 }
func (DeterministicRandomizer) UInt() uint              { return 1 }
func (DeterministicRandomizer) UInt32() uint32          { return 1 }
func (DeterministicRandomizer) UInt64() uint64          { return 1 }
func (DeterministicRandomizer) Name() string            { return "Alice" }
func (DeterministicRandomizer) Email() string           { return "alice@example.com" }
func (DeterministicRandomizer) Hostname() string        { return "example.com" }
func (DeterministicRandomizer) IPv4Address() net.IP     { return net.IPv4zero }
func (DeterministicRandomizer) IPv6Address() net.IP     { return net.IPv6zero }
func (DeterministicRandomizer) URL() string             { return "https://example.com/foo" }
func (DeterministicRandomizer) Characters(n int) string { return strings.Repeat("a", n) }
func (DeterministicRandomizer) UUID() string            { return "550e8400-e29b-41d4-a716-446655440000" }
