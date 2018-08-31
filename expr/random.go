package expr

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"

	"github.com/manveru/faker"
)

// Random generates consistent random values of different types given a seed.
// The random values are consistent in that given the same seed the same random values get
// generated.
// The generator tracks the user types that it has processed to avoid infinite recursions, this
// means a new generator should be created when wanting to generate a new random value for a user
// type.
type Random struct {
	Seed  string
	Seen  map[string]*interface{}
	faker *faker.Faker
	rand  *rand.Rand
}

// NewRandom returns a random value generator seeded from the given string value.
func NewRandom(seed string) *Random {
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
	return &Random{
		Seed:  seed,
		faker: faker,
		rand:  ran,
	}
}

// Int produces a random integer.
func (r *Random) Int() int {
	return r.rand.Int()
}

// Int32 produces a random 32-bit integer.
func (r *Random) Int32() int32 {
	return r.rand.Int31()
}

// Int64 produces a random 64-bit integer.
func (r *Random) Int64() int64 {
	return r.rand.Int63()
}

// String produces a random string.
func (r *Random) String() string {
	return r.faker.Sentence(2, false)

}

// Bool produces a random boolean.
func (r *Random) Bool() bool {
	return r.rand.Int()%2 == 0
}

// Float32 produces a random float32 value.
func (r *Random) Float32() float32 {
	return r.rand.Float32()
}

// Float64 produces a random float64 value.
func (r *Random) Float64() float64 {
	return r.rand.Float64()
}
