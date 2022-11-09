package expr

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
	"net"

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

// NewRandom creates a randomizer that uses faker to generate fake but
// reasonable values.
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

	return &FakerRandom{
		Seed:  seed,
		faker: faker,
		rand:  ran,
	}
}

// NewRandom returns a random value generator seeded from the given string value.
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

// FakerRandom implements the Random interface, using the Faker library.
type FakerRandom struct {
	Seed  string
	faker *faker.Faker
	rand  *rand.Rand
}

func (r *FakerRandom) ArrayLength() int {
	return r.Int()%3 + 2
}

// Int produces a random integer.
func (r *FakerRandom) Int() int {
	return r.rand.Int()
}

// Int32 produces a random 32-bit integer.
func (r *FakerRandom) Int32() int32 {
	return r.rand.Int31()
}

// Int64 produces a random 64-bit integer.
func (r *FakerRandom) Int64() int64 {
	return r.rand.Int63()
}

// String produces a random string.
func (r *FakerRandom) String() string {
	return r.faker.Sentence(2, false)

}

// Bool produces a random boolean.
func (r *FakerRandom) Bool() bool {
	return r.rand.Int()%2 == 0
}

// Float32 produces a random float32 value.
func (r *FakerRandom) Float32() float32 {
	return r.rand.Float32()
}

// Float64 produces a random float64 value.
func (r *FakerRandom) Float64() float64 {
	return r.rand.Float64()
}

// UInt produces a random uint value.
func (r *FakerRandom) UInt() uint {
	return uint(r.UInt64())
}

// UInt32 produces a random uint32 value.
func (r *FakerRandom) UInt32() uint32 {
	return r.rand.Uint32()
}

// UInt64 produces a random uint64 value.
func (r *FakerRandom) UInt64() uint64 {
	return r.rand.Uint64()
}

func (r *FakerRandom) Email() string {
	return r.faker.Email()
}

func (r *FakerRandom) Hostname() string {
	return r.faker.DomainName() + "." + r.faker.DomainSuffix()
}

func (r *FakerRandom) IPv4Address() net.IP {
	return r.faker.IPv4Address()
}
func (r *FakerRandom) IPv6Address() net.IP {
	return r.faker.IPv6Address()
}
func (r *FakerRandom) URL() string {
	return r.faker.URL()
}
func (r *FakerRandom) Characters(n int) string {
	return r.faker.Characters(n)
}

func (r *FakerRandom) Name() string {
	return r.faker.Name()
}
