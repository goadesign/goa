// Package accessibility supplies external types whose names have different
// visibility from the values readable through their exported fields and methods.
package accessibility

type (
	word string

	// Word is the public counterpart of the private scalar.
	Word string

	// Words has an exported name and private element type.
	Words []word

	// Keys has an exported name and private key type.
	Keys map[word]string

	// Values has an exported name and private value type.
	Values map[string]word

	// PublicWords has public element types.
	PublicWords []Word

	// PublicKeys has public key types.
	PublicKeys map[Word]string

	// PublicValues has public value types.
	PublicValues map[string]Word

	// Record exposes values whose scalar type cannot be named by another package.
	Record struct {
		// Scalar is a readable private scalar.
		Scalar word
		// Words is a readable collection with private elements.
		Words Words
		// Keys is a readable map with private keys.
		Keys Keys
		// Values is a readable map with private values.
		Values Values
	}

	// PublicRecord has the same layout with public scalar types.
	PublicRecord struct {
		// Scalar is a public scalar.
		Scalar Word
		// Words is a collection with public elements.
		Words PublicWords
		// Keys is a map with public keys.
		Keys PublicKeys
		// Values is a map with public values.
		Values PublicValues
	}

	// Choice exposes an exported collection through the external union protocol.
	Choice struct {
		words Words
	}

	// PublicChoice exposes a collection whose elements are public too.
	PublicChoice struct {
		words PublicWords
	}

	// Envelope exercises both ordinary fields and union getters.
	Envelope struct {
		// Nested contains ordinary fields.
		Nested *Record
		// Choice contains the selected words branch.
		Choice Choice
	}

	// PublicEnvelope is the fully public counterpart of Envelope.
	PublicEnvelope struct {
		// Nested contains ordinary fields.
		Nested *PublicRecord
		// Choice contains the selected words branch.
		Choice PublicChoice
	}

	hiddenRecord struct {
		Scalar string
	}

	// HiddenEnvelope requires a private object name in a conversion helper.
	HiddenEnvelope struct {
		// Nested is readable but its declared type is private.
		Nested *hiddenRecord
	}

	privateRoot struct {
		Scalar string
	}

	// Recursive exercises sharing recursive conversion helper declarations.
	Recursive struct {
		// Foo is required text.
		Foo string
		// Bar is a required integer.
		Bar int
		// Goo is a required floating-point value.
		Goo float32
		// Goo2 is a required unsigned integer.
		Goo2 uint
		// Rec is an optional recursive value.
		Rec *Recursive
	}

	// Sibling exercises required and optional calls of one conversion helper.
	Sibling struct {
		// Label is required text.
		Label string
		// Left is the required recursive child in the design.
		Left *Sibling
		// Right is the optional recursive child in the design.
		Right *Sibling
	}
)

// NewEnvelope returns nil, empty, or populated collections for states 0, 1, or 2.
func NewEnvelope(state int) *Envelope {
	record := &Record{}
	if state > 0 {
		record.Words = Words{}
		record.Keys = Keys{}
		record.Values = Values{}
	}
	if state == 2 {
		record.Scalar = "scalar"
		record.Words = Words{"first", "second"}
		record.Keys["key"] = "value"
		record.Values["key"] = "value"
	}
	result := &Envelope{Nested: record}
	result.Choice.SetWords(record.Words)
	return result
}

// NewPublicEnvelope returns the public counterpart of NewEnvelope.
func NewPublicEnvelope(state int) *PublicEnvelope {
	record := &PublicRecord{}
	if state > 0 {
		record.Words = PublicWords{}
		record.Keys = PublicKeys{}
		record.Values = PublicValues{}
	}
	if state == 2 {
		record.Scalar = "scalar"
		record.Words = PublicWords{"first", "second"}
		record.Keys["key"] = "value"
		record.Values["key"] = "value"
	}
	result := &PublicEnvelope{Nested: record}
	result.Choice.SetWords(record.Words)
	return result
}

// PrivateRoot supplies a mapping whose root name is inaccessible.
func PrivateRoot() any {
	return privateRoot{}
}

// Kind identifies the selected words branch.
func (c Choice) Kind() string {
	return "words"
}

// AsWords returns the selected words, including nil and empty collections.
func (c Choice) AsWords() (Words, bool) {
	return c.words, true
}

// SetWords selects the words branch.
func (c *Choice) SetWords(words Words) {
	c.words = words
}

// Kind identifies the selected words branch.
func (c PublicChoice) Kind() string {
	return "words"
}

// AsWords returns the selected public words.
func (c PublicChoice) AsWords() (PublicWords, bool) {
	return c.words, true
}

// SetWords selects the public words branch.
func (c *PublicChoice) SetWords(words PublicWords) {
	c.words = words
}
