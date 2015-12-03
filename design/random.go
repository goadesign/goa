package design

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"

	"github.com/manveru/faker"
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

// Int produces a random integer.
func (r *RandomGenerator) Int() int {
	return r.rand.Int() % 1000
}

// String produces a random string of the given
func (r *RandomGenerator) String() string {
	return r.faker.Sentence(2, false)

}

// Bool produces a random boolean.
func (r *RandomGenerator) Bool() bool {
	return r.rand.Int()%2 == 0
}

// Float64 produces a random float64 value.
func (r *RandomGenerator) Float64() float64 {
	return r.rand.Float64()
}
