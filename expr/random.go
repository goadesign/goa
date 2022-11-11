package expr

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
	"net"
	"strings"

	"github.com/manveru/faker"
)

// Randomizer generates consistent random values of different types given a seed.
//
// The random values should be consistent in that given the same seed the same
// random values get generated.
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
}

// NewRandom returns a random value generator seeded from the given string
// value, using the faker library to generate random but realistic values.
func NewRandom(seed string) *ExampleGenerator {
	return &ExampleGenerator{
		Randomizer: NewFakerRandomizer(seed),
	}
}

type ExampleGenerator struct {
	Randomizer
	seen map[string]*interface{}
}

// PreviouslySeen returns the previously seen value for a given ID
func (r *ExampleGenerator) PreviouslySeen(typeID string) (*interface{}, bool) {
	if r.seen == nil {
		return nil, false
	}
	val, haveSeen := r.seen[typeID]
	return val, haveSeen
}

// HaveSeen stores the seen value in the randomizer, for reuse later
func (r *ExampleGenerator) HaveSeen(typeID string, val *interface{}) {
	if r.seen == nil {
		r.seen = make(map[string]*interface{})
	}

	r.seen[typeID] = val
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
